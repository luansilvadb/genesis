package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
)

type Client struct {
	Conn     *websocket.Conn
	TenantID string
	Send     chan []byte

	mu        sync.Mutex
	writeMu   sync.Mutex // protects concurrent writes to Conn
	closed    bool
	closeOnce sync.Once
	done      chan struct{} // closed when the client is shutting down
}

func (c *Client) closeConn() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.Conn.Close()
	})
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]bool
}

const (
	// pongWait is the time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// pingPeriod is the interval between ping messages. Must be less than pongWait.
	pingPeriod = 30 * time.Second
	// maxMessageSize is the maximum message size allowed from the peer (4KB).
	maxMessageSize = 4096
)

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.TenantID] == nil {
		h.rooms[client.TenantID] = make(map[*Client]bool)
	}
	h.rooms[client.TenantID][client] = true
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.TenantID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			client.mu.Lock()
			if !client.closed {
				client.closed = true
				if client.Send != nil {
					close(client.Send)
				}
			}
			client.mu.Unlock()
			if len(clients) == 0 {
				delete(h.rooms, client.TenantID)
			}
		}
	}
}

func (h *Hub) Broadcast(tenantID string, msg dto.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("websocket marshal error: %v", err)
		return
	}

	h.mu.RLock()
	clients, ok := h.rooms[tenantID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	snapshot := make([]*Client, 0, len(clients))
	for client := range clients {
		snapshot = append(snapshot, client)
	}
	h.mu.RUnlock()

	for _, client := range snapshot {
		if !client.trySend(data) {
			h.Unregister(client)
		}
	}
}

func (h *Hub) BroadcastAll(msg dto.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("websocket marshal error: %v", err)
		return
	}

	h.mu.RLock()
	all := make([]*Client, 0)
	for _, clients := range h.rooms {
		for client := range clients {
			all = append(all, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range all {
		if !client.trySend(data) {
			h.Unregister(client)
		}
	}
}

// Run starts the hub's background tasks (currently a no-op — the hub is
// event-driven via Register/Unregister/Broadcast. Kept for future heartbeat
// or connection health-check functionality).
func (h *Hub) Run() {
	// Placeholder for future background tasks (e.g. heartbeat, stale connection cleanup).
}

func HandleClient(hub *Hub, conn *websocket.Conn, tenantID string) {
	// Set read limit to prevent memory exhaustion from large messages.
	conn.SetReadLimit(maxMessageSize)
	// Set initial read deadline. The pong handler will extend this.
	conn.SetReadDeadline(time.Now().Add(pongWait))
	// Register pong handler to extend read deadline on each pong.
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	client := &Client{
		Conn:     conn,
		TenantID: tenantID,
		Send:     make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	hub.Register(client)

	go client.writePump()
	client.readPump(hub)
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Unregister(c)
		c.closeConn()
	}()

	// Ping ticker in a separate goroutine so pings fire even while ReadMessage blocks.
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-c.done:
				return
			case <-ticker.C:
				c.mu.Lock()
				if c.closed {
					c.mu.Unlock()
					return
				}
				c.mu.Unlock()
				c.writeMu.Lock()
				c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.writeMu.Unlock()
					return
				}
				c.writeMu.Unlock()
			}
		}
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	}
}

func (c *Client) trySend(data []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.Send == nil {
		return false
	}

	select {
	case c.Send <- data:
		return true
	default:
		// Buffer full — message dropped. Log so operators can monitor.
		log.Printf("websocket: send buffer full for tenant %s, dropping message", c.TenantID)
		return false
	}
}

func (c *Client) writePump() {
	defer c.closeConn()

	for msg := range c.Send {
		c.writeMu.Lock()
		c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		c.writeMu.Unlock()
		if err != nil {
			return
		}
	}
}
