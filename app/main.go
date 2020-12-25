package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/c16a/hermes/config"
	"github.com/c16a/hermes/lib"
)

func main() {

	serverConfig, err := config.ParseConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	ctx := lib.NewServerContext(serverConfig)

	go startWebSocketServer(serverConfig, ctx)
	startTcpServer(serverConfig, ctx)
}

func startTcpServer(serverConfig *config.Config, ctx *lib.ServerContext) {
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
		go handleMqttConnection(conn, ctx)
	}
}

func handleMqttConnection(conn net.Conn, ctx *lib.ServerContext) {
	handler := &lib.MqttHandler{}

	for true {
		handler.Handle(conn, ctx)
	}
}

func startWebSocketServer(serverConfig *config.Config, ctx *lib.ServerContext) {
	httpAddr := serverConfig.Server.HttpAddress
	http.HandleFunc("/publish", lib.PublishHttp(ctx))
	http.HandleFunc("/subscribe", lib.SubscribeHttp(ctx))

	fmt.Printf("Starting Websocket server on %s\n", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
