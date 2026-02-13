package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage is the WebSocket message format sent to clients
type WSMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    string      `json:"user_id"`
	Data      interface{} `json:"data"`
}

// Client represents a single WebSocket connection
type Client struct {
	UserID string          // Identifies which user owns this connection
	Conn   *websocket.Conn // Actual WebSocket connection
	Send   chan []byte     // Buffered channel for outgoing messages
}

type Manager struct {
	clients    map[string]*Client // userID → Client mapping
	mu         sync.RWMutex       // Thread-safe access to clients map
	register   chan *Client       // Channel for new connections
	unregister chan *Client       // Channel for disconnections
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the manager's main loop for handling register/unregister events. - go routine
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			// If the user already has a connection, close the old one
			if existing, ok := m.clients[client.UserID]; ok {
				close(existing.Send)
				existing.Conn.Close()
				log.Printf("WebSocket: replaced existing connection for user %s", client.UserID)
			}
			m.clients[client.UserID] = client
			m.mu.Unlock()
			log.Printf("WebSocket: client connected (user: %s, total: %d)", client.UserID, m.ClientCount())

		case client := <-m.unregister:
			m.mu.Lock()
			// Only remove if it's the same connection (prevents race with replacement)
			if existing, ok := m.clients[client.UserID]; ok && existing == client {
				close(client.Send)
				delete(m.clients, client.UserID)
			}
			m.mu.Unlock()
			log.Printf("WebSocket: client disconnected (user: %s, total: %d)", client.UserID, m.ClientCount())
		}
	}
}

// Register adds a client to the manager
func (m *Manager) Register(client *Client) {
	m.register <- client
}

// Unregister removes a client from the manager
func (m *Manager) Unregister(client *Client) {
	m.unregister <- client
}

// BroadcastToAll sends a message to ALL connected clients
func (m *Manager) BroadcastToAll(msg *WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WebSocket: failed to marshal broadcast message: %v", err)
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		select {
		case client.Send <- data:
		default:
			// Client's send buffer is full, skip
			log.Printf("WebSocket: send buffer full for user %s, skipping", client.UserID)
		}
	}
}

// BroadcastExcluding sends a message to all connected clients EXCEPT the specified user
func (m *Manager) BroadcastExcluding(msg *WSMessage, excludeUserID string) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WebSocket: failed to marshal broadcast message: %v", err)
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for userID, client := range m.clients {
		if userID == excludeUserID {
			continue
		}
		select {
		case client.Send <- data:
		default:
			log.Printf("WebSocket: send buffer full for user %s, skipping", client.UserID)
		}
	}
}

// ClientCount returns the number of connected clients
func (m *Manager) ClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}
