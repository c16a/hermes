package main

import (
	"github.com/c16a/hermes/config"
	"github.com/c16a/hermes/lib"
	"github.com/c16a/hermes/lib/auth"
	"log"
	"os"
)

func main() {

	configFilePath := os.Getenv("CONFIG_FILE_PATH")

	serverConfig, err := config.ParseConfig(configFilePath)
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
