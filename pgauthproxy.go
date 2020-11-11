package main

import (
	log "github.com/sirupsen/logrus"
	"net"
)

const (
	// TODO: remove hardcode
	BindAddress    = ":5432"
	TargetHost     = "pgbouncer01.d.m4"
	TargetPort     = "5432"
	TargetUser     = "igalkin"
	TargetPassword = "SnwUD5pS9Z4N"
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
				front.Run()
			}()
		}
	}

}
