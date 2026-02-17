package handler

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

var startTime = time.Now()

type Health struct {
	manager *whatsapp.DeviceManager
	logger  zerolog.Logger
}

func NewHealth(manager *whatsapp.DeviceManager, logger zerolog.Logger) *Health {
	return &Health{manager: manager, logger: logger}
}

func (h *Health) Basic(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, nil, "ok")
}

func (h *Health) Detailed(c *fiber.Ctx) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	sessions := h.manager.ListSessions()

	data := fiber.Map{
		"uptime":      time.Since(startTime).String(),
		"goroutines":  runtime.NumGoroutine(),
		"memoryAlloc": mem.Alloc,
		"memorySys":   mem.Sys,
		"deviceCount": len(sessions),
		"devices":     sessions,
	}

	return response.Success(c, fiber.StatusOK, data, "ok")
}
