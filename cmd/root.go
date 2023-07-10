package cmd

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/overmindtech/connect"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/k8s-source/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-source",
	Short: "Kubernetes source",
	Long: `Gathers details from existing kubernetes clusters
`,
	Run: func(cmd *cobra.Command, args []string) {
		exitcode := run(cmd, args)
		os.Exit(exitcode)
	},
}

func run(cmd *cobra.Command, args []string) int {
	natsServers := viper.GetStringSlice("nats-servers")
	natsNamePrefix := viper.GetString("nats-name-prefix")
	natsJWT := viper.GetString("nats-jwt")
	natsNKeySeed := viper.GetString("nats-nkey-seed")
	kubeconfig := viper.GetString("kubeconfig")
	maxParallel := viper.GetInt("max-parallel")
	hostname, err := os.Hostname()

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Could not determine hostname")

		return 1
	}

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Could not determine hostname for use in NATS connection name")

		return 1
	}

	var natsNKeySeedLog string
	var tokenClient connect.TokenClient

	if natsNKeySeed != "" {
		natsNKeySeedLog = "[REDACTED]"
	}

	log.WithFields(log.Fields{
		"nats-servers":     natsServers,
		"nats-name-prefix": natsNamePrefix,
		"max-parallel":     maxParallel,
		"nats-jwt":         natsJWT,
		"nats-nkey-seed":   natsNKeySeedLog,
		"kubeconfig":       kubeconfig,
	}).Info("Got config")

	var clientSet *kubernetes.Clientset
	var restConfig *rest.Config

	if kubeconfig == "" {
		log.Info("Using in-cluster config")

		restConfig, err = rest.InClusterConfig()

		if err != nil {
			sentry.CaptureException(err)
			log.WithError(err).Error("Could not load in-cluster config")

			return 1
		}
	} else {
		// Load kubernetes config from a file
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

		if err != nil {
			sentry.CaptureException(err)
			log.WithError(err).Error("Could not load kubernetes config")

			return 1
		}
	}

	// Create clientset
	clientSet, err = kubernetes.NewForConfig(restConfig)

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Could not create kubernetes client")

		return 1
	}

	//
	// Discover info
	//
	// Now that we have a connection to the kubernetes cluster we need to go
	// about generating some sources.
	var k8sURL *url.URL

	k8sURL, err = url.Parse(restConfig.Host)

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Errorf("Could not parse kubernetes url: %v", restConfig.Host)

		return 1
	}

	// If there is no port then set one
	if k8sURL.Port() == "" {
		switch k8sURL.Scheme {
		case "http":
			k8sURL.Host = k8sURL.Host + ":80"
		case "https":
			k8sURL.Host = k8sURL.Host + ":443"
		}
	}

	// Validate the auth params and create a token client if we are using
	// auth
	if natsJWT != "" || natsNKeySeed != "" {
		var err error

		tokenClient, err = createTokenClient(natsJWT, natsNKeySeed)

		if err != nil {
			sentry.CaptureException(err)
			log.WithError(err).Error("Error validating authentication info")

			return 1
		}
	}

	// Calculate the SHA-1 hash of the config to use as the queue name. This
	// means that sources with the same config will be in the same queue.
	// Note that the config object implements redaction in the String()
	// method so we don't have to worry about leaking secrets
	configHash := fmt.Sprintf("%x", sha1.Sum([]byte(restConfig.String())))

	e, err := discovery.NewEngine()
	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Error initializing Engine")

		return 1
	}
	e.Name = "k8s-source"
	e.NATSOptions = &connect.NATSOptions{
		NumRetries:        -1,
		RetryDelay:        5 * time.Second,
		Servers:           natsServers,
		ConnectionName:    fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
		ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
		MaxReconnects:     999,                // We are in a container so wait forever
		ReconnectWait:     2 * time.Second,
		ReconnectJitter:   2 * time.Second,
		TokenClient:       tokenClient,
	}
	e.NATSQueueName = fmt.Sprintf("k8s-source-%v", configHash)
	e.MaxParallelExecutions = maxParallel

	// Start HTTP server for status
	healthCheckPort := viper.GetInt("health-check-port")
	healthCheckPath := "/healthz"

	http.HandleFunc(healthCheckPath, func(rw http.ResponseWriter, r *http.Request) {
		if e.IsNATSConnected() {
			fmt.Fprint(rw, "ok")
		} else {
			http.Error(rw, "NATS not connected", http.StatusInternalServerError)
		}
	})

	log.WithFields(log.Fields{
		"port": healthCheckPort,
		"path": healthCheckPath,
	}).Debug("Starting healthcheck server")

	go func() {
		defer sentry.Recover()

		err := http.ListenAndServe(fmt.Sprintf(":%v", healthCheckPort), nil)

		log.WithError(err).WithFields(log.Fields{
			"port": healthCheckPort,
			"path": healthCheckPath,
		}).Error("Could not start HTTP server for /healthz health checks")
	}()

	// Create channels for interrupts
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	restart := make(chan watch.Event, 1024)

	// Get the initial starting point
	list, err := clientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("could not list namespaces")

		return 1
	}

	// Watch namespaces from here
	wi, err := clientSet.CoreV1().Namespaces().Watch(context.Background(), metav1.ListOptions{
		ResourceVersion: list.ResourceVersion,
	})

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("could not start watching namespaces")

		return 1
	}

	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()

	go func() {
		for {
			select {
			case event, ok := <-wi.ResultChan():
				if !ok {
					log.Error("Namespace watch channel closed")
					log.Info("Re-subscribing to namespace watch")

					// Get the initial starting point
					list, err = clientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})

					if err != nil {
						sentry.CaptureException(err)
						log.WithError(err).Error("could not list namespaces")

						// Send a fatal event that will kill the main goroutine
						restart <- watch.Event{
							Type: watch.EventType("FATAL"),
						}

						return
					}

					// Watch namespaces from here
					wi, err = clientSet.CoreV1().Namespaces().Watch(context.Background(), metav1.ListOptions{
						ResourceVersion: list.ResourceVersion,
					})

					if err != nil {
						sentry.CaptureException(err)
						log.WithError(err).Error("could not start watching namespaces")

						// Send a fatal event that will kill the main goroutine
						restart <- watch.Event{
							Type: watch.EventType("FATAL"),
						}

						return
					}
				}

				// Restart the engine
				restart <- event
			case <-watchCtx.Done():
				return
			}
		}
	}()

	start := func() error {
		// Query all namespaces
		log.Info("Listing namespaces")
		list, err := clientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})

		if err != nil {
			return err
		}

		namespaces := make([]string, len(list.Items))

		for i := range list.Items {
			namespaces[i] = list.Items[i].Name
		}

		log.Infof("got %v namespaces", len(namespaces))

		// Create the sources
		sourceList := sources.LoadAllSources(clientSet, k8sURL.Host, namespaces)

		// Add sources to the engine
		e.AddSources(sourceList...)

		// Start the engine
		err = e.Start()

		return err
	}

	stop := func() error {
		// Stop the engine
		err = e.Stop()

		if err != nil {
			return err
		}

		// Clear the sources
		e.ClearSources()

		return nil
	}

	// Start the service initially
	err = start()
	defer stop()

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Could not start engine")

		return 1
	}

	for {
		select {
		case <-quit:
			log.Info("Stopping engine")

			// Stopping will be handled by deferred stop()

			return 0
		case event := <-restart:
			switch event.Type {
			case "":
				// Discard empty events. After a certain period kubernetes
				// starts sending occasional empty events, I can't work out why,
				// maybe it's to keep the connection open. Either way they don't
				// represent anything and should be discarded
				log.Debug("Discarding empty event")
			case "FATAL":
				// This is a custom event type that should signal the main
				// goroutine to exit
				log.Error("Fatal error in watch goroutine")
				return 1
			default:
				err = stop()

				if err != nil {
					sentry.CaptureException(err)
					log.WithError(err).Error("Could not stop engine")

					return 1
				}

				err = start()

				if err != nil {
					sentry.CaptureException(err)
					log.WithError(err).Error("Could not start engine")

					return 1
				}
			}
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	var logLevel string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/srcman/config/k8s-source.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")

	rootCmd.PersistentFlags().StringArray("nats-servers", []string{"nats://localhost:4222", "nats://nats:4222"}, "A list of NATS servers to connect to")
	rootCmd.PersistentFlags().String("nats-name-prefix", "", "A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client")
	rootCmd.PersistentFlags().String("nats-jwt", "", "The JWT token that should be used to authenticate to NATS, provided in raw format e.g. eyJ0eXAiOiJKV1Q...")
	rootCmd.PersistentFlags().String("nats-nkey-seed", "", "The NKey seed which corresponds to the NATS JWT e.g. SUAFK6QUC...")
	rootCmd.PersistentFlags().Int("health-check-port", 8080, "The port on which to serve the /healthz endpoint")
	rootCmd.PersistentFlags().Int("max-parallel", (runtime.NumCPU() * 2), "Max number of requests to run in parallel")

	// source-specific flags
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kubeconfig file containing cluster details. If this is blank, the in-cluster config will be used")

	// Bind these to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		log.AddHook(TerminationLogHook{})

		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				viper.BindPFlag(f.Name, f)
			}
		})
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	replacer := strings.NewReplacer("-", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}
}

