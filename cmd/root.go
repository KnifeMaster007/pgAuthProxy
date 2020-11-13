package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"pgAuthProxy/proxy"
	"pgAuthProxy/utils"
)

const (
	defaultListen = ":5432"
)

var (
	cfgFile string
	listen  string

	rootCmd = &cobra.Command{
		Use:   "pgAuthProxy",
		Short: "PostgreSQL authentication proxy",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			proxy.Start()
		},
	}
)

func initCobra() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, utils.FlagConfigFile, "", "configuration file path")

	rootCmd.PersistentFlags().StringVar(&listen, utils.FlagListen, ":5432", "bind address")
	_ = viper.BindPFlag(utils.FlagListen, rootCmd.PersistentFlags().Lookup(utils.FlagListen))
	viper.SetDefault(utils.FlagListen, defaultListen)
}

func initConfig() {
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/")

		viper.SetConfigName("pgAuthProxy.yaml")
	}
	err := viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			if cfgFile != "" {
				log.WithField(utils.PropConfigFile, cfgFile).Fatal("Config not found: ")
				os.Exit(1)
			}
			log.Warn("No config found, running with default parameters")
		default:
			log.WithError(err).Fatal("Unable to read config")
			os.Exit(1)
		}
	} else {
		log.WithField(utils.PropConfigFile, viper.ConfigFileUsed()).Info("Using config")
	}
	viper.SetEnvPrefix("pgproxy")
	viper.AutomaticEnv()
}

func RootCommand() error {
	initCobra()
	return rootCmd.Execute()
}
