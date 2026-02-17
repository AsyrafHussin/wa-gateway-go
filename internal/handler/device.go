package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

type Device struct {
	manager *whatsapp.DeviceManager
	logger  zerolog.Logger
}

func NewDevice(manager *whatsapp.DeviceManager, logger zerolog.Logger) *Device {
	return &Device{manager: manager, logger: logger}
}

type connectRequest struct {
	Token  string `json:"token"`
	Method string `json:"method"` // "qr" or "code"
}

func (h *Device) Connect(c *fiber.Ctx) error {
	var req connectRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if req.Token == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token (phone number) is required")
	}

	if req.Method == "" {
		req.Method = "qr"
	}

	if req.Method != "qr" && req.Method != "code" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_METHOD", "Method must be 'qr' or 'code'")
	}

	h.logger.Info().Str("token", req.Token).Str("method", req.Method).Msg("connecting device")

	if err := h.manager.Connect(c.Context(), req.Token, req.Method); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CONNECTION_FAILED", err.Error())
	}

	msg := "QR code sent via WebSocket"
	if req.Method == "code" {
		msg = "Pairing code sent via WebSocket"
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"token":  req.Token,
		"method": req.Method,
	}, msg)
}

func (h *Device) Disconnect(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	h.logger.Info().Str("token", token).Msg("disconnecting device")

	if err := h.manager.Disconnect(c.Context(), token); err != nil {
		return response.Error(c, fiber.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{"token": token}, "Device disconnected and logged out")
}
