package lib

import "net"

func HandleMqttConnection(conn net.Conn, ctx *ServerContext) {
	handler := &MqttHandler{
		authProvider: ctx.authProvider,
	}

	for true {
		handler.Handle(conn, ctx)
	}
}
