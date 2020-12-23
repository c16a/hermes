package lib

import (
	"errors"
)

// ParseSubscribeCall parses an incoming TCP command into a subscribe call
//
// SUB [clientID] {clientGroup} [topic1]
// ClientID uniquely identifies a single client
// clientGroup identifies a group of clients acting as one
// At least one topic is mandatory
func ParseSubscribeCall(elements []string) (clientID string, clientGroup string, topic string, err error) {
	if len(elements) < 3 {
		return "", "", "", errors.New("clientID and topic to subscribe are mandatory")
	}

	clientID = elements[1]
	if len(elements) == 4 {
		clientGroup = elements[2]
		topic = elements[3]
	} else {
		topic = elements[2]
	}

	return
}

// ParsePublishCall parses an incoming TCP command into a publish call
//
// PUB [topic] [payload]
// At least one topic is mandatory
func ParsePublishCall(elements []string) (topic string, payload string, err error) {
	if len(elements) < 3 {
		return "", "", errors.New("topic and payload to publish to are mandatory")
	}
	return elements[1], elements[2], nil
}
