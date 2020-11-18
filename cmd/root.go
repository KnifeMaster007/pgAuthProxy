package cmd

import (
	"github.com/KnifeMaster007/pgAuthProxy/proxy"
	"github.com/KnifeMaster007/pgAuthProxy/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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
			a := viper.GetString(utils.ConfigCleartextPassword)
			if a != "" {
			}
			proxy.Start()
		},
	}
)

func initCobra() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, utils.ConfigFileFlag, "", "configuration file path")

	rootCmd.PersistentFlags().StringVar(&listen, utils.ConfigListenFlag, ":5432", "bind address")
	_ = viper.BindPFlag(utils.ConfigListenFlag, rootCmd.PersistentFlags().Lookup(utils.ConfigListenFlag))
	viper.SetDefault(utils.ConfigListenFlag, defaultListen)

	rootCmd.PersistentFlags().Bool(utils.ConfigCleartextPasswordFlag, false,
		"use cleartext password instead of MD5-hashed")
	_ = viper.BindPFlag(utils.ConfigCleartextPassword,
		rootCmd.PersistentFlags().Lookup(utils.ConfigCleartextPasswordFlag))
	_ = viper.BindEnv(utils.ConfigCleartextPassword, utils.ConfigCleartextPasswordEnv)
	viper.SetDefault(utils.ConfigCleartextPasswordFlag, false)
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
	viper.SetEnvPrefix(utils.ConfigEnvPrefix)
	viper.AutomaticEnv()
}

func RootCommand() error {
	initCobra()
	return rootCmd.Execute()
}
