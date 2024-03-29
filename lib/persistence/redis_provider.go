package persistence

import (
	"context"
	"errors"
	"fmt"
	"github.com/c16a/hermes/lib/config"
	"github.com/eclipse/paho.golang/packets"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

type RedisProvider struct {
	client *redis.Client
}

func NewRedisProvider(config *config.Config, logger *zap.Logger) (Provider, error) {
	offlineConfig := config.Server.Persistence.Redis

	rdb := redis.NewClient(&redis.Options{
		Addr:     offlineConfig.Url,
		Password: offlineConfig.Password,
		DB:       0,
	})

	err := rdb.Echo(context.Background(), "HELLO").Err()
	if err != nil {
		logger.Error("Could not connect to redis persistence provider", zap.Error(err))
		return nil, err
	} else {
		logger.Info("Connected to redis persistence provider")
	}
	return &RedisProvider{client: rdb}, nil
}

func (r *RedisProvider) SaveForOfflineDelivery(clientId string, publish *packets.Publish) error {
	_, err := r.client.TxPipelined(context.Background(), func(pipeliner redis.Pipeliner) error {
		key := fmt.Sprintf("urn:messages:%s", clientId)

		publishBytes, err := getBytes(publish)
		if err != nil {
			return err
		}
		pipeliner.LPush(context.Background(), key, publishBytes)

		// Set expiry
		if publish.Properties != nil && publish.Properties.MessageExpiry != nil {
			pipeliner.Expire(context.Background(), key, time.Duration(int(*publish.Properties.MessageExpiry))*time.Second)
		}
		return nil
	})
	return err
}

func (r *RedisProvider) GetMissedMessages(clientId string) ([]*packets.Publish, error) {
	publishPackets := make([]*packets.Publish, 0)
	key := fmt.Sprintf("urn:messages:%s", clientId)

	// Get the length of the list
	length, err := r.client.LLen(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	// Pop everything in the list
	payloads, err := r.client.LPopCount(context.Background(), key, int(length)).Result()
	if err != nil {
		return nil, err
	}

	for _, payload := range payloads {
		payloadBytes := []byte(payload)
		publishPacket, err := getPublishPacket(payloadBytes)
		if err != nil {
			continue
		}
		publishPackets = append(publishPackets, publishPacket)
	}
	return publishPackets, err

}

func (r *RedisProvider) ReservePacketID(clientID string, packetID uint16) error {
	_, err := r.client.TxPipelined(context.Background(), func(pipeliner redis.Pipeliner) error {
		key := fmt.Sprintf("urn:packets:%s:%d", clientID, packetID)
		pipeliner.Set(context.Background(), key, PacketReserved, 24*time.Hour)
		return nil
	})
	return err
}

func (r *RedisProvider) FreePacketID(clientID string, packetID uint16) error {
	_, err := r.client.TxPipelined(context.Background(), func(pipeliner redis.Pipeliner) error {
		key := fmt.Sprintf("urn:packets:%s:%d", clientID, packetID)
		pipeliner.Del(context.Background(), key)
		return nil
	})
	return err
}

func (r *RedisProvider) CheckForPacketIdReuse(clientID string, packetID uint16) (bool, error) {
	reuseFlag := false
	_, err := r.client.TxPipelined(context.Background(), func(pipeliner redis.Pipeliner) error {
		key := fmt.Sprintf("urn:packets:%s:%d", clientID, packetID)
		resBytes, err := pipeliner.Get(context.Background(), key).Bytes()
		if err != nil {
			return err
		}
		if resBytes[0] == PacketReserved {
			reuseFlag = true
		} else {
			return errors.New("some weird error")
		}
		return nil
	})
	return reuseFlag, err
}
