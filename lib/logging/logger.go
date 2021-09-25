package logging

import (
	"github.com/eclipse/paho.golang/packets"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

var logOnce sync.Once

var loggerInstance *log.Logger

func GetLogger() *log.Logger {
	logOnce.Do(func() {
		loggerInstance = createLogger()
	})
	return loggerInstance
}

func createLogger() *log.Logger {
	var logger = log.New()
	logger.Out = os.Stdout
	logger.Level = log.DebugLevel
	logger.Formatter = &log.JSONFormatter{}
	return logger
}

func LogControlPacket(packet *packets.ControlPacket) {
	logger := GetLogger()
	logger.WithFields(log.Fields{
		"packetID": packet.PacketID(),
		"type":     getPacketType(packet.Type),
	}).Debug("Received packet")
}

func LogOutgoingPacket(packetType byte) {
	logger := GetLogger()
	logger.WithFields(log.Fields{
		"type": getPacketType(packetType),
	}).Debug("Writing packet")
}

func LogCustom(msg string, level log.Level) {
	logger := GetLogger()
	logger.WithFields(log.Fields{

	}).Log(level, msg)
}

func getPacketType(packetType byte) string {
	switch packetType {
	case packets.CONNECT:
		return "CONNECT"
	case packets.CONNACK:
		return "CONNACK"
	case packets.PUBLISH:
		return "PUBLISH"
	case packets.PUBACK:
		return "PUBACK"
	case packets.PUBREC:
		return "PUBREC"
	case packets.PUBREL:
		return "PUBREL"
	case packets.PUBCOMP:
		return "PUBCOMP"
	case packets.SUBSCRIBE:
		return "SUBSCRIBE"
	case packets.SUBACK:
		return "SUBACK"
	case packets.UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case packets.UNSUBACK:
		return "UNSUBACK"
	case packets.PINGREQ:
		return "PINREQ"
	case packets.PINGRESP:
		return "PINGRESP"
	case packets.DISCONNECT:
		return "DISCONNECT"
	case packets.AUTH:
		return "AUTH"
	}
	return ""
}
