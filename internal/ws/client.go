package ws

import (
	"crypto/subtle"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/rs/zerolog"
)

const (
	writeWait          = 10 * time.Second
	pongWait           = 60 * time.Second
	pingPeriod         = (pongWait * 9) / 10
	maxAuthMessageSize = 1024
	maxReadMessageSize = 4096
)

type authMessage struct {
	Type   string `json:"type"`
	APIKey string `json:"apiKey"`
}

type authResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	authenticated atomic.Bool
	authDone      chan struct{}
	authOnce      sync.Once
	apiKey        []byte
	logger        zerolog.Logger
	writeDone     chan struct{}
}

func NewClient(hub *Hub, conn *websocket.Conn, apiKey []byte, logger zerolog.Logger) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		authDone:  make(chan struct{}),
		apiKey:    apiKey,
		logger:    logger,
		writeDone: make(chan struct{}),
	}
}

func (c *Client) IsAuthenticated() bool {
	return c.authenticated.Load()
}

func (c *Client) AuthTimeoutPump(timeout time.Duration) {
	select {
	case <-time.After(timeout):
		c.logger.Warn().Str("remote", c.remoteAddr()).Msg("WebSocket auth timeout")
		resp := authResponse{
			Type:    "auth",
			Success: false,
			Message: "Authentication timeout",
		}
		if data, err := json.Marshal(resp); err == nil {
			select {
			case c.send <- data:
			default:
			}
		}
		// Give WritePump time to flush the timeout message
		time.Sleep(100 * time.Millisecond)
		c.hub.Unregister(c)
		_ = c.conn.Close()
	case <-c.authDone:
		// Auth resolved (success or failure), nothing to do
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
		close(c.writeDone)
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxReadMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// First message must be auth
	if !c.handleAuth() {
		return
	}

	// After auth, just read and discard (keep connection alive for pong)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) handleAuth() bool {
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		c.closeAuthDone()
		return false
	}

	// Reject oversized messages
	if len(msg) > maxAuthMessageSize {
		c.logger.Warn().Str("remote", c.remoteAddr()).Msg("WebSocket auth message too large")
		c.sendAuthError("Message too large")
		c.closeAuthDone()
		return false
	}

	var authMsg authMessage
	if err := json.Unmarshal(msg, &authMsg); err != nil {
		c.logger.Warn().Str("remote", c.remoteAddr()).Msg("WebSocket auth invalid JSON")
		c.sendAuthError("Invalid message format")
		c.closeAuthDone()
		return false
	}

	if authMsg.Type != "auth" {
		c.logger.Warn().Str("remote", c.remoteAddr()).Str("type", authMsg.Type).Msg("WebSocket auth wrong message type")
		c.sendAuthError("First message must be auth")
		c.closeAuthDone()
		return false
	}

	if len(authMsg.APIKey) == 0 {
		c.logger.Warn().Str("remote", c.remoteAddr()).Msg("WebSocket auth empty API key")
		c.sendAuthError("API key is required")
		c.closeAuthDone()
		return false
	}

	if subtle.ConstantTimeCompare([]byte(authMsg.APIKey), c.apiKey) != 1 {
		c.logger.Warn().Str("remote", c.remoteAddr()).Msg("WebSocket auth invalid API key")
		c.sendAuthError("Invalid API key")
		c.closeAuthDone()
		return false
	}

	// Auth successful — close authDone first to cancel timeout, then set authenticated
	c.closeAuthDone()
	c.authenticated.Store(true)
	c.logger.Debug().Str("remote", c.remoteAddr()).Msg("WebSocket auth successful")

	resp := authResponse{Type: "auth", Success: true}
	if data, err := json.Marshal(resp); err == nil {
		select {
		case c.send <- data:
		default:
		}
	}

	return true
}

func (c *Client) closeAuthDone() {
	c.authOnce.Do(func() { close(c.authDone) })
}

func (c *Client) sendAuthError(message string) {
	resp := authResponse{
		Type:    "auth",
		Success: false,
		Message: message,
	}
	if data, err := json.Marshal(resp); err == nil {
		select {
		case c.send <- data:
		default:
		}
	}
	// Give WritePump time to flush the error message
	time.Sleep(100 * time.Millisecond)
}

func (c *Client) WaitWriteDone() {
	<-c.writeDone
}

func (c *Client) remoteAddr() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return "unknown"
}
