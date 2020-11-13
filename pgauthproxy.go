package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"pgAuthProxy/cmd"
)

func main() {
	if err := cmd.RootCommand(); err != nil {
		log.WithError(err).Fatal("Application start failed")
		os.Exit(1)
	}
}
