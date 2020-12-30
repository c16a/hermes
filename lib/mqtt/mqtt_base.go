package mqtt

import (
	"github.com/eclipse/paho.golang/packets"
	"io"
)

type MqttBase interface {
	AddClient(io.Writer, *packets.Connect) (reasonCode byte, sessionExists bool, maxQos byte)
	Disconnect(io.Writer, *packets.Disconnect)
	Publish(*packets.Publish)
	Subscribe(io.Writer, *packets.Subscribe) []byte
	Unsubscribe(io.Writer, *packets.Unsubscribe) []byte

	ReservePacketID(io.Writer, *packets.Publish) error
	FreePacketID(io.Writer, *packets.Pubrel) error
}
