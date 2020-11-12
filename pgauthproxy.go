package main

import (
	log "github.com/sirupsen/logrus"
	"net"
)

const (
	// TODO: remove hardcode
	BindAddress = ":15432"
)

func main() {
	//log.SetReportCaller(true)
	log.Info("Starting auth pgAuthProxy...")
	log.SetLevel(log.DebugLevel)

	const BindAddr = BindAddress
	server, _ := net.Listen("tcp", BindAddr)
	log.WithField("address", BindAddr).Info("Started listening")
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Debug("Connection initialization error: " + err.Error())
		} else {
			go func() {
				defer conn.Close()
				front := NewProxyFront(conn, authStub)
				defer front.Close()
				front.Run()
			}()
		}
	}

}
