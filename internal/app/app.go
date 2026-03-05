package app

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
)

// App represents the application with all its dependencies
type App struct {
	config   *config.Config
	server   *echo.Echo
	adapters *Adapters
}

// New creates a new application instance without dependency injection container
func New(cfg *config.Config) (*App, error) {
	// Initialize WebSocket Hub
	websocket.NewAndStart()

	// Initialize adapters (infrastructure layer)
	adapters, err := newAdapters(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize repositories (data access layer)
	repos := newRepositories(adapters.EntClient)

	// Initialize services (use cases / business logic)
	services := newServices(repos, adapters, cfg)

	app := &App{
		config:   cfg,
		adapters: adapters,
	}

	// Setup HTTP server and routes
	app.setupHTTPServer(services)

	return app, nil
}

// Start starts the HTTP server
func (a *App) Start() error {
	slog.Info("Starting HTTP server", slog.String("port", a.config.Port))
	return a.server.Start(a.config.Port)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down application...")

	// Shutdown HTTP server first
	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", slog.Any("error", err))
		return err
	}
	slog.Info("HTTP server shut down successfully")

	// Close database connection
	if a.adapters.EntClient != nil {
		if err := a.adapters.EntClient.Close(); err != nil {
			slog.Error("Failed to close database connection", slog.Any("error", err))
		} else {
			slog.Info("Database connection closed successfully")
		}
	}

	// Close Redis connection
	if a.adapters.Redis != nil {
		if err := a.adapters.Redis.Close(); err != nil {
			slog.Error("Failed to close Redis connection", slog.Any("error", err))
		} else {
			slog.Info("Redis connection closed successfully")
		}
	}

	// Close gRPC client connection
	if a.adapters.WordAligner != nil {
		if err := a.adapters.WordAligner.Close(); err != nil {
			slog.Error("Failed to close gRPC connection", slog.Any("error", err))
		} else {
			slog.Info("gRPC connection closed successfully")
		}
	}

	return nil
}
