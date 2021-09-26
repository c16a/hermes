package main

import (
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/mqtt"
	"github.com/c16a/hermes/lib/transports"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {

	configFilePath := os.Getenv("CONFIG_FILE_PATH")

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	serverConfig, err := config.ParseConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, err := mqtt.NewServerContext(serverConfig, logger)
	if err != nil {
		log.Fatal(err)
	}

	go transports.StartWebSocketServer(serverConfig, ctx, logger)
	transports.StartTcpServer(serverConfig, ctx, logger)
}
