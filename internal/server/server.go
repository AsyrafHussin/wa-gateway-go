package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/AsyrafHussin/wa-gateway-go/config"
	"github.com/AsyrafHussin/wa-gateway-go/internal/cache"
	"github.com/AsyrafHussin/wa-gateway-go/internal/handler"
	"github.com/AsyrafHussin/wa-gateway-go/internal/middleware"
	"github.com/AsyrafHussin/wa-gateway-go/internal/webhook"
	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/validator"
	"github.com/rs/zerolog"
)

type Server struct {
	App     *fiber.App
	Config  *config.Config
	Manager *whatsapp.DeviceManager
	Hub     *ws.Hub
	Logger  zerolog.Logger
}

func New(cfg *config.Config, manager *whatsapp.DeviceManager, hub *ws.Hub, dispatcher *webhook.Dispatcher, phoneCache *cache.PhoneCache, logger zerolog.Logger) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": err.Error(),
				},
			})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
	}))
	app.Use(middleware.RequestLogger(logger))

	v := validator.New(cfg.PhoneCountryCode, cfg.PhoneMinLength, cfg.PhoneMaxLength)
	auth := middleware.NewAuth(cfg.APIKey)

	// Health (no auth)
	healthHandler := handler.NewHealth(manager, logger)
	app.Get("/health", healthHandler.Basic)
	app.Get("/health/detailed", auth.Require(), healthHandler.Detailed)

	// WebSocket (no auth — browser can't set headers)
	wsHandler := handler.NewWS(hub)
	app.Get("/ws", wsHandler.Upgrade)

	// Authenticated routes
	api := app.Group("", auth.Require())

	deviceHandler := handler.NewDevice(manager, logger)
	api.Post("/devices", middleware.RateLimit(cfg.RateLimitDevices), deviceHandler.Connect)
	api.Delete("/devices/:token", middleware.RateLimit(cfg.RateLimitDevices), deviceHandler.Disconnect)

	messageHandler := handler.NewMessage(manager, v, logger)
	api.Post("/messages", middleware.RateLimit(cfg.RateLimitMessages), messageHandler.Send)

	validationHandler := handler.NewValidation(manager, v, phoneCache, logger)
	api.Post("/validate/phone", middleware.RateLimit(cfg.RateLimitValidate), validationHandler.ValidatePhone)

	contactHandler := handler.NewContact(manager, logger)
	api.Get("/contacts/:token", middleware.RateLimit(cfg.RateLimitMessages), contactHandler.List)

	cacheHandler := handler.NewCache(phoneCache, logger)
	api.Delete("/cache", middleware.RateLimit(cfg.RateLimitDevices), cacheHandler.Clear)

	return &Server{
		App:     app,
		Config:  cfg,
		Manager: manager,
		Hub:     hub,
		Logger:  logger,
	}
}
