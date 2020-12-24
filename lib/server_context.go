package lib

import (
	"fmt"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	"net"
	"sync"
)

// ServerContext stores the state of the cluster node
type ServerContext struct {
	connectedClientsMap map[string]*ConnectedClient
	mu                  *sync.RWMutex
}

// NewServerContext creates a new server context.
//
// This should only be called once per cluster node.
func NewServerContext() *ServerContext {
	return &ServerContext{
		mu:                  &sync.RWMutex{},
		connectedClientsMap: make(map[string]*ConnectedClient, 0),
	}
}

func (ctx *ServerContext) AddClient(conn net.Conn, connect *packets.Connect) byte {
	clientExists := ctx.checkForClient(connect.ClientID)

	if clientExists {
		return 130
	}
	// Add new client
	fmt.Println("Adding new client")
	newClient := &ConnectedClient{
		Connection: conn,
		ClientID:   connect.ClientID,
	}
	ctx.mu.Lock()
	ctx.connectedClientsMap[connect.ClientID] = newClient
	ctx.mu.Unlock()
	return 0
}

// Publish publishes a message to a topic
//
// This supports client grouping and chooses one of the eligible clients under the group at random.
// This can later be switched to any weight-based algorithm.
func (ctx *ServerContext) Publish(publish *packets.Publish) {

}

// RemoveClient removes a ConnectedClient from the ServerContext
func (ctx *ServerContext) RemoveClient(conn net.Conn) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	var clientIdToRemove string
	for clientID, client := range ctx.connectedClientsMap {
		if client.Connection.RemoteAddr().String() == conn.RemoteAddr().String() {
			clientIdToRemove = clientID
		}
	}

	// Removing indexToRemove and not caring about the order
	delete(ctx.connectedClientsMap, clientIdToRemove)
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

func (ctx *ServerContext) subscribe(clientID string, topic string) {
	for id, client := range ctx.connectedClientsMap {
		if clientID == id {
			client.Topics = append(client.Topics, topic)
		}
	}
}

func checkForTopicInArray(topic string, topics []string) bool {
	for _, t := range topics {
		if topic == t {
			return true
		}
	}
	return false
}

// ConnectedClient stores the information about a currently connected client
type ConnectedClient struct {
	Connection   net.Conn
	Topics       []string
	ClientID     string
	ClientGroup  string
	IsConnected  bool
	Subscription map[string]paho.SubscribeOptions
}
