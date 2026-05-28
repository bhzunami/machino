package realtime

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type      string `json:"type"`
	ProjectID string `json:"projectId,omitempty"`
}

type Hub struct {
	logger   *slog.Logger
	mu       sync.RWMutex
	clients  map[string]map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[string]map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				parsed, err := url.Parse(origin)
				return err == nil && parsed.Host == r.Host
			},
		},
	}
}

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, projectID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("upgrade websocket", "error", err)
		return
	}
	h.add(projectID, conn)
	defer h.remove(projectID, conn)
	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	})
	for {
		if _, _, err := conn.NextReader(); err != nil {
			return
		}
	}
}

func (h *Hub) Broadcast(projectID string) {
	payload, err := json.Marshal(Event{Type: "project_changed", ProjectID: projectID})
	if err != nil {
		h.logger.Error("marshal realtime event", "error", err)
		return
	}
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.clients[projectID]))
	for conn := range h.clients[projectID] {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.logger.Warn("write websocket event", "error", err)
			h.remove(projectID, conn)
		}
	}
}

func (h *Hub) add(projectID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[projectID] == nil {
		h.clients[projectID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[projectID][conn] = struct{}{}
}

func (h *Hub) remove(projectID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[projectID] != nil {
		delete(h.clients[projectID], conn)
		if len(h.clients[projectID]) == 0 {
			delete(h.clients, projectID)
		}
	}
	_ = conn.Close()
}
