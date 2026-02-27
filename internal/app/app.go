package app

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner"
	"github.com/samber/do/v2"
)

// App represents the application with all its dependencies
type App struct {
	injector *do.RootScope
	config   *config.Config
	server   *echo.Echo
}

// New creates a new application instance with dependency injection
func New(cfg *config.Config) (*App, error) {
	injector := do.New()

	app := &App{
		injector: injector,
		config:   cfg,
	}

	// Register config
	do.ProvideValue(injector, cfg)

	// Initialize WebSocket Hub
	websocket.NewAndStart()

	// Register adapters (infrastructure layer)
	app.registerAdapters()

	// Register repositories (data access layer)
	app.registerRepositories()

	// Register services (use cases / business logic)
	app.registerServices()

	// Setup HTTP server and routes
	app.setupHTTPServer()

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
	if client, err := do.Invoke[*ent.Client](a.injector); err == nil && client != nil {
		if err := client.Close(); err != nil {
			slog.Error("Failed to close database connection", slog.Any("error", err))
		} else {
			slog.Info("Database connection closed successfully")
		}
	}

	// Close Redis connection
	if redisCache, err := do.Invoke[*redis_cache.RedisCache](a.injector); err == nil && redisCache != nil {
		if err := redisCache.Close(); err != nil {
			slog.Error("Failed to close Redis connection", slog.Any("error", err))
		} else {
			slog.Info("Redis connection closed successfully")
		}
	}

	// Close gRPC client connection
	if grpcClient, err := do.Invoke[*word_aligner.Client](a.injector); err == nil && grpcClient != nil {
		if err := grpcClient.Close(); err != nil {
			slog.Error("Failed to close gRPC connection", slog.Any("error", err))
		} else {
			slog.Info("gRPC connection closed successfully")
		}
	}

	// Shutdown DI container
	if err := a.injector.Shutdown(); err != nil {
		slog.Error("Failed to shutdown DI container", slog.Any("error", err))
		return err
	}

	slog.Info("Application shut down successfully")
	return nil
}

// GetInjector returns the DI injector for testing purposes
func (a *App) GetInjector() *do.RootScope {
	return a.injector
}
