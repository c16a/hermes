package lib

import (
	"fmt"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	uuid "github.com/satori/go.uuid"
	"net"
)

type MqttHandler struct {
}

func (handler *MqttHandler) Handle(conn net.Conn, ctx *ServerContext) {
	cPacket, err := packets.ReadPacket(conn)
	if err != nil {
		return
	}

	var packetHandler func(net.Conn, *packets.ControlPacket, *ServerContext)

	switch cPacket.Type {
	case packets.CONNECT:
		packetHandler = handleConnect
		break
	case packets.PUBLISH:
		packetHandler = handlePublish
		break
	case packets.SUBSCRIBE:
		packetHandler = handleSubscribe
	case packets.DISCONNECT:
		packetHandler = handleDisconnect
	default:
		return
	}

	packetHandler(conn, cPacket, ctx)

	return
}

func handleConnect(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	connectPacket, ok := controlPacket.Content.(*packets.Connect)

	if len(connectPacket.ClientID) == 0 {
		connectPacket.ClientID = uuid.NewV4().String()
	}
	if ok {
	} else {
		return
	}

	reasonCode, sessionExists := ctx.AddClient(conn, connectPacket)
	connAckPacket := packets.Connack{
		Properties: &packets.Properties{
			AssignedClientID: connectPacket.ClientID,
			MaximumQOS:       paho.Byte(2),
		},
		ReasonCode:     reasonCode,
		SessionPresent: sessionExists,
	}

	_, err := connAckPacket.WriteTo(conn)
	if err != nil {
		return
	}
}

func handleDisconnect(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	disconnectPacket, ok := controlPacket.Content.(*packets.Disconnect)
	if !ok {
		return
	}

	ctx.Disconnect(conn, disconnectPacket)
}

func handlePublish(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	publishPacket, ok := controlPacket.Content.(*packets.Publish)
	if !ok {
		return
	}

	switch publishPacket.QoS {
	case 0:
		handlePubQos0(publishPacket, ctx)
		break
	case 1:
		handlePubQoS1(conn, publishPacket, ctx)
	}

	return
}

func handlePubQos0(publishPacket *packets.Publish, ctx *ServerContext) {
	ctx.Publish(publishPacket)
}

func handlePubQoS1(conn net.Conn, publishPacket *packets.Publish, ctx *ServerContext) {
	pubAck := packets.Puback{
		ReasonCode: packets.PubackSuccess,
		PacketID:   publishPacket.PacketID,
	}
	_, err := pubAck.WriteTo(conn)
	if err != nil {
		return
	}
	ctx.Publish(publishPacket)
}

func handleSubscribe(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	subscribePacket, ok := controlPacket.Content.(*packets.Subscribe)
	if !ok {
		return
	}

	ctx.Subscribe(conn, subscribePacket)

	subAck := packets.Suback{
		PacketID: subscribePacket.PacketID,
		Reasons:  []byte{packets.SubackGrantedQoS0},
	}
	_, err := subAck.WriteTo(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
}
