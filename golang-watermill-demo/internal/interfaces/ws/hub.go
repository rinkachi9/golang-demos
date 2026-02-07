package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/gorilla/websocket"
)

type Hub struct {
	upgrader websocket.Upgrader
	logger   watermill.LoggerAdapter

	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
	outbox  chan []byte
}

func NewHub(logger watermill.LoggerAdapter) *Hub {
	return &Hub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logger:  logger,
		clients: make(map[*websocket.Conn]struct{}),
		outbox:  make(chan []byte, 256),
	}
}

func (h *Hub) Run() {
	for msg := range h.outbox {
		h.mu.RLock()
		for conn := range h.clients {
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				h.mu.RUnlock()
				h.removeClient(conn)
				h.mu.RLock()
			}
		}
		h.mu.RUnlock()
	}
}

func (h *Hub) Broadcast(payload []byte) {
	select {
	case h.outbox <- payload:
	default:
		h.logger.Error("websocket_outbox_full", nil, watermill.LogFields{
			"size": cap(h.outbox),
		})
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket_upgrade_failed", err, nil)
		return
	}
	h.addClient(conn)

	go h.readLoop(conn)
}

func (h *Hub) addClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
	h.logger.Info("websocket_connected", watermill.LogFields{
		"clients": len(h.clients),
	})
}

func (h *Hub) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	_ = conn.Close()
	delete(h.clients, conn)
	h.logger.Info("websocket_disconnected", watermill.LogFields{
		"clients": len(h.clients),
	})
}

func (h *Hub) readLoop(conn *websocket.Conn) {
	defer h.removeClient(conn)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
