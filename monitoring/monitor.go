package monitoring

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Monitor struct {
	connectionPool map[string]*websocket.Conn

	// TODO add some other components to be monitored.
}

var monitor = Monitor{connectionPool: make(map[string]*websocket.Conn)}

func GetMonitor() *Monitor {
	return &monitor
}

func (m *Monitor) AddConnection(c *websocket.Conn) {
	// Using pointer address as connection ID - Hope there is no collision :)
	connectionId := fmt.Sprintf("%p", c)
	if _, hasKey := m.connectionPool[connectionId]; hasKey {
		log.Println("WARNING: Cannot add connection ID that is already being monitored: ", connectionId)
	} else {
		log.Println("DEBUG: Adding connection ID to monitoring pool", connectionId)
		m.connectionPool[connectionId] = c
	}
}

func (m *Monitor) RemoveConnection(c *websocket.Conn) {
	connectionId := fmt.Sprintf("%p", c)
	if _, hasKey := m.connectionPool[connectionId]; hasKey {
		log.Println("DEBUG: Deleting connection ID from monitoring pool: ", connectionId)
		delete(m.connectionPool, connectionId)
	} else {
		log.Println("WARNING: Trying to remove non-existent connection ID from monitor pool: ", connectionId)
	}
}

func (m *Monitor) GetStats() string {
	return fmt.Sprintln("ClientPoolSize: ", len(m.connectionPool))
}
