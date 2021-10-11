package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dylanratcliffe/discovery"
	"github.com/dylanratcliffe/k8s-source/internal/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-source",
	Short: "Remote primary source for kubernetes",
	Long: `This is designed to be run as part of srcman
(https://github.com/dylanratcliffe/srcman)

It responds to requests for items relating to kubernetes clusters.
Each namespace is a separate context, as are non-namespaced resources
within each cluster.

This can be configured using a yaml file and the --config flag, or by 
using appropriately named environment variables, for example "nats-name-prefix"
can be set using an environment variable named "NATS_NAME_PREFIX"
`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				viper.BindPFlag(f.Name, f)
			}
		})

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		natsServers := viper.GetStringSlice("nats-servers")
		natsNamePrefix := viper.GetString("nats-name-prefix")
		natsCAFile := viper.GetString("nats-ca-file")
		natsJWTFile := viper.GetString("nats-jwt-file")
		natsNKeyFile := viper.GetString("nats-nkey-file")
		kubeconfig := viper.GetString("kubeconfig")
		maxParallel := viper.GetInt("max-parallel")
		hostname, err := os.Hostname()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not determine hostname for use in NATS connection name")

			os.Exit(1)
		}

		log.WithFields(log.Fields{
			"nats-servers":     natsServers,
			"nats-name-prefix": natsNamePrefix,
			"nats-ca-file":     natsCAFile,
			"nats-jwt-file":    natsJWTFile,
			"nats-nkey-file":   natsNKeyFile,
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
		var k8sHost string
		var k8sPort string
		var nss sources.NamespaceStorage
		var sourceList []discovery.Source

		k8sURL, err = url.Parse(rc.Host)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorf("Could not parse kubernetes url: %v", rc.Host)

			os.Exit(1)
		}

		// Calculate the cluster name
		k8sHost, k8sPort, err = net.SplitHostPort(k8sURL.Host)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorf("Could not detect port from host: %v", k8sURL.Host)

			os.Exit(1)
		}

		if k8sPort == "" || k8sPort == "443" {
			// If a port isn't specific or it's a standard port then just return
			// the hostname
			sources.ClusterName = k8sHost
		} else {
			// If it is running on a custom port then return host:port
			sources.ClusterName = k8sHost + ":" + k8sPort
		}

		// Get list of namspaces
		nss = sources.NamespaceStorage{
			CS:            clientSet,
			CacheDuration: (10 * time.Second),
		}

		// Load all sources
		for _, srcFunction := range sources.SourceFunctions {
			src := srcFunction(clientSet)
			src.NSS = &nss

			sourceList = append(sourceList, &src)
		}

		e := discovery.Engine{
			Name: "kubernetes-source",
			NATSOptions: &discovery.NATSOptions{
				URLs:           natsServers,
				ConnectionName: fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
				ConnectTimeout: (10 * time.Second), // TODO: Make configurable
				NumRetries:     999,                // We are in a container so wait forever
				CAFile:         natsCAFile,
				NkeyFile:       natsNKeyFile,
				JWTFile:        natsJWTFile,
			},
			MaxParallelExecutions: maxParallel,
		}

		e.AddSources(sourceList...)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8s-source.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")

	rootCmd.PersistentFlags().StringArray("nats-servers", []string{"nats://localhost:4222", "nats://nats:4222"}, "A list of NATS servers to connect to")
	rootCmd.PersistentFlags().String("nats-name-prefix", "", "A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client")
	rootCmd.PersistentFlags().String("nats-ca-file", "", "Path to the CA file that NATS should use when connecting over TLS")
	rootCmd.PersistentFlags().String("nats-jwt-file", "", "Path to the file containing the user JWT")
	rootCmd.PersistentFlags().String("nats-nkey-file", "", "Path to the file containing the NKey seed")
	rootCmd.PersistentFlags().String("kubeconfig", "/etc/srcman/config/kubeconfig", "Path to the kubeconfig file containing cluster details")
	rootCmd.PersistentFlags().Int("max-parallel", 12, "Max number of requests to run in parallel")

	// Bind these to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".k8s-source" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".k8s-source")
	}

	replacer := strings.NewReplacer("-", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}
}
