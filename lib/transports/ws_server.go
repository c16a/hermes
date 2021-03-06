package transports

import (
	"fmt"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/mqtt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func StartWebSocketServer(serverConfig *config.Config, ctx *mqtt.ServerContext) {
	upgrader := websocket.Upgrader{}

	httpAddr := serverConfig.Server.HttpAddress
	http.HandleFunc("/socket", func(writer http.ResponseWriter, request *http.Request) {
		c, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		go mqtt.HandleMqttConnection(c.UnderlyingConn(), ctx)
	})

	fmt.Printf("Starting Websocket server on %s\n", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
