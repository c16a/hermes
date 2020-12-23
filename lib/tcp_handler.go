package lib

import (
	"bufio"
	"net"
	"strings"
)

type SimpleTcpHandler struct {
}

func (s *SimpleTcpHandler) Handle(conn net.Conn, ctx *ServerContext) ([]byte, bool, error) {
	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, true, err
	}

	temp := strings.TrimSpace(data)
	if temp == "STOP" {
		return nil, true, err
	}

	if temp == "" {
		return nil, false, nil
	}

	elements := strings.Split(temp, " ")
	if len(elements) == 0 {
		return nil, false, nil
	}

	var responseBytes []byte
	command := elements[0]
	switch command {
	case "PUB":
		responseBytes, err = s.handlePublishCall(elements, ctx)
		break
	case "SUB":
		responseBytes, err = s.handleSubscribeCall(elements, conn, ctx)
	}

	return responseBytes, false, err
}

func (s *SimpleTcpHandler) handlePublishCall(elements []string, ctx *ServerContext) ([]byte, error) {
	topic, payload, err := ParsePublishCall(elements)
	if err != nil {
		return nil, err
	}
	ctx.Publish(topic, payload)
	return []byte("OK\n"), nil
}

func (s *SimpleTcpHandler) handleSubscribeCall(elements []string, conn net.Conn, ctx *ServerContext) ([]byte, error) {
	clientID, clientGroup, topic, err := ParseSubscribeCall(elements)
	if err != nil {
		return nil, err
	}

	err = ctx.AddSubscribingClient(conn, clientID, clientGroup, topic)
	if err != nil {
		return []byte(err.Error() + "\n"), err
	}
	return []byte("OK\n"), nil
}

func (s *SimpleTcpHandler) handleStopCall(conn net.Conn, ctx *ServerContext) ([]byte, error) {
	ctx.RemoveClient(conn)
	return nil, nil
}
