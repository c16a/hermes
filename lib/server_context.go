package lib

import (
	"fmt"
	"github.com/eclipse/paho.golang/packets"
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

func (ctx *ServerContext) Subscribe(conn net.Conn, subscribe *packets.Subscribe) {
	for _, client := range ctx.connectedClientsMap {
		if conn == client.Connection {
			for topic, options := range subscribe.Subscriptions {
				client.Subscription[topic] = options
			}
		}
	}
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
