package mqtt

import "net"

func HandleMqttConnection(conn net.Conn, ctx *ServerContext) {
	handler := &MqttHandler{base: ctx}

	for true {
		handler.Handle(conn)
	}
}
