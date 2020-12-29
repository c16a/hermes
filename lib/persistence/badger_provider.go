package persistence

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/c16a/hermes/lib/config"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/eclipse/paho.golang/packets"
	uuid "github.com/satori/go.uuid"
)

const (
	PacketReserved byte = 1
)

type BadgerProvider struct {
	db *badger.DB
}

func NewBadgerProvider(config *config.Config) (*BadgerProvider, error) {
	db, err := openDB(config)
	if err != nil {
		return nil, err
	}
	return &BadgerProvider{db: db}, nil
}

func openDB(config *config.Config) (*badger.DB, error) {
	offlineConfig := config.Server.Offline

	var opts badger.Options
	if offlineConfig == nil {
		return nil, errors.New("offline configuration disabled")
	} else {
		if len(offlineConfig.Path) == 0 {
			opts = badger.DefaultOptions("").WithInMemory(true)
		} else {
			opts = badger.DefaultOptions(offlineConfig.Path)
		}
		opts.ValueLogLoadingMode = options.FileIO
		opts.NumMemtables = offlineConfig.NumTables
		opts.KeepL0InMemory = false
		opts.MaxTableSize = offlineConfig.MaxTableSize
	}

	return badger.Open(opts)
}

func (b *BadgerProvider) SaveForOfflineDelivery(clientId string, publish *packets.Publish) error {
	return b.db.Update(func(txn *badger.Txn) error {
		payloadBytes, err := getBytes(publish)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s:%s", clientId, uuid.NewV4().String())
		return txn.Set([]byte(key), payloadBytes)
	})
}

func (b *BadgerProvider) GetMissedMessages(clientID string) ([]*packets.Publish, error) {
	messages := make([]*packets.Publish, 0)

	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(clientID)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				publish, err := getPublishPacket(val)
				if err != nil {
					return err
				}
				messages = append(messages, publish)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return messages, err
}

func (b *BadgerProvider) ReservePacketID(clientID string, packetID uint16) error {
	return b.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("packet:%s:%d", clientID, packetID)
		return txn.Set([]byte(key), []byte{PacketReserved})
	})
}

func (b *BadgerProvider) FreePacketID(clientID string, packetID uint16) error {
	return b.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("packet:%s:%d", clientID, packetID)
		return txn.Delete([]byte(key))
	})
}

func (b *BadgerProvider) CheckForPacketIdReuse(clientID string, packetID uint16) (bool, error) {
	reuseFlag := false
	err := b.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("packet:%s:%d", clientID, packetID)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			if val[0] == PacketReserved {
				reuseFlag = true
			} else {
				return errors.New("some weird error")
			}
			return nil
		})
	})
	return reuseFlag, err
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
