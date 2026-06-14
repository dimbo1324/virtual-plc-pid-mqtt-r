// Package smoke contains end-to-end smoke tests that exercise the full
// application stack without external services (no MQTT broker, no storage,
// no web server). All tests must pass without Docker or network access.
package smoke

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/app"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
)

func minimalConfig() config.Config {
	cfg := config.Default()
	cfg.MQTT.Enabled = false
	cfg.Web.Enabled = false
	cfg.Storage.Enabled = false
	cfg.PLC.ScanIntervalMS = 5
	cfg.PLC.PublishIntervalMS = 10
	cfg.PLC.UIUpdateIntervalMS = 10
	cfg.PLC.ScanOverrunWarningMS = 100
	return cfg
}

func TestSmoke_DefaultConfigIsValid(t *testing.T) {
	if err := config.Default().Validate(); err != nil {
		t.Fatalf("Default config is invalid: %v", err)
	}
}

func TestSmoke_AppStartsAndStops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := app.New(minimalConfig(), logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	if err := application.RunRuntime(ctx); err != nil {
		t.Fatalf("RunRuntime() error = %v", err)
	}
}

func TestSmoke_AppRunsFoundationCheck(t *testing.T) {
	cfg := config.Default()
	// Run() performs foundation checks only (no long-lived services).
	application := app.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() foundation check error = %v", err)
	}
}

func TestSmoke_MultipleLoopsStart(t *testing.T) {
	cfg := minimalConfig()
	if len(cfg.Loops) < 2 {
		t.Skip("default config has fewer than 2 loops")
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := app.New(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	if err := application.RunRuntime(ctx); err != nil {
		t.Fatalf("RunRuntime() with %d loops error = %v", len(cfg.Loops), err)
	}
}

// TestSmoke_DegradedStorageFallback verifies that when SQLite is unavailable
// and fallback_on_error is true, the app starts successfully and writes events
// to the JSONL fallback file instead of failing hard.
func TestSmoke_DegradedStorageFallback(t *testing.T) {
	tmp := t.TempDir()

	// Make SQLite path point to an impossible location (file used as directory).
	blocker := filepath.Join(tmp, "history.db")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := minimalConfig()
	cfg.Storage.Enabled = true
	cfg.Storage.Type = "sqlite"
	cfg.Storage.SQLitePath = filepath.Join(blocker, "nested.db") // open will fail
	cfg.Storage.EventsJSONLPath = filepath.Join(tmp, "events.jsonl")
	cfg.Storage.AppLogPath = ""
	cfg.Storage.RetentionMaxSamples = 1000
	cfg.Storage.WriteQueueSize = 32
	cfg.Storage.FallbackOnError = true
	cfg.Storage.FallbackType = "jsonl"

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := app.New(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	if err := application.RunRuntime(ctx); err != nil {
		t.Fatalf("RunRuntime() in degraded mode error = %v", err)
	}

	// JSONL fallback file must exist after at least one PLC event.
	if _, err := os.Stat(filepath.Join(tmp, "events.jsonl")); os.IsNotExist(err) {
		t.Error("events.jsonl not created; degraded JSONL fallback did not activate")
	}
}
