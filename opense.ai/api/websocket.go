package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins; restrict in production
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// handleWebSocket upgrades HTTP connections to WebSocket and manages
// bidirectional communication for streaming analysis updates.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WSClient{
		hub:  s.wsHub,
		send: make(chan WSMessage, 256),
	}

	s.wsHub.Register(client)

	// Start reader and writer goroutines
	go wsWritePump(conn, client)
	go wsReadPump(conn, client, s)
}

// wsReadPump pumps messages from the WebSocket connection to the hub.
func wsReadPump(conn *websocket.Conn, client *WSClient, s *Server) {
	defer func() {
		client.hub.Unregister(client)
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Parse incoming message
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle client messages (e.g., subscribe to ticker updates)
		switch msg.Type {
		case "subscribe":
			// Acknowledge subscription
			client.send <- WSMessage{
				Type: "subscribed",
				Data: msg.Data,
			}
		case "ping":
			client.send <- WSMessage{Type: "pong"}
		}
	}
}

// wsWritePump pumps messages from the hub to the WebSocket connection.
func wsWritePump(conn *websocket.Conn, client *WSClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("WebSocket marshal error: %v", err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

			// Flush queued messages
			n := len(client.send)
			for i := 0; i < n; i++ {
				nextMsg := <-client.send
				nextData, err := json.Marshal(nextMsg)
				if err != nil {
					continue
				}
				if err := conn.WriteMessage(websocket.TextMessage, nextData); err != nil {
					return
				}
			}

		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
