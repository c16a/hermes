package mqtt

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/c16a/hermes/lib/auth"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/persistence"
	"github.com/c16a/hermes/lib/utils"
	"github.com/eclipse/paho.golang/packets"
	"go.uber.org/zap"
)

// ServerContext stores the state of the cluster node
type ServerContext struct {
	connectedClientsMap map[string]*ConnectedClient
	mu                  *sync.RWMutex
	config              *config.Config
	authProvider        auth.AuthorisationProvider
	persistenceProvider persistence.Provider

	logger *zap.Logger
}

// NewServerContext creates a new server context.
//
// This should only be called once per cluster node.
func NewServerContext(c *config.Config, logger *zap.Logger) (*ServerContext, error) {
	authProvider, err := auth.FetchProviderFromConfig(c)
	if err != nil {
		logger.Error("auth provider setup failed", zap.Error(err))
	}

	var providerSetupFn func(*config.Config, *zap.Logger) (persistence.Provider, error)
	var persistenceProvider persistence.Provider
	switch c.Server.Persistence.Type {
	case "memory":
		providerSetupFn = persistence.NewBadgerProvider
	case "redis":
		providerSetupFn = persistence.NewRedisProvider
	}

	if providerSetupFn == nil {
		logger.Error("persistence provider cannot be chosen")
	} else {
		persistenceProvider, err = providerSetupFn(c, logger)
		if err != nil {
			logger.Error("persistence provider setup failed", zap.Error(err))
		}
	}

	return &ServerContext{
		mu:                  &sync.RWMutex{},
		connectedClientsMap: make(map[string]*ConnectedClient, 0),
		config:              c,
		authProvider:        authProvider,
		persistenceProvider: persistenceProvider,
		logger:              logger,
	}, nil
}

func (ctx *ServerContext) AddClient(conn io.Writer, connect *packets.Connect) (code byte, sessionExists bool, maxQos byte) {
	maxQos = ctx.config.Server.MaxQos

	if ctx.authProvider != nil {
		if authError := ctx.authProvider.Validate(connect.Username, string(connect.Password)); authError != nil {
			code = 135
			sessionExists = false
			ctx.logger.Error("auth failed")
			return
		}
		ctx.logger.Info(fmt.Sprintf("auth succeeed for user: %s", connect.Username))
	}

	clientExists := ctx.checkForClient(connect.ClientID)
	clientRequestForFreshSession := connect.CleanStart
	if clientExists {
		if clientRequestForFreshSession {
			// If client asks for fresh session, delete existing ones
			ctx.logger.Info(fmt.Sprintf("Removing old connection for clientID: %s", connect.ClientID))
			delete(ctx.connectedClientsMap, connect.ClientID)
			ctx.doAddClient(conn, connect)
		} else {
			ctx.logger.Info(fmt.Sprintf("Updating clientID: %s with new connection", connect.ClientID))
			ctx.doUpdateClient(connect.ClientID, conn)
			if ctx.persistenceProvider != nil {
				ctx.logger.Info(fmt.Sprintf("Fetching missed messages for clientID: %s", connect.ClientID))
				err := ctx.sendMissedMessages(connect.ClientID, conn)
				if err != nil {
					ctx.logger.Error("failed to fetch offline messages", zap.Error(err))
				}
			}
		}
	} else {
		ctx.doAddClient(conn, connect)
	}
	code = 0
	sessionExists = clientExists && !clientRequestForFreshSession
	return
}

func (ctx *ServerContext) Disconnect(conn io.Writer, disconnect *packets.Disconnect) {
	var clientIdToRemove string
	shouldDelete := false
	for clientID, client := range ctx.connectedClientsMap {
		if client.Connection == conn {
			clientIdToRemove = clientID
			if client.IsClean {
				shouldDelete = true
			}
		}
	}

	if shouldDelete {
		ctx.logger.Info(fmt.Sprintf("Deleting connection for clientID: %s", clientIdToRemove))
		delete(ctx.connectedClientsMap, clientIdToRemove)
	} else {
		ctx.logger.Info(fmt.Sprintf("Marking connection as disconnected for clientID: %s", clientIdToRemove))
		ctx.mu.Lock()
		ctx.connectedClientsMap[clientIdToRemove].IsConnected = false
		ctx.mu.Unlock()
	}
}

