package lib

import (
	"errors"
	"fmt"
	"github.com/c16a/hermes/lib/auth"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/persistence"
	"github.com/eclipse/paho.golang/packets"
	"io"
	"sync"
)

// ServerContext stores the state of the cluster node
type ServerContext struct {
	connectedClientsMap map[string]*ConnectedClient
	mu                  *sync.RWMutex
	config              *config.Config
	authProvider        auth.AuthorisationProvider
	persistenceProvider persistence.Provider
}

// NewServerContext creates a new server context.
//
// This should only be called once per cluster node.
func NewServerContext(config *config.Config) (*ServerContext, error) {
	authProvider, err := auth.FetchProviderFromConfig(config)
	if err != nil {
		fmt.Println("auth provider setup failed:", err)
	}

	persistenceProvider, err := persistence.NewBadgerProvider(config)
	if err != nil {
		fmt.Println("persistence provider setup failed:", err)
	}
	return &ServerContext{
		mu:                  &sync.RWMutex{},
		connectedClientsMap: make(map[string]*ConnectedClient, 0),
		config:              config,
		authProvider:        authProvider,
		persistenceProvider: persistenceProvider,
	}, nil
}

func (ctx *ServerContext) AddClient(conn io.Writer, connect *packets.Connect) (code byte, sessionExists bool) {
	clientExists := ctx.checkForClient(connect.ClientID)
	clientRequestForFreshSession := connect.CleanStart
	if clientExists {
		if clientRequestForFreshSession {
			// If client asks for fresh session, delete existing ones
			delete(ctx.connectedClientsMap, connect.ClientID)
			ctx.doAddClient(conn, connect)
		} else {
			ctx.doUpdateClient(connect.ClientID, conn)
			if ctx.persistenceProvider != nil {
				_ = ctx.sendMissedMessages(connect.ClientID, conn)
			}
		}
	} else {
		ctx.doAddClient(conn, connect)
	}
	return 0, clientExists && !clientRequestForFreshSession
}

func (ctx *ServerContext) doAddClient(conn io.Writer, connect *packets.Connect) {
	newClient := &ConnectedClient{
		Connection:    conn,
		ClientID:      connect.ClientID,
		IsClean:       connect.CleanStart,
		IsConnected:   true,
		Subscriptions: make(map[string]packets.SubOptions, 0),
	}
	ctx.mu.Lock()
	ctx.connectedClientsMap[connect.ClientID] = newClient
	ctx.mu.Unlock()
}

func (ctx *ServerContext) doUpdateClient(clientID string, conn io.Writer) {
	ctx.connectedClientsMap[clientID].Connection = conn
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
		delete(ctx.connectedClientsMap, clientIdToRemove)
	} else {
		ctx.connectedClientsMap[clientIdToRemove].IsConnected = false
	}
}

// Publish publishes a message to a topic
func (ctx *ServerContext) Publish(publish *packets.Publish) {
	for _, client := range ctx.connectedClientsMap {
		topicToTarget := publish.Topic
		if _, ok := client.Subscriptions[topicToTarget]; ok {
			if !client.IsConnected && !client.IsClean {
				// save for offline usage
				if ctx.persistenceProvider != nil {
					ctx.persistenceProvider.SaveForOfflineDelivery(client.ClientID, publish)
				}
			}
			if client.IsConnected {
				publish.WriteTo(client.Connection)
			}
		}
	}
}

func (ctx *ServerContext) Subscribe(conn io.Writer, subscribe *packets.Subscribe) []byte {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	var subAckBytes []byte
	for _, client := range ctx.connectedClientsMap {
		if conn == client.Connection {
			for topic, options := range subscribe.Subscriptions {
				client.Subscriptions[topic] = options
				var subAckByte byte

				if options.QoS > ctx.config.Server.MaxQos {
					subAckByte = packets.SubackImplementationspecificerror
				} else {
					switch options.QoS {
					case 0:
						subAckByte = packets.SubackGrantedQoS0
						break
					case 1:
						subAckByte = packets.SubackGrantedQoS1
						break
					case 2:
						subAckByte = packets.SubackGrantedQoS2
						break
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
		msg.WriteTo(conn)
	}
	return nil
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
