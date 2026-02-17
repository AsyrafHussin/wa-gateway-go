package handler

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
)

type WS struct {
	hub     *ws.Hub
	handler fiber.Handler
}

func NewWS(hub *ws.Hub) *WS {
	h := &WS{hub: hub}
	h.handler = websocket.New(h.handleConnection)
	return h
}

func (h *WS) Upgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return h.handler(c)
	}
	return fiber.ErrUpgradeRequired
}

func (h *WS) handleConnection(c *websocket.Conn) {
	client := ws.NewClient(h.hub, c)
	h.hub.Register(client)

	go client.WritePump()
	client.ReadPump()
}
