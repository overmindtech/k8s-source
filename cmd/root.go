package cmd

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/k8s-source/adapters"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdp-go/auth"
	"github.com/overmindtech/sdp-go/sdpconnect"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"

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

func run(_ *cobra.Command, _ []string) int {
	kubeconfig := viper.GetString("kubeconfig")
	maxParallel := viper.GetInt("max-parallel")
	apiKey := viper.GetString("api-key")
	app := viper.GetString("app")
	sourceName := viper.GetString("source-name")
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

	log.WithFields(log.Fields{
		"max-parallel": maxParallel,
		"kubeconfig":   kubeconfig,
		"app":          app,
		"source-name":  sourceName,
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

	restConfig.Wrap(func(rt http.RoundTripper) http.RoundTripper { return otelhttp.NewTransport(rt) })

	// Set up rate limiting
	restConfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(
		float32(viper.GetFloat64("rate-limit-qps")),
		viper.GetInt("rate-limit-burst"),
	)

	// Create clientSet
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
	// about generating some adapters.
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

	if apiKey == "" {
		log.Error("No API key provided, exiting")
		return 1
	}

	// Discover details of the Overmind instance
	log.Debug("Getting Overmind instance details")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oi, err := sdp.NewOvermindInstance(ctx, app)
	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Could not get Overmind instance details")

		return 1
	}

	var tokenClient auth.TokenClient

	// Validate the auth params and create a token client if we are using
	// auth
	tokenClient, err = auth.NewAPIKeyClient(oi.ApiUrl.String(), apiKey)

	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Could not create API key client")

		return 1
	}

	// Calculate the SHA-1 hash of the config to use as the queue name. This
	// means that adapters with the same config will be in the same queue.
	// Note that the config object implements redaction in the String()
	// method so we don't have to worry about leaking secrets
	configHash := fmt.Sprintf("%x", sha256.Sum256([]byte(restConfig.String())))

	// Work out the cluster name
	clusterName := viper.GetString("cluster-name")

	if clusterName == "" {
		clusterName = k8sURL.Host
	}

	e, err := discovery.NewEngine()
	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("Error initializing Engine")

		return 1
	}
	e.Name = sourceName
	e.NATSOptions = &auth.NATSOptions{
		NumRetries:        -1,
		RetryDelay:        5 * time.Second,
		Servers:           []string{oi.NatsUrl.String()},
		ConnectionName:    hostname,
		ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
		MaxReconnects:     999,                // We are in a container so wait forever
		ReconnectWait:     2 * time.Second,
		ReconnectJitter:   2 * time.Second,
		TokenClient:       tokenClient,
	}
	e.NATSQueueName = fmt.Sprintf("k8s-source-%v", configHash)
	e.MaxParallelExecutions = maxParallel

	// Set up heartbeat
	tokenSource := auth.NewAPIKeyTokenSource(apiKey, oi.ApiUrl.String())
	transport := oauth2.Transport{
		Source: tokenSource,
		Base:   http.DefaultTransport,
	}
	authenticatedClient := http.Client{
		Transport: otelhttp.NewTransport(&transport),
	}

	e.HeartbeatOptions = &discovery.HeartbeatOptions{
		ManagementClient: sdpconnect.NewManagementServiceClient(
			&authenticatedClient,
			oi.ApiUrl.String(),
		),
		Frequency: time.Second * 30,
		HealthCheck: func() error {
			// Make sure we can list nodes in the cluster
			_, err := clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{
				Limit: 1,
			})
			if err != nil {
				return fmt.Errorf("health check (listing nodes) failed: %w", err)
			}

			// Make sure we're connected to NATS
			if !e.IsNATSConnected() {
				return errors.New("health check (NATS) failed: not connected")
			}

			return nil
		},
	}
	e.Version = ServiceVersion
	e.Managed = sdp.SourceManaged_LOCAL
	e.Type = "k8s"

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

		server := &http.Server{
			Addr: fmt.Sprintf(":%v", healthCheckPort),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check NATS connections
				if e.IsNATSConnected() {
					// Return 200
					w.WriteHeader(http.StatusOK)
				} else {
					// Return 500 including the error
					http.Error(w, "NATS not connected", http.StatusInternalServerError)
				}
			}),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		err := server.ListenAndServe()

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
		attempts := 0
		sleep := 1 * time.Second

		for {
			select {
			case event, ok := <-wi.ResultChan():
				if !ok {
					// If the channel is closed then we need to restart the
					// watch

					log.Error("Namespace watch channel closed")
					log.Info("Re-subscribing to namespace watch")

					wi, err = watchNamespaces(watchCtx, clientSet)

					// Check for transient network errors
					if err != nil {
						var netErr *net.OpError

						if errors.As(err, &netErr) {
							// Mark a failure
							attempts++

							// If we have had less than 3 failures then retry
							if attempts < 4 {
								// The watch interface will be nil if we
								// couldn't connect, so create a fake watcher
								// that is closed so that we end up in this loop
								// again
								wi = watch.NewFake()
								wi.Stop()

								jitter := time.Duration(rand.Int63n(int64(sleep))) // nolint:gosec // we don't need cryptographically secure randomness here
								sleep = sleep + jitter/2

								log.WithError(err).Errorf("Transient network error, retrying in %v seconds", sleep.String())
								time.Sleep(sleep)
								continue
							}
						}

						sentry.CaptureException(err)
						log.WithError(err).Error("could not list namespaces")

						// Send a fatal event that will kill the main goroutine
						restart <- watch.Event{
							Type: watch.EventType("FATAL"),
						}

						return
					}

					// If it's worked, reset the failure counter
					attempts = 0
				} else {
					// If a watch event is received then we need to restart the
					// engine
					restart <- event
				}
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

		// Create the adapter list
		adapterList := adapters.LoadAllAdapters(clientSet, clusterName, namespaces)

		// Add adapters to the engine
		e.AddAdapters(adapterList...)

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

		// Clear the adapters
		e.ClearAdapters()

		return nil
	}

	// Start the service initially
	err = start()
	if err != nil {
		err = fmt.Errorf("Could not start engine: %w", err)
		sentry.CaptureException(err)
		log.WithError(err)

		return 1
	}

	defer func() {
		err := stop()
		if err != nil {
			err = fmt.Errorf("Could not stop engine: %w", err)
			sentry.CaptureException(err)
			log.WithError(err)
		}
	}()

	for {
		select {
		case <-quit:
			log.Info("Stopping engine")

			// Stopping will be handled by deferred stop()

			return 0
		case event := <-restart:
			switch event.Type { // nolint:exhaustive // we on purpose fall through to default
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
			case "MODIFIED":
				log.Debug("Namespace modified, ignoring")
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

// Watches k8s namespaces from the current state, sending new events for each change
func watchNamespaces(ctx context.Context, clientSet *kubernetes.Clientset) (watch.Interface, error) {
	// Get the initial starting point
	list, err := clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	// Watch namespaces from here
	wi, err := clientSet.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{
		ResourceVersion: list.ResourceVersion,
	})

	if err != nil {
		return nil, err
	}

	return wi, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	var logLevel string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/srcman/config/k8s-source.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")

	rootCmd.PersistentFlags().String("api-key", "", "The API key to use to authenticate to the Overmind API")
	rootCmd.PersistentFlags().String("app", "https://app.overmind.tech", "The URL of the Overmind instance to connect to")
	rootCmd.PersistentFlags().String("source-name", "k8s-source", "The name of the source")

	rootCmd.PersistentFlags().Int("health-check-port", 8080, "The port on which to serve the /healthz endpoint")
	rootCmd.PersistentFlags().Int("max-parallel", 2_000, "Max number of requests to run in parallel")

	// source-specific flags
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kubeconfig file containing cluster details. If this is blank, the in-cluster config will be used")
	rootCmd.PersistentFlags().Float32("rate-limit-qps", 10.0, "The maximum sustained queries per second from this source to the kubernetes API")
	rootCmd.PersistentFlags().Int("rate-limit-burst", 30, "The maximum burst of queries from this source to the kubernetes API")
	rootCmd.PersistentFlags().String("cluster-name", "", "The descriptive name of the cluster this source is running on. If this is blank, the hostname will be used from the Kube config")

	// tracing
	rootCmd.PersistentFlags().String("honeycomb-api-key", "", "If specified, configures opentelemetry libraries to submit traces to honeycomb")
	rootCmd.PersistentFlags().String("sentry-dsn", "", "If specified, configures sentry libraries to capture errors")
	rootCmd.PersistentFlags().String("run-mode", "release", "Set the run mode for this service, 'release', 'debug' or 'test'. Defaults to 'release'.")

	// Bind these to viper
	cobra.CheckErr(viper.BindPFlags(rootCmd.PersistentFlags()))

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		log.AddHook(TerminationLogHook{})
		log.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
			log.AllLevels[:log.GetLevel()+1]...,
		)))

		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				err := viper.BindPFlag(f.Name, f)
				if err != nil {
					log.WithError(err).Errorf("Could not bind flag %s to viper", f.Name)
				}
			}
		})

		if sentryDSN := viper.GetString("sentry-dsn"); sentryDSN != "" {
			err := initSentry(sentryDSN)

			if err != nil {
				log.Errorf("error initializing sentry: %s", err)
			}
		}

		if honeycombAPIKey := viper.GetString("honeycomb-api-key"); honeycombAPIKey != "" {
			tracingOpts := []otlptracehttp.Option{
				otlptracehttp.WithEndpoint("api.honeycomb.io"),
				otlptracehttp.WithHeaders(map[string]string{"x-honeycomb-team": honeycombAPIKey}),
			}

			if err := initOtel(tracingOpts...); err != nil {
				log.Fatal(err)
			}
		}
	}

	// shut down tracing at the end of the process
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		shutdownTracing()
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
