package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
		log.WithFields(log.Fields{
			"nats-servers":     viper.Get("nats-servers"),
			"nats-name-prefix": viper.Get("nats-name-prefix"),
			"nats-ca-file":     viper.Get("nats-ca-file"),
			"nats-jwt-file":    viper.Get("nats-jwt-file"),
			"nats-nkey-file":   viper.Get("nats-nkey-file"),
			"kubeconfig":       viper.Get("kubeconfig"),
		}).Info("Got config")
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
