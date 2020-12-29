package lib

import (
	"fmt"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	uuid "github.com/satori/go.uuid"
	"io"
)

type MqttHandler struct {
	base MqttBase
}

func (handler *MqttHandler) Handle(readWriter io.ReadWriter) {
	cPacket, err := packets.ReadPacket(readWriter)
	if err != nil {
		return
	}
	LogControlPacket(cPacket)

	var packetHandler func(io.ReadWriter, *packets.ControlPacket, MqttBase)

	switch cPacket.Type {
	case packets.CONNECT:
		packetHandler = handleConnect
		break
	case packets.PUBLISH:
		packetHandler = handlePublish
		break
	case packets.PUBREL:
		packetHandler = handlePubRel
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

	packetHandler(readWriter, cPacket, handler.base)

	return
}

func handleConnect(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	connectPacket, ok := controlPacket.Content.(*packets.Connect)
	if !ok {
		return
	}

	if len(connectPacket.ClientID) == 0 {
		connectPacket.ClientID = uuid.NewV4().String()
	}

	reasonCode, sessionPresent, maxQos := base.AddClient(readWriter, connectPacket)

	connAckPacket := packets.Connack{
		ReasonCode:     reasonCode,
		SessionPresent: sessionPresent,
		Properties: &packets.Properties{
			AssignedClientID: connectPacket.ClientID,
			MaximumQOS:       paho.Byte(maxQos),
		},
	}

	LogOutgoingPacket(packets.CONNACK)
	_, err := connAckPacket.WriteTo(readWriter)
	if err != nil {
		return
	}
}

func handleDisconnect(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	disconnectPacket, ok := controlPacket.Content.(*packets.Disconnect)
	if !ok {
		return
	}

	base.Disconnect(readWriter, disconnectPacket)
}

func handlePingRequest(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	_, ok := controlPacket.Content.(*packets.Pingreq)
	if !ok {
		return
	}

	pingResponsePacket := packets.Pingresp{}

	LogOutgoingPacket(packets.PINGRESP)
	_, err := pingResponsePacket.WriteTo(readWriter)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handlePublish(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	publishPacket, ok := controlPacket.Content.(*packets.Publish)
	if !ok {
		return
	}

	switch publishPacket.QoS {
	case 0:
		handlePubQos0(publishPacket, base)
		break
	case 1:
		handlePubQoS1(readWriter, publishPacket, base)
		break
	case 2:
		handlePubQos2(readWriter, publishPacket, base)
	}

	return
}

func handlePubQos0(publishPacket *packets.Publish, base MqttBase) {
	base.Publish(publishPacket)
}

func handlePubQoS1(readWriter io.ReadWriter, publishPacket *packets.Publish, base MqttBase) {
	pubAck := packets.Puback{
		ReasonCode: packets.PubackSuccess,
		PacketID:   publishPacket.PacketID,
	}

	LogOutgoingPacket(packets.PUBACK)
	_, err := pubAck.WriteTo(readWriter)
	if err != nil {
		return
	}
	base.Publish(publishPacket)
}

func handlePubQos2(readWriter io.ReadWriter, publishPacket *packets.Publish, base MqttBase) {
	pubReceived := packets.Pubrec{
		ReasonCode: packets.PubrecSuccess,
		PacketID:   publishPacket.PacketID,
	}

	err := base.ReservePacketID(readWriter, publishPacket)
	if err != nil {
		pubReceived.ReasonCode = packets.PubrecImplementationSpecificError
	}

	LogOutgoingPacket(packets.PUBREC)
	_, err = pubReceived.WriteTo(readWriter)
	if err != nil {
		return
	}
	base.Publish(publishPacket)
}

func handlePubRel(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	pubRelPacket, ok := controlPacket.Content.(*packets.Pubrel)
	if !ok {
		return
	}

	pubComplete := packets.Pubcomp{
		ReasonCode: packets.PubrecSuccess,
		PacketID:   pubRelPacket.PacketID,
	}

	_ = base.FreePacketID(readWriter, pubRelPacket)

	LogOutgoingPacket(packets.PUBCOMP)
	_, err := pubComplete.WriteTo(readWriter)
	if err != nil {
		return
	}
}

func handleSubscribe(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	subscribePacket, ok := controlPacket.Content.(*packets.Subscribe)
	if !ok {
		return
	}

	subAck := packets.Suback{
		PacketID: subscribePacket.PacketID,
		Reasons:  base.Subscribe(readWriter, subscribePacket),
	}

	LogOutgoingPacket(packets.SUBACK)
	_, err := subAck.WriteTo(readWriter)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handleUnsubscribe(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) {
	unsubscribePacket, ok := controlPacket.Content.(*packets.Unsubscribe)
	if !ok {
		return
	}

	unsubAck := packets.Unsuback{
		PacketID: unsubscribePacket.PacketID,
		Reasons:  base.Unsubscribe(readWriter, unsubscribePacket),
	}

	LogOutgoingPacket(packets.UNSUBACK)
	_, err := unsubAck.WriteTo(readWriter)
	if err != nil {
		fmt.Println(err)
		return
	}
}
