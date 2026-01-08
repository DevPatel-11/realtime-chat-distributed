package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Connection represents a WebSocket connection
type Connection struct {
	ID        string
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	LastPing  time.Time
	mu        sync.Mutex
}

// WSHandler manages WebSocket connections
type WSHandler struct {
	connections map[string]*Connection
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
	presence    PresenceStore
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler(presence PresenceStore) *WSHandler {
	return &WSHandler{
		connections: make(map[string]*Connection),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		presence: presence,
	}
}

// HandleConnection manages a single WebSocket connection
func (h *WSHandler) HandleConnection(ctx context.Context, conn *websocket.Conn, userID string) error {
	connID := uuid.New().String()
	
	// Check for existing connection
	h.mu.Lock()
	if existing, ok := h.connections[userID]; ok {
		// Close old connection
		existing.Close()
		delete(h.connections, userID)
	}
	
	c := &Connection{
		ID:       connID,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		LastPing: time.Now(),
	}
	
	h.connections[userID] = c
	h.mu.Unlock()
	
	// Store in Redis
	err := h.presence.SetUserSession(ctx, userID, connID)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}
	
	// Start read/write pumps
	go h.writePump(c)
	go h.readPump(ctx, c)
	
	return nil
}

// readPump reads messages from WebSocket
func (h *WSHandler) readPump(ctx context.Context, c *Connection) {
	defer func() {
		h.removeConnection(ctx, c)
	}()
	
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastPing = time.Now()
		c.mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("websocket error: %v\n", err)
			}
			break
		}
		
		// Process message (forward to chat service)
		fmt.Printf("Received from %s: %s\n", c.UserID, message)
	}
}

// writePump writes messages to WebSocket
func (h *WSHandler) writePump(c *Connection) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to a specific user
func (h *WSHandler) SendMessage(userID string, message []byte) error {
	h.mu.RLock()
	c, ok := h.connections[userID]
	h.mu.RUnlock()
	
	if !ok {
		return fmt.Errorf("user not connected: %s", userID)
	}
	
	select {
	case c.Send <- message:
		return nil
	default:
		return fmt.Errorf("send buffer full for user: %s", userID)
	}
}

// removeConnection cleans up a disconnected connection
func (h *WSHandler) removeConnection(ctx context.Context, c *Connection) {
	h.mu.Lock()
	delete(h.connections, c.UserID)
	h.mu.Unlock()
	
	close(c.Send)
	h.presence.RemoveUserSession(ctx, c.UserID)
}

// Close closes a connection
func (c *Connection) Close() {
	c.Conn.Close()
}
