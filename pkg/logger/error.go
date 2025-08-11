package logger

import "log/slog"

func Error(msg string, err error, operation string) {
	slog.Error(
		msg,
		slog.String("error", err.Error()),
		slog.String("operation", operation),
	)
}
