package main

import (
	"github.com/c16a/hermes/config"
	"github.com/c16a/hermes/lib"
	"github.com/c16a/hermes/lib/auth"
	"log"
)

func main() {

	serverConfig, err := config.ParseConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	provider, err := auth.FetchProviderFromConfig(serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	ctx := lib.NewServerContext(serverConfig, provider)

	go lib.StartWebSocketServer(serverConfig, ctx)
	lib.StartTcpServer(serverConfig, ctx)
}
