package handler

import (
	"log"
	"net/http"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler handles WebSocket connections for attendance updates.
type WSHandler struct {
	hub *ws.Hub
}

// NewWSHandler creates a WSHandler.
func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// Attendance handles GET /ws/classes/{id}/attendance — upgrades to WebSocket.
func (h *WSHandler) Attendance(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS: upgrade failed: %v", err)
		return
	}

	h.hub.Subscribe(classID, conn)

	// Read loop — keeps connection alive and handles client close
	go func() {
		defer func() {
			h.hub.Unsubscribe(classID, conn)
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
