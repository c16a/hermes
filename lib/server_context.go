package lib

import (
	"errors"
	"fmt"
	"github.com/c16a/hermes/lib/auth"
	"github.com/c16a/hermes/lib/config"
	"github.com/c16a/hermes/lib/persistence"
	"github.com/eclipse/paho.golang/packets"
	log "github.com/sirupsen/logrus"
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

func (ctx *ServerContext) AddClient(conn io.Writer, connect *packets.Connect) (code byte, sessionExists bool, maxQos byte) {
	maxQos = ctx.config.Server.MaxQos

	if ctx.authProvider != nil {
		if authError := ctx.authProvider.Validate(connect.Username, string(connect.Password)); authError != nil {
			code = 135
			sessionExists = false
			LogCustom("auth failed", log.ErrorLevel)
			return
		}
		LogCustom(fmt.Sprintf("auth succeeed for user: %s", connect.Username), log.DebugLevel)
	}

	clientExists := ctx.checkForClient(connect.ClientID)
	clientRequestForFreshSession := connect.CleanStart
	if clientExists {
		if clientRequestForFreshSession {
			// If client asks for fresh session, delete existing ones
			LogCustom(fmt.Sprintf("Removing old connection for clientID: %s", connect.ClientID), log.DebugLevel)
			delete(ctx.connectedClientsMap, connect.ClientID)
			ctx.doAddClient(conn, connect)
		} else {
			LogCustom(fmt.Sprintf("Updating clientID: %s with new connection", connect.ClientID), log.DebugLevel)
			ctx.doUpdateClient(connect.ClientID, conn)
			if ctx.persistenceProvider != nil {
				LogCustom(fmt.Sprintf("Fetching missed messages for clientID: %s", connect.ClientID), log.DebugLevel)
				_ = ctx.sendMissedMessages(connect.ClientID, conn)
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
		LogCustom(fmt.Sprintf("Deleting connection for clientID: %s", clientIdToRemove), log.DebugLevel)
		delete(ctx.connectedClientsMap, clientIdToRemove)
	} else {
		LogCustom(fmt.Sprintf("Marking connection as disconnected for clientID: %s", clientIdToRemove), log.DebugLevel)
		ctx.mu.Lock()
		ctx.connectedClientsMap[clientIdToRemove].IsConnected = false
		ctx.mu.Unlock()
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
					LogCustom(fmt.Sprintf("Saving offline delivery message for clientID: %s", client.ClientID), log.DebugLevel)
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
		msg.WriteTo(conn)
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

	LogCustom(fmt.Sprintf("Creating new connection for clientID: %s", connect.ClientID), log.DebugLevel)
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
