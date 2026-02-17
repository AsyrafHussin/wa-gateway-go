package handler

import (
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

type WS struct {
	hub            *ws.Hub
	handler        fiber.Handler
	apiKey         []byte
	authTimeout    time.Duration
	allowedOrigins []string
	logger         zerolog.Logger
}

func NewWS(hub *ws.Hub, apiKey string, authTimeout time.Duration, allowedOrigins string, logger zerolog.Logger) *WS {
	h := &WS{
		hub:         hub,
		apiKey:      []byte(apiKey),
		authTimeout: authTimeout,
		logger:      logger,
	}

	// Parse comma-separated origins, default to wildcard if empty
	for _, origin := range strings.Split(allowedOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			h.allowedOrigins = append(h.allowedOrigins, origin)
		}
	}
	if len(h.allowedOrigins) == 0 {
		h.allowedOrigins = []string{"*"}
	}

	h.handler = websocket.New(h.handleConnection)
	return h
}

func (h *WS) Upgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	origin := c.Get("Origin")
	if !h.isOriginAllowed(origin) {
		h.logger.Warn().
			Str("origin", origin).
			Str("remote", c.IP()).
			Msg("WebSocket origin rejected")
		return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "Origin not allowed")
	}

	return h.handler(c)
}

func (h *WS) isOriginAllowed(origin string) bool {
	for _, allowed := range h.allowedOrigins {
		if allowed == "*" {
			return true
		}
	}

	if origin == "" {
		return false
	}

	for _, allowed := range h.allowedOrigins {
		if strings.EqualFold(allowed, origin) {
			return true
		}
	}

	return false
}

func (h *WS) handleConnection(c *websocket.Conn) {
	client := ws.NewClient(h.hub, c, h.apiKey, h.logger)
	h.hub.Register(client)

	go client.AuthTimeoutPump(h.authTimeout)
	go client.WritePump()
	client.ReadPump()
}
