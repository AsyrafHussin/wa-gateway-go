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
	done       chan struct{}
	logger     zerolog.Logger
}

func NewHub(logger zerolog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

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
			h.mu.Lock()
			for client := range h.clients {
				if !client.IsAuthenticated() {
					continue
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.done)
}

func (h *Hub) Broadcast(token, event string, data interface{}) {
	msg := Message{
		Event: event,
		Token: token,
		Data:  data,
	}
	h.sendMessage(msg)
}

func (h *Hub) BroadcastWithMessage(token, event, message string) {
	msg := Message{
		Event:   event,
		Token:   token,
		Message: message,
	}
	h.sendMessage(msg)
}

func (h *Hub) sendMessage(msg Message) {
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
