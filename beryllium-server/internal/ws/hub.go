package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message is a WebSocket push message.
type Message struct {
	Type       string `json:"type"`
	ClassID    string `json:"class_id"`
	Attendance any    `json:"attendance"`
}

// Hub manages WebSocket connections grouped by class ID.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

// NewHub creates a Hub.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*websocket.Conn]struct{}),
	}
}

// Subscribe adds a connection to a class room.
func (h *Hub) Subscribe(classID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[classID] == nil {
		h.rooms[classID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[classID][conn] = struct{}{}
	log.Printf("WS: client subscribed to class %s (total: %d)", classID[:8], len(h.rooms[classID]))
}

// Unsubscribe removes a connection from a class room.
func (h *Hub) Unsubscribe(classID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[classID] != nil {
		delete(h.rooms[classID], conn)
		if len(h.rooms[classID]) == 0 {
			delete(h.rooms, classID)
		}
	}
}

// Broadcast sends a message to all connections watching a class.
func (h *Hub) Broadcast(classID string, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := h.rooms[classID]
	if len(conns) == 0 {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WS: failed to marshal message: %v", err)
		return
	}

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("WS: write error, removing client: %v", err)
			conn.Close()
			go h.Unsubscribe(classID, conn)
		}
	}
}
