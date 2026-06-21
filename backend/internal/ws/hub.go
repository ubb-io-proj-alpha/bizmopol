package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Notification struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*client]bool),
	}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*client]bool)
	}
	h.clients[c.userID][c] = true
	slog.Info("WS: client connected", "user_id", c.userID, "total_for_user", len(h.clients[c.userID]))
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[c.userID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.clients, c.userID)
		}
	}
	close(c.send)
	slog.Info("WS: client disconnected", "user_id", c.userID)
}

func (h *Hub) SendToUser(userID string, notification Notification) {
	data, err := json.Marshal(notification)
	if err != nil {
		slog.Error("WS: failed to marshal notification", "error", err)
		return
	}

	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		slog.Debug("WS: no clients for user, notification dropped", "user_id", userID, "type", notification.Type)
		return
	}

	for c := range clients {
		select {
		case c.send <- data:
		default:
			slog.Warn("WS: client send buffer full, dropping", "user_id", userID)
		}
	}

	slog.Info("WS: notification sent", "user_id", userID, "type", notification.Type, "clients", len(clients))
}

func (h *Hub) HandleWS(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, _ := userID.(string)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WS: upgrade failed", "error", err)
		return
	}

	cl := &client{
		conn:   conn,
		send:   make(chan []byte, 32),
		userID: uid,
	}

	h.register(cl)

	go cl.writePump()
	go cl.readPump(h)
}

func (c *client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *client) readPump(h *Hub) {
	defer func() {
		h.unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
