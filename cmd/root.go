package cmd

import (
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

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/k8s-source/internal/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	Short: "Remote primary source for kubernetes",
	Long: `This is designed to be run as part of srcman
(https://github.com/overmindtech/srcman)

It responds to requests for items relating to kubernetes clusters.
Each namespace is a separate context, as are non-namespaced resources
within each cluster.

This can be configured using a yaml file and the --config flag, or by 
using appropriately named environment variables, for example "nats-name-prefix"
can be set using an environment variable named "NATS_NAME_PREFIX"
`,
	Run: func(cmd *cobra.Command, args []string) {
		natsServers := viper.GetStringSlice("nats-servers")
		natsNamePrefix := viper.GetString("nats-name-prefix")
		natsJWT := viper.GetString("nats-jwt")
		natsNKeySeed := viper.GetString("nats-nkey-seed")
		kubeconfig := viper.GetString("kubeconfig")
		maxParallel := viper.GetInt("max-parallel")
		hostname, err := os.Hostname()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not determine hostname for use in NATS connection name")

			os.Exit(1)
		}

		var natsNKeySeedLog string
		var tokenClient discovery.TokenClient

		if natsNKeySeed != "" {
			natsNKeySeedLog = "[REDACTED]"
		}

		log.WithFields(log.Fields{
			"nats-servers":     natsServers,
			"nats-name-prefix": natsNamePrefix,
			"nats-jwt":         natsJWT,
			"nats-nkey-seed":   natsNKeySeedLog,
			"kubeconfig":       kubeconfig,
		}).Info("Got config")

		var rc *rest.Config
		var clientSet *kubernetes.Clientset

		// Load kubernetes config
		rc, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not load kubernetes config")

			os.Exit(1)
		}

		// Create clientset
		clientSet, err = kubernetes.NewForConfig(rc)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not create kubernetes client")

			os.Exit(1)
		}

		//
		// Discover info
		//
		// Now that we have a connection to the kubernetes cluster we need to go
		// about generating some sources.
		var k8sURL *url.URL
		var nss sources.NamespaceStorage
		var sourceList []discovery.Source

		k8sURL, err = url.Parse(rc.Host)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorf("Could not parse kubernetes url: %v", rc.Host)

			os.Exit(1)
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

		sources.ClusterName = k8sURL.Host

		// Get list of namspaces
		nss = sources.NamespaceStorage{
			CS:            clientSet,
			CacheDuration: (10 * time.Second),
		}

		// Load all sources
		for _, srcFunction := range sources.SourceFunctions {
			src, err := srcFunction(clientSet)

			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"sourceName": src.Name(),
				}).Error("Failed loading source")

				continue
			}

			src.NSS = &nss

			sourceList = append(sourceList, &src)
		}

		// Validate the auth params and create a token client if we are using
		// auth
		if natsJWT != "" || natsNKeySeed != "" {
			var err error

			tokenClient, err = createTokenClient(natsJWT, natsNKeySeed)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Fatal("Error validating authentication info")
			}
		}

		e := discovery.Engine{
			Name: "kubernetes-source",
			NATSOptions: &discovery.NATSOptions{
				URLs:            natsServers,
				ConnectionName:  fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
				ConnectTimeout:  (10 * time.Second), // TODO: Make configurable
				MaxReconnect:    -1,
				ReconnectWait:   1 * time.Second,
				ReconnectJitter: 1 * time.Second,
				TokenClient:     tokenClient,
			},
			MaxParallelExecutions: maxParallel,
		}

		e.AddSources(sourceList...)

		// Start HTTP server for status
		healthCheckPort := 8080
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
			log.Fatal(http.ListenAndServe(":8080", nil))
		}()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start HTTP server for /healthz health checks")

			os.Exit(1)
		}

		err = e.Start()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start engine")

			os.Exit(1)
		}

		sigs := make(chan os.Signal, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs

		log.Info("Stopping engine")

		err = e.Stop()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not stop engine")

			os.Exit(1)
		}

		log.Info("Stopped")

		os.Exit(0)
	},
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
	rootCmd.PersistentFlags().String("kubeconfig", "/etc/srcman/config/kubeconfig", "Path to the kubeconfig file containing cluster details")
	rootCmd.PersistentFlags().Int("max-parallel", (runtime.NumCPU() * 2), "Max number of requests to run in parallel")

	// Bind these to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}

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
func createTokenClient(natsJWT string, natsNKeySeed string) (discovery.TokenClient, error) {
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

	return discovery.NewBasicTokenClient(natsJWT, kp), nil
}
