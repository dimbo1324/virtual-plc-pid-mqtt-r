package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New creates a text logger that writes to stdout.
func New(level string) *slog.Logger {
	return NewTextLogger(level, os.Stdout)
}

// NewTextLogger creates a standard-library text logger for the supplied output.
func NewTextLogger(level string, output io.Writer) *slog.Logger {
	options := &slog.HandlerOptions{Level: parseLevel(level)}
	return slog.New(slog.NewTextHandler(output, options))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
