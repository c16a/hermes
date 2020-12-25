package lib

import "net"

func HandleMqttConnection(conn net.Conn, ctx *ServerContext) {
	handler := &MqttHandler{}

	for true {
		handler.Handle(conn, ctx)
	}
}
