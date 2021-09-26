package mqtt

import "net"

func HandleMqttConnection(conn net.Conn, ctx *ServerContext) {
	handler := &MqttHandler{base: ctx, logger: ctx.logger}

	for true {
		handler.Handle(conn)
	}
}
