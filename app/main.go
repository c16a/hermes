package main

import (
	"github.com/c16a/hermes/config"
	"github.com/c16a/hermes/lib"
	"log"
	"os"
)

func main() {

	configFilePath := os.Getenv("CONFIG_FILE_PATH")

	serverConfig, err := config.ParseConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	ctx := lib.NewServerContext(serverConfig)

	go lib.StartWebSocketServer(serverConfig, ctx)
	lib.StartTcpServer(serverConfig, ctx)
}
