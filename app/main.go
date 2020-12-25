package main

import (
	"github.com/c16a/hermes/config"
	"github.com/c16a/hermes/lib"
	"log"
)

func main() {

	serverConfig, err := config.ParseConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	ctx := lib.NewServerContext(serverConfig)

	go lib.StartWebSocketServer(serverConfig, ctx)
	lib.StartTcpServer(serverConfig, ctx)
}