// createTokenClient Creates a basic token client that will authenticate to NATS
// using the given values
func createTokenClient(natsJWT string, natsNKeySeed string) (connect.TokenClient, error) {
	var kp nkeys.KeyPair
	var err error

	if natsJWT == "" {
		return nil, errors.New("nats-jwt was blank. This is required when using authentication")
	}

	if natsNKeySeed == "" {
		return nil, errors.New("nats-nkey-seed was blank. This is required when using authentication")
	}

	if _, err = jwt.DecodeUserClaims(natsJWT); err != nil {
		return nil, fmt.Errorf("could not parse nats-jwt: %v", err)
	}

	if kp, err = nkeys.FromSeed([]byte(natsNKeySeed)); err != nil {
		return nil, fmt.Errorf("could not parse nats-nkey-seed: %v", err)
	}

	return connect.NewBasicTokenClient(natsJWT, kp), nil
}

// TerminationLogHook A hook that logs fatal errors to the termination log
type TerminationLogHook struct{}

func (t TerminationLogHook) Levels() []log.Level {
	return []log.Level{log.FatalLevel}
}

func (t TerminationLogHook) Fire(e *log.Entry) error {
	tLog, err := os.OpenFile("/dev/termination-log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	var message string

	message = e.Message

	for k, v := range e.Data {
		message = fmt.Sprintf("%v %v=%v", message, k, v)
	}

	_, err = tLog.WriteString(message)

	return err
}
