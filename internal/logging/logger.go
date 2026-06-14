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

// NewTeeLogger creates a text logger that writes to both w1 and w2 simultaneously.
// Useful for tee-ing log output to both stdout and a file.
func NewTeeLogger(level string, w1, w2 io.Writer) *slog.Logger {
	return NewTextLogger(level, io.MultiWriter(w1, w2))
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
