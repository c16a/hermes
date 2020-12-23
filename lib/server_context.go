package lib

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
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

// AddSubscribingClient submits a new entry for client subscription
//
// If the ConnectedClient already exists, and there is a new topic being subscribed to,
// only the topic subscription is added.
func (ctx *ServerContext) AddSubscribingClient(conn net.Conn, clientID string, clientGroup string, topic string) error {
	clientExists, clientErr := ctx.checkForClient(conn, clientID)
	if clientErr != nil {
		fmt.Println(clientErr)
		return clientErr
	}
	if clientExists {
		// Just subscribe to the new topic
		fmt.Println("Existing client called SUB")
		ctx.subscribe(clientID, topic)
	} else {
		// Add new client
		fmt.Println("Adding new client")
		newClient := &ConnectedClient{
			Connection:  conn,
			Topics:      []string{topic},
			ClientID:    clientID,
			ClientGroup: clientGroup,
		}
		ctx.mu.Lock()
		ctx.connectedClientsMap[clientID] = newClient
		ctx.mu.Unlock()
	}
	return nil
}

// Publish publishes a message to a topic
//
// This supports client grouping and chooses one of the eligible clients under the group at random.
// This can later be switched to any weight-based algorithm.
func (ctx *ServerContext) Publish(topic string, payload string) {
	var eligibleGroupedClients []*ConnectedClient
	for _, client := range ctx.connectedClientsMap {
		if checkForTopicInArray(topic, client.Topics) {
			if client.ClientGroup == "" {
				_, err := client.Connection.Write([]byte(payload + "\n"))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				eligibleGroupedClients = append(eligibleGroupedClients, client)
			}
		}
	}

	for _, clients := range convertToMapOfClients(eligibleGroupedClients) {
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
		_, _ = client.Connection.Write([]byte(payload + "\n"))
	}
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

func (ctx *ServerContext) checkForClient(conn net.Conn, clientID string) (clientExists bool, clientIdMismatchErr error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	newAddr := conn.RemoteAddr().String()
	for oldClientID, existingClient := range ctx.connectedClientsMap {
		oldAddr := existingClient.Connection.RemoteAddr().String()

		if clientID == oldClientID && newAddr == oldAddr {
			return true, nil
		}

		if clientID == oldClientID && newAddr != oldAddr {
			return true, errors.New("clientID was used elsewhere")
		} else if clientID != oldClientID && newAddr == oldAddr {
			return true, errors.New("this connection was previously bound to another clientID")
		}
	}
	return false, nil
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
	Connection  net.Conn
	Topics      []string
	ClientID    string
	ClientGroup string
	IsActive    bool
}
