package lib

import (
	"errors"
	"fmt"
	"github.com/c16a/hermes/config"
	"github.com/eclipse/paho.golang/packets"
	"net"
	"sync"
)

// ServerContext stores the state of the cluster node
type ServerContext struct {
	connectedClientsMap map[string]*ConnectedClient
	mu                  *sync.RWMutex
	config              *config.Config
}

// NewServerContext creates a new server context.
//
// This should only be called once per cluster node.
func NewServerContext(config *config.Config) *ServerContext {
	return &ServerContext{
		mu:                  &sync.RWMutex{},
		connectedClientsMap: make(map[string]*ConnectedClient, 0),
		config:              config,
	}
}

func (ctx *ServerContext) AddClient(conn net.Conn, connect *packets.Connect) (code byte, sessionExists bool) {
	clientExists := ctx.checkForClient(connect.ClientID)
	clientRequestForFreshSession := connect.CleanStart
	if clientExists {
		if clientRequestForFreshSession {
			// If client asks for fresh session, delete existing ones
			delete(ctx.connectedClientsMap, connect.ClientID)
			ctx.doAddClient(conn, connect)
		} else {
			ctx.doUpdateClient(connect.ClientID, conn)
		}
	} else {
		ctx.doAddClient(conn, connect)
	}
	return 0, clientExists && !clientRequestForFreshSession
}

func (ctx *ServerContext) doAddClient(conn net.Conn, connect *packets.Connect) {
	newClient := &ConnectedClient{
		Connection:   conn,
		ClientID:     connect.ClientID,
		IsClean:      connect.CleanStart,
		IsConnected:  true,
		Subscription: make(map[string]packets.SubOptions, 0),
	}
	ctx.mu.Lock()
	ctx.connectedClientsMap[connect.ClientID] = newClient
	ctx.mu.Unlock()
}

func (ctx *ServerContext) doUpdateClient(clientID string, conn net.Conn) {
	ctx.connectedClientsMap[clientID].Connection = conn
}

func (ctx *ServerContext) Disconnect(conn net.Conn, disconnect *packets.Disconnect) {
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
		if _, ok := client.Subscription[topicToTarget]; ok {
			if !client.IsConnected {
				continue
			}
			_, err := publish.WriteTo(client.Connection)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (ctx *ServerContext) Subscribe(conn net.Conn, subscribe *packets.Subscribe) []byte {
	var subAckBytes []byte
	for _, client := range ctx.connectedClientsMap {
		if conn == client.Connection {
			for topic, options := range subscribe.Subscriptions {
				client.Subscription[topic] = options

				var subAckByte byte
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
				subAckBytes = append(subAckBytes, subAckByte)
			}
		}
	}
	return subAckBytes
}

func (ctx *ServerContext) Unsubscribe(conn net.Conn, unsubscribe *packets.Unsubscribe) []byte {
	client, _ := ctx.getClientForConnection(conn)

	var unsubAckBytes []byte
	for _, topic := range unsubscribe.Topics {
		_, ok := client.Subscription[topic]
		if ok {
			delete(client.Subscription, topic)
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

func (ctx *ServerContext) getClientForConnection(conn net.Conn) (*ConnectedClient, error) {
	for _, client := range ctx.connectedClientsMap {
		if conn == client.Connection {
			return client, nil
		}
	}
	return nil, errors.New("client not found for connection")
}

func checkForTopicInArray(topic string, topics []string) bool {
	for _, t := range topics {
		if topic == t {
			return true
		}
	}
	return false
}

func convertToMapOfClients(clients []*ConnectedClient) map[string][]*ConnectedClient {
	var clientMap = make(map[string][]*ConnectedClient, 0)

	for _, client := range clients {
		groupedClients, ok := clientMap[client.ClientGroup]
		if ok {
			groupedClients = append(groupedClients, client)
			clientMap[client.ClientGroup] = groupedClients
		} else {
			clientMap[client.ClientGroup] = []*ConnectedClient{client}
		}
	}
	return clientMap
}

// ConnectedClient stores the information about a currently connected client
type ConnectedClient struct {
	Connection   net.Conn
	ClientID     string
	ClientGroup  string
	IsConnected  bool
	IsClean      bool
	Subscription map[string]packets.SubOptions
}
