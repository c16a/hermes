package persistence

import (
	"bytes"
	"encoding/gob"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/eclipse/paho.golang/packets"
)

type BadgerProvider struct {
	db *badger.DB
}

func NewBadgerProvider() (*BadgerProvider, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	return &BadgerProvider{db: db}, nil
}

func openDB() (*badger.DB, error) {
	opts := badger.DefaultOptions("").WithInMemory(true)
	return badger.Open(opts)
}

func (b *BadgerProvider) SaveForOfflineDelivery(clientId string, publish *packets.Publish) error {
	return b.db.Update(func(txn *badger.Txn) error {
		payloadBytes, err := getBytes(publish)
		if err != nil {
			return err
		}
		return txn.Set([]byte(clientId), payloadBytes)
	})
}

func (b *BadgerProvider) GetMissedMessages(clientId string) ([]*packets.Publish, error) {
	publish := new(packets.Publish)
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(clientId))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			publish, err = getPublishPacket(val)
			if err != nil {
				return err
			}
			return nil
		})
	})
	return []*packets.Publish{publish}, err
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
