package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/validator"
)

type Message struct {
	manager   *whatsapp.DeviceManager
	validator *validator.Validator
	logger    zerolog.Logger
}

func NewMessage(manager *whatsapp.DeviceManager, v *validator.Validator, logger zerolog.Logger) *Message {
	return &Message{manager: manager, validator: v, logger: logger}
}

type sendRequest struct {
	Token string `json:"token"`
	To    string `json:"to"`
	Text  string `json:"text"`
}

func (h *Message) Send(c *fiber.Ctx) error {
	var req sendRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if req.Token == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	if err := validator.ValidateToken(req.Token); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "Token must be a phone number (7-15 digits)")
	}

	if err := h.validator.ValidateMessage(req.Text); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_MESSAGE", "Message text cannot be empty")
	}

	phone, err := h.validator.ValidatePhone(req.To)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PHONE", "Invalid phone number")
	}

	result, err := h.manager.SendText(c.Context(), req.Token, phone, req.Text)
	if err != nil {
		h.logger.Error().Err(err).Str("token", req.Token).Str("to", phone).Msg("failed to send message")
		return response.Error(c, fiber.StatusInternalServerError, "SEND_FAILED", "Failed to send message")
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"messageId": result.ID,
		"timestamp": result.Timestamp,
	}, "Message sent")
}
