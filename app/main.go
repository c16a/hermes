package main

import (
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/mqtt"
	"github.com/c16a/hermes/lib/transports"
	"log"
	"os"
)

func main() {

	configFilePath := os.Getenv("CONFIG_FILE_PATH")

	serverConfig, err := config.ParseConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, err := mqtt.NewServerContext(serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	go transports.StartWebSocketServer(serverConfig, ctx)
	transports.StartTcpServer(serverConfig, ctx)
}
