package ws

import (
	"encoding/json"
	"sync"

	"github.com/rs/zerolog"
)

type Message struct {
	Event   string      `json:"event"`
	Token   string      `json:"token"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	logger     zerolog.Logger
}

func NewHub(logger zerolog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Debug().Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Debug().Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(token, event string, data interface{}) {
	msg := Message{
		Event: event,
		Token: token,
		Data:  data,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to marshal WebSocket message")
		return
	}

	select {
	case h.broadcast <- bytes:
	default:
		h.logger.Warn().Msg("WebSocket broadcast channel full, dropping message")
	}
}

func (h *Hub) BroadcastWithMessage(token, event, message string) {
	msg := Message{
		Event:   event,
		Token:   token,
		Message: message,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to marshal WebSocket message")
		return
	}

	select {
	case h.broadcast <- bytes:
	default:
		h.logger.Warn().Msg("WebSocket broadcast channel full, dropping message")
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
