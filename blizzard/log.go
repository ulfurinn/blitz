package blizzard

import (
	"log"
	"sync"
)

var logger *log.Logger
var loggerMutex sync.RWMutex

func Logger() *log.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return logger
}

func SetLogger(l *log.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	logger = l
}
