package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/cache"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

type Cache struct {
	phoneCache *cache.PhoneCache
	logger     zerolog.Logger
}

func NewCache(phoneCache *cache.PhoneCache, logger zerolog.Logger) *Cache {
	return &Cache{phoneCache: phoneCache, logger: logger}
}

func (h *Cache) Clear(c *fiber.Ctx) error {
	count := h.phoneCache.Count()
	h.phoneCache.Clear()
	h.logger.Info().Int("cleared", count).Msg("phone cache cleared")

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"cleared": count,
	}, "Cache cleared")
}
