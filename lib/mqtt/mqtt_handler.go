package mqtt

import (
	"errors"
	"io"

	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type MqttHandler struct {
	base   MqttBase
	logger *zap.Logger
}

func (handler *MqttHandler) Handle(readWriter io.ReadWriter) {
	cPacket, err := packets.ReadPacket(readWriter)
	if err != nil {
		return
	}

	handler.logger.With(
		zap.Uint16("packetID", cPacket.PacketID()),
		zap.String("type", cPacket.PacketType()),
	).Info("Received packet")

	var packetHandler func(io.ReadWriter, *packets.ControlPacket, MqttBase) error

	switch cPacket.Type {
	case packets.CONNECT:
		packetHandler = handleConnect
	case packets.PUBLISH:
		packetHandler = handlePublish
	case packets.PUBREL:
		packetHandler = handlePubRel
	case packets.SUBSCRIBE:
		packetHandler = handleSubscribe
	case packets.UNSUBSCRIBE:
		packetHandler = handleUnsubscribe
	case packets.DISCONNECT:
		packetHandler = handleDisconnect
	case packets.PINGREQ:
		packetHandler = handlePingRequest
	default:
		return
	}

	err = packetHandler(readWriter, cPacket, handler.base)
	if err != nil {
		handler.logger.Error("error handling packet", zap.Error(err))
	}

	handler.logger.With(
		zap.Uint16("packetID", cPacket.PacketID()),
		zap.String("type", cPacket.PacketType()),
	).Info("Writing packet")
}

func handleConnect(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	connectPacket, ok := controlPacket.Content.(*packets.Connect)
	if !ok {
		return errors.New("invalid packet")
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

	_, err := connAckPacket.WriteTo(readWriter)
	return err
}

func handleDisconnect(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	disconnectPacket, ok := controlPacket.Content.(*packets.Disconnect)
	if !ok {
		return errors.New("invalid packet")
	}

	base.Disconnect(readWriter, disconnectPacket)
	return nil
}

func handlePingRequest(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	_, ok := controlPacket.Content.(*packets.Pingreq)
	if !ok {
		return errors.New("invalid packet")
	}

	pingResponsePacket := packets.Pingresp{}

	_, err := pingResponsePacket.WriteTo(readWriter)
	return err
}

func handlePublish(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	publishPacket, ok := controlPacket.Content.(*packets.Publish)
	if !ok {
		return errors.New("invalid packet")
	}

	switch publishPacket.QoS {
	case 0:
		return handlePubQos0(publishPacket, base)
	case 1:
		return handlePubQoS1(readWriter, publishPacket, base)
	case 2:
		return handlePubQos2(readWriter, publishPacket, base)
	}

	return nil
}

func handlePubQos0(publishPacket *packets.Publish, base MqttBase) error {
	base.Publish(publishPacket)
	return nil
}

func handlePubQoS1(readWriter io.ReadWriter, publishPacket *packets.Publish, base MqttBase) error {
	pubAck := packets.Puback{
		ReasonCode: packets.PubackSuccess,
		PacketID:   publishPacket.PacketID,
	}

	_, err := pubAck.WriteTo(readWriter)
	if err != nil {
		return err
	}
	base.Publish(publishPacket)
	return nil
}

func handlePubQos2(readWriter io.ReadWriter, publishPacket *packets.Publish, base MqttBase) error {
	pubReceived := packets.Pubrec{
		ReasonCode: packets.PubrecSuccess,
		PacketID:   publishPacket.PacketID,
	}

	err := base.ReservePacketID(readWriter, publishPacket)
	if err != nil {
		pubReceived.ReasonCode = packets.PubrecImplementationSpecificError
	}

	_, err = pubReceived.WriteTo(readWriter)
	if err != nil {
		return err
	}
	base.Publish(publishPacket)
	return nil
}

func handlePubRel(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	pubRelPacket, ok := controlPacket.Content.(*packets.Pubrel)
	if !ok {
		return errors.New("invalid packet")
	}

	pubComplete := packets.Pubcomp{
		ReasonCode: packets.PubrecSuccess,
		PacketID:   pubRelPacket.PacketID,
	}

	_ = base.FreePacketID(readWriter, pubRelPacket)

	_, err := pubComplete.WriteTo(readWriter)
	return err
}

func handleSubscribe(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	subscribePacket, ok := controlPacket.Content.(*packets.Subscribe)
	if !ok {
		return errors.New("invalid packet")
	}

	subAck := packets.Suback{
		PacketID: subscribePacket.PacketID,
		Reasons:  base.Subscribe(readWriter, subscribePacket),
	}

	_, err := subAck.WriteTo(readWriter)
	return err
}

func handleUnsubscribe(readWriter io.ReadWriter, controlPacket *packets.ControlPacket, base MqttBase) error {
	unsubscribePacket, ok := controlPacket.Content.(*packets.Unsubscribe)
	if !ok {
		return errors.New("invalid packet")
	}

	unsubAck := packets.Unsuback{
		PacketID: unsubscribePacket.PacketID,
		Reasons:  base.Unsubscribe(readWriter, unsubscribePacket),
	}

	_, err := unsubAck.WriteTo(readWriter)
	return err
}
