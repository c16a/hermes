package persistence

import "github.com/eclipse/paho.golang/packets"

type Provider interface {
	SaveForOfflineDelivery(clientId string, publish *packets.Publish) error
	GetMissedMessages(clientId string) ([]*packets.Publish, error)

	ReservePacketID(clientID string, packetID uint16) error
	FreePacketID(clientID string, packetID uint16) error
	CheckForPacketIdReuse(clientID string, packetID uint16) (bool, error)
}
