package lib

import (
	"fmt"
	"github.com/c16a/hermes/lib/auth"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	uuid "github.com/satori/go.uuid"
	"net"
)

type MqttHandler struct {
	authProvider auth.AuthorisationProvider
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
		break
	case packets.UNSUBSCRIBE:
		packetHandler = handleUnsubscribe
		break
	case packets.DISCONNECT:
		packetHandler = handleDisconnect
		break
	case packets.PINGREQ:
		packetHandler = handlePingRequest
	default:
		return
	}

	packetHandler(conn, cPacket, ctx)

	return
}

func handleConnect(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	connectPacket, ok := controlPacket.Content.(*packets.Connect)
	if !ok {
		return
	}

	if len(connectPacket.ClientID) == 0 {
		connectPacket.ClientID = uuid.NewV4().String()
	}

	connAckPacket := packets.Connack{
		Properties: &packets.Properties{
			AssignedClientID: connectPacket.ClientID,
			MaximumQOS:       paho.Byte(ctx.config.Server.MaxQos),
		},
	}

	var reasonCode byte
	var sessionPresent bool
	var authError error

	if ctx.authProvider != nil {
		authError = ctx.authProvider.Validate(connectPacket.Username, string(connectPacket.Password))
		if authError != nil {
			reasonCode = 135
			sessionPresent = false
		}
	}

	if authError == nil {
		reasonCode, sessionPresent = ctx.AddClient(conn, connectPacket)
	}

	connAckPacket.ReasonCode = reasonCode
	connAckPacket.SessionPresent = sessionPresent

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

func handlePingRequest(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	_, ok := controlPacket.Content.(*packets.Pingreq)
	if !ok {
		return
	}

	pingResponsePacket := packets.Pingresp{}
	_, err := pingResponsePacket.WriteTo(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
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

	subAck := packets.Suback{
		PacketID: subscribePacket.PacketID,
		Reasons:  ctx.Subscribe(conn, subscribePacket),
	}
	_, err := subAck.WriteTo(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handleUnsubscribe(conn net.Conn, controlPacket *packets.ControlPacket, ctx *ServerContext) {
	unsubscribePacket, ok := controlPacket.Content.(*packets.Unsubscribe)
	if !ok {
		return
	}

	unsubAck := packets.Unsuback{
		PacketID: unsubscribePacket.PacketID,
		Reasons:  ctx.Unsubscribe(conn, unsubscribePacket),
	}
	_, err := unsubAck.WriteTo(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
}
