package transports

import (
	"crypto/tls"
	"fmt"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/mqtt"
	"net"
)

func StartTcpServer(serverConfig *config.Config, ctx *mqtt.ServerContext) {
	var listener net.Listener
	var listenerErr error

	tcpAddress := serverConfig.Server.TcpAddress

	tlsConfigFromFile := serverConfig.Server.Tls
	if tlsConfigFromFile == nil {
		listener, listenerErr = net.Listen("tcp", tcpAddress)
	} else {
		if len(tlsConfigFromFile.CertFile) == 0 || len(tlsConfigFromFile.KeyFile) == 0 {
			// TCP config invalid - don't start TCP server
			return
		}
		cert, err := tls.LoadX509KeyPair(tlsConfigFromFile.CertFile, tlsConfigFromFile.KeyFile)
		if err != nil {
			// Could not read certs - don't start TCP server
			return
		}
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		listener, listenerErr = tls.Listen("tcp", tcpAddress, &tlsConfig)
	}

	if listenerErr != nil {
		return
	}
	defer listener.Close()

	fmt.Printf("Starting TCP server on %s\n", tcpAddress)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go mqtt.HandleMqttConnection(conn, ctx)
	}
}
