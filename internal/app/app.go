package app

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
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

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	if err := a.injector.Shutdown(); err != nil {
		return err
	}

	slog.Info("Application shut down successfully")
	return nil
}

// GetInjector returns the DI injector for testing purposes
func (a *App) GetInjector() *do.RootScope {
	return a.injector
}
