package persistence

import "github.com/eclipse/paho.golang/packets"

type Provider interface {
	SaveForOfflineDelivery(clientId string, publish *packets.Publish) error
	GetMissedMessages(clientId string) ([]*packets.Publish, error)
}
