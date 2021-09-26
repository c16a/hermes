package persistence

import (
	"bytes"
	"encoding/gob"
	"github.com/eclipse/paho.golang/packets"
)

type Provider interface {
	SaveForOfflineDelivery(clientId string, publish *packets.Publish) error
	GetMissedMessages(clientId string) ([]*packets.Publish, error)

	ReservePacketID(clientID string, packetID uint16) error
	FreePacketID(clientID string, packetID uint16) error
	CheckForPacketIdReuse(clientID string, packetID uint16) (bool, error)
}

func getBytes(bundle interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(bundle)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getPublishPacket(src []byte) (*packets.Publish, error) {
	buf := bytes.NewBuffer(src)
	decoder := gob.NewDecoder(buf)

	var publish packets.Publish
	err := decoder.Decode(&publish)
	return &publish, err
}