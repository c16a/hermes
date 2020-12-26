package main

import (
	"github.com/c16a/hermes/lib"
	"github.com/c16a/hermes/lib/config"
	"log"
	"os"
)

func main() {

	configFilePath := os.Getenv("CONFIG_FILE_PATH")

	serverConfig, err := config.ParseConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, err := lib.NewServerContext(serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	go lib.StartWebSocketServer(serverConfig, ctx)
	lib.StartTcpServer(serverConfig, ctx)
}
