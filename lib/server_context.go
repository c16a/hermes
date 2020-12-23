package lib

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type ServerContext struct {
	connectedClients []*ConnectedClient
}

func NewServerContext() *ServerContext {
	return &ServerContext{}
}

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
		ctx.connectedClients = append(ctx.connectedClients, newClient)
	}
	return nil
}

func (ctx *ServerContext) Publish(topic string, payload string) {
	var eligibleGroupedClients []*ConnectedClient
	for _, client := range ctx.connectedClients {
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

func (ctx *ServerContext) RemoveClient(conn net.Conn) {
	var indexToRemove int
	for index, client := range ctx.connectedClients {
		if client.Connection.RemoteAddr().String() == conn.RemoteAddr().String() {
			indexToRemove = index
		}
	}

	// Removing indexToRemove and not caring about the order
	ctx.connectedClients[indexToRemove] = ctx.connectedClients[len(ctx.connectedClients)-1]
	ctx.connectedClients = ctx.connectedClients[:len(ctx.connectedClients)-1]
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
	newAddr := conn.RemoteAddr().String()
	for _, existingClient := range ctx.connectedClients {
		oldAddr := existingClient.Connection.RemoteAddr().String()
		oldClientID := existingClient.ClientID

		if clientID == oldClientID && newAddr == oldAddr {
			return true, nil
		} else {
			if clientID == oldClientID && newAddr != oldAddr {
				return true, errors.New("clientID was used elsewhere")
			} else if clientID != oldClientID && newAddr == oldAddr {
				return true, errors.New("this connection was previously bound to another clientID")
			}
		}
	}
	return false, nil
}

func (ctx *ServerContext) subscribe(clientID string, topic string) {
	for _, client := range ctx.connectedClients {
		if clientID == client.ClientID {
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

type ConnectedClient struct {
	Connection  net.Conn
	Topics      []string
	ClientID    string
	ClientGroup string
	IsActive    bool
}
