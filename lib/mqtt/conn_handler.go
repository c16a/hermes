package mqtt

import "net"

func HandleMqttConnection(conn net.Conn, ctx *ServerContext) {
	handler := &MqttHandler{base: ctx, logger: ctx.logger}

	for {
		handler.Handle(conn)
	}
}
