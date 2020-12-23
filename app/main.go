package main

import (
	"fmt"
	"github.com/c16a/hermes/config"
	"github.com/c16a/hermes/lib"
	"log"
	"net"
	"net/http"
)

func main() {

	serverConfig, err := config.ParseConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	ctx := lib.NewServerContext()

	go startWebSocketServer(serverConfig, ctx)
	startTcpServer(serverConfig, ctx)
}

func startTcpServer(serverConfig *config.Config, ctx *lib.ServerContext) {
	tcpAddress := serverConfig.Server.TcpAddress
	l, err := net.Listen("tcp", tcpAddress)
	if err != nil {
		return
	}
	defer l.Close()

	fmt.Printf("Starting TCP server on %s\n", tcpAddress)
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go handleTcpConnection(conn, ctx)
	}
}

func handleTcpConnection(conn net.Conn, ctx *lib.ServerContext) {
	fmt.Printf("Serving %s\n", conn.RemoteAddr().String())

	handler := &lib.SimpleTcpHandler{}

	for true {
		response, quit, err := handler.Handle(conn, ctx)
		if err != nil {
			conn.Write(response)
			continue
		}
		if quit {
			conn.Close()
			break
		}
		conn.Write(response)
	}
}

func startWebSocketServer(serverConfig *config.Config, ctx *lib.ServerContext) {
	httpAddr := serverConfig.Server.HttpAddress
	http.HandleFunc("/publish", lib.PublishHttp(ctx))
	http.HandleFunc("/subscribe", lib.SubscribeHttp(ctx))

	fmt.Printf("Starting Websocket server on %s\n", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
