package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/cache"
	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/validator"
)

type Validation struct {
	manager   *whatsapp.DeviceManager
	validator *validator.Validator
	cache     *cache.PhoneCache
	logger    zerolog.Logger
}

func NewValidation(manager *whatsapp.DeviceManager, v *validator.Validator, phoneCache *cache.PhoneCache, logger zerolog.Logger) *Validation {
	return &Validation{manager: manager, validator: v, cache: phoneCache, logger: logger}
}

type validateRequest struct {
	Token string `json:"token"`
	Phone string `json:"phone"`
}

func (h *Validation) ValidatePhone(c *fiber.Ctx) error {
	var req validateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if req.Token == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	if err := validator.ValidateToken(req.Token); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "Token must be a phone number (7-15 digits)")
	}

	phone, err := h.validator.ValidatePhone(req.Phone)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PHONE", "Invalid phone number")
	}

	// Check cache first
	if entry, found := h.cache.Get(phone); found {
		return response.Success(c, fiber.StatusOK, fiber.Map{
			"phone":        phone,
			"isOnWhatsApp": entry.IsOnWhatsApp,
			"jid":          entry.JID,
			"cached":       true,
		}, "Phone validation result (cached)")
	}

	results, err := h.manager.ValidatePhone(c.Context(), req.Token, phone)
	if err != nil {
		h.logger.Error().Err(err).Str("token", req.Token).Str("phone", phone).Msg("phone validation failed")
		return response.Error(c, fiber.StatusInternalServerError, "VALIDATION_FAILED", "Phone validation failed")
	}

	if len(results) == 0 {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Phone number not found on WhatsApp")
	}

	r := results[0]

	// Cache the result
	h.cache.Set(phone, cache.PhoneCacheEntry{
		IsOnWhatsApp: r.IsOnWhatsApp,
		JID:          r.JID,
	})

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"phone":        phone,
		"isOnWhatsApp": r.IsOnWhatsApp,
		"jid":          r.JID,
		"cached":       false,
	}, "Phone validation result")
}
