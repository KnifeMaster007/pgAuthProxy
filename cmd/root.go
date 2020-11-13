package cmd

import (
	"github.com/spf13/cobra"
	"pgAuthProxy/proxy"
)

var (
	cfgFile string
	listen  string

	rootCmd = &cobra.Command{
		Use:   "pgAuthProxy",
		Short: "PostgreSQL authentication proxy",
		Run: func(cmd *cobra.Command, args []string) {
			proxy.Start()
		},
	}
)

func RootCommand() error {
	return rootCmd.Execute()
}
