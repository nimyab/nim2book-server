package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Env string
}

func New(cfg Config) *slog.Logger {
	switch cfg.Env {
	case "dev":
		return slog.New(NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "prod":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "local":
		return slog.New(NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
