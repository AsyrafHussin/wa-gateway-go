package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/config"
	"github.com/AsyrafHussin/wa-gateway-go/internal/cache"
	"github.com/AsyrafHussin/wa-gateway-go/internal/server"
	"github.com/AsyrafHussin/wa-gateway-go/internal/webhook"
	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %s\n", err)
		os.Exit(1)
	}

	// Setup logger
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger().Level(level)

	logger.Info().
		Int("port", cfg.Port).
		Str("host", cfg.Host).
		Str("dataDir", cfg.DataDir).
		Msg("starting wa-gateway-go")

	// Create components
	hub := ws.NewHub(logger)
	go hub.Run()

	dispatcher := webhook.NewDispatcher(cfg.WebhookURL, cfg.WebhookSecret, cfg.WebhookTimeout, logger)
	phoneCache := cache.NewPhoneCache(cfg.CacheTTL)
	manager := whatsapp.NewDeviceManager(cfg, hub, dispatcher, logger)

	// Auto-reconnect existing sessions
	ctx := context.Background()
	manager.AutoReconnect(ctx)

	// Create and start server
	srv := server.New(cfg, manager, hub, dispatcher, phoneCache, logger)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		logger.Info().Str("addr", addr).Msg("server listening")
		if err := srv.App.Listen(addr); err != nil {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	logger.Info().Msg("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	manager.ShutdownAll(shutdownCtx)
	hub.Shutdown()
	srv.App.Shutdown()

	logger.Info().Msg("goodbye")
}
