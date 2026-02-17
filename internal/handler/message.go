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

	if err := h.validator.ValidateMessage(req.Text); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_MESSAGE", err.Error())
	}

	phone, err := h.validator.ValidatePhone(req.To)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PHONE", err.Error())
	}

	result, err := h.manager.SendText(c.Context(), req.Token, phone, req.Text)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "SEND_FAILED", err.Error())
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"messageId": result.ID,
		"timestamp": result.Timestamp,
	}, "Message sent")
}
