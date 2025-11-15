package logging

import (
	"log/slog"
	"os"
)

// NewLogger builds a structured logger with the provided level. Supported levels:
// debug, info, warn, error. Defaults to info on unknown values.
func NewLogger(level string) *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: false})
	logger := slog.New(handler)

	// Map level string to Level
	switch level {
	case "debug":
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "info":
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "warn", "warning":
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case "error":
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	default:
		// leave default info
		_ = level
	}

	return logger
}
