package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (increased to support CRUD JSON payloads)
	maxMessageSize = 4096

	// Send channel buffer size - can hold upto 256 messages before blocking
	sendBufferSize = 256
)

// upgrader - converts http connection to ws
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles new WebSocket connection requests.
// The crudHandler parameter enables bidirectional CRUD operations over WebSocket.
//
// GET /ws?userId={uuid}
func HandleWebSocket(manager *Manager, crudHandler *CRUDHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, `{"error":"userId query parameter is required"}`, http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket: upgrade failed for user %s: %v", userID, err)
			return
		}

		client := &Client{
			UserID: userID,
			Conn:   conn,
			Send:   make(chan []byte, sendBufferSize),
		}

		manager.Register(client)

		// Start read and write pumps in separate goroutines
		go writePump(client, manager)
		go readPump(client, manager, crudHandler)
	}
}

// readPump reads messages from the WebSocket connection and forwards them
// to the CRUDHandler for processing. This enables bidirectional CRUD over WebSocket.
func readPump(client *Client, manager *Manager, crudHandler *CRUDHandler) {
	defer func() {
		manager.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket: unexpected close for user %s: %v", client.UserID, err)
			}
			break
		}

		// Forward the message to the CRUD handler for processing
		if crudHandler != nil && len(message) > 0 {
			crudHandler.HandleMessage(client.UserID, message)
		}
	}
}

// send messages to client and handle pings - sending notifications
func writePump(client *Client, manager *Manager) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Manager closed the channel — send close frame
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain any queued messages into the same write
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
