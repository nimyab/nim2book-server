package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/app"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

// @title						Nim2Book api
// @version						1.0
// @BasePath					/api/v1
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Setup logger
	slogLogger := logger.New(logger.Config{
		Env: cfg.Env,
	})
	slog.SetDefault(slogLogger)

	// Create application
	application, err := app.New(cfg)
	if err != nil {
		slog.Error("Failed to create application", slog.Any("error", err))
		os.Exit(1)
	}

	// Start HTTP server in goroutine
	go func() {
		if err := application.Start(); err != nil {
			slog.Error("Server error", slog.Any("error", err))
		}
	}()

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		slog.Error("Shutdown error", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Application stopped gracefully")
}
