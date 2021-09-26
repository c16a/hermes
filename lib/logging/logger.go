package logging

import (
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