// Publish publishes a message to a topic
func (ctx *ServerContext) Publish(publish *packets.Publish) {
	var shareNameClientMap = make(map[string][]*ConnectedClient, 0)
	for _, client := range ctx.connectedClientsMap {
		topicToTarget := publish.Topic
		for topicFilter := range client.Subscriptions {
			matches, isShared, shareName := utils.TopicMatches(topicToTarget, topicFilter)
			if matches {
				if !isShared {
					// non-shared subscriptions
					if !client.IsConnected && !client.IsClean && ctx.persistenceProvider != nil {
						// save for offline usage
						ctx.logger.Info(fmt.Sprintf("Saving offline delivery message for clientID: %s", client.ClientID))
						err := ctx.persistenceProvider.SaveForOfflineDelivery(client.ClientID, publish)
						if err != nil {
							ctx.logger.Error("failed to save offline message", zap.Error(err))
						}
					}
					if client.IsConnected {
						// send direct message
						publish.WriteTo(client.Connection)
					}
				} else {
					// share subscriptions
					if len(shareNameClientMap[shareName]) == 0 {
						shareNameClientMap[shareName] = make([]*ConnectedClient, 0)
					}
					shareNameClientMap[shareName] = append(shareNameClientMap[shareName], client)
				}
			}
		}
	}

	for _, clients := range shareNameClientMap {
		var client *ConnectedClient
		if len(clients) == 1 {
			client = clients[0]
		} else {
			rand.Seed(time.Now().Unix())
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s) // initialize local pseudorandom generator
			luckyClientIndex := r.Intn(len(clients))
			client = clients[luckyClientIndex]
		}
		publish.WriteTo(client.Connection)
	}
}

func (ctx *ServerContext) Subscribe(conn io.Writer, subscribe *packets.Subscribe) []byte {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	var subAckBytes []byte
	for _, client := range ctx.connectedClientsMap {
		if conn == client.Connection {
			for _, options := range subscribe.Subscriptions {
				client.Subscriptions[options.Topic] = options
				var subAckByte byte

				if options.QoS > ctx.config.Server.MaxQos {
					subAckByte = packets.SubackImplementationspecificerror
				} else {
					switch options.QoS {
					case 0:
						subAckByte = packets.SubackGrantedQoS0
					case 1:
						subAckByte = packets.SubackGrantedQoS1
					case 2:
						subAckByte = packets.SubackGrantedQoS2
					default:
						subAckByte = packets.SubackUnspecifiederror
					}
				}
				subAckBytes = append(subAckBytes, subAckByte)
			}
		}
	}
	return subAckBytes
}

func (ctx *ServerContext) Unsubscribe(conn io.Writer, unsubscribe *packets.Unsubscribe) []byte {
	client, _ := ctx.getClientForConnection(conn)

	var unsubAckBytes []byte
	for _, topic := range unsubscribe.Topics {
		_, ok := client.Subscriptions[topic]
		if ok {
			delete(client.Subscriptions, topic)
			unsubAckBytes = append(unsubAckBytes, packets.UnsubackSuccess)
		} else {
			unsubAckBytes = append(unsubAckBytes, packets.UnsubackNoSubscriptionFound)
		}
	}
	return unsubAckBytes
}

func (ctx *ServerContext) ReservePacketID(conn io.Writer, publish *packets.Publish) error {
	client, err := ctx.getClientForConnection(conn)
	if err != nil {
		return err
	}
	return ctx.persistenceProvider.ReservePacketID(client.ClientID, publish.PacketID)
}

func (ctx *ServerContext) FreePacketID(conn io.Writer, pubRel *packets.Pubrel) error {
	client, err := ctx.getClientForConnection(conn)
	if err != nil {
		return err
	}
	return ctx.persistenceProvider.FreePacketID(client.ClientID, pubRel.PacketID)
}

func (ctx *ServerContext) checkForClient(clientID string) bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	for oldClientID := range ctx.connectedClientsMap {
		if clientID == oldClientID {
			return true
		}
	}
	return false
}

func (ctx *ServerContext) getClientForConnection(conn io.Writer) (*ConnectedClient, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	for _, client := range ctx.connectedClientsMap {
		if conn == client.Connection {
			return client, nil
		}
	}
	return nil, errors.New("client not found for connection")
}

func (ctx *ServerContext) sendMissedMessages(clientId string, conn io.Writer) error {
	missedMessages, err := ctx.persistenceProvider.GetMissedMessages(clientId)
	if err != nil {
		return err
	}

	for _, msg := range missedMessages {
		if _, writeErr := msg.WriteTo(conn); writeErr != nil {
			if ctx.persistenceProvider.SaveForOfflineDelivery(clientId, msg) != nil {
				ctx.logger.Error("failed to save offline message", zap.Error(err))
			}
		}
	}
	return nil
}

func (ctx *ServerContext) doAddClient(conn io.Writer, connect *packets.Connect) {
	newClient := &ConnectedClient{
		Connection:    conn,
		ClientID:      connect.ClientID,
		IsClean:       connect.CleanStart,
		IsConnected:   true,
		Subscriptions: make(map[string]packets.SubOptions, 0),
	}

	ctx.logger.Info(fmt.Sprintf("Creating new connection for clientID: %s", connect.ClientID))
	ctx.mu.Lock()
	ctx.connectedClientsMap[connect.ClientID] = newClient
	ctx.mu.Unlock()
}

func (ctx *ServerContext) doUpdateClient(clientID string, conn io.Writer) {
	ctx.connectedClientsMap[clientID].Connection = conn
}

// ConnectedClient stores the information about a currently connected client
type ConnectedClient struct {
	Connection    io.Writer
	ClientID      string
	ClientGroup   string
	IsConnected   bool
	IsClean       bool
	Subscriptions map[string]packets.SubOptions
}
