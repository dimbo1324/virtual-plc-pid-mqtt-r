// Package smoke contains end-to-end smoke tests that exercise the full
// application stack without external services (no MQTT broker, no storage,
// no web server). All tests must pass without Docker or network access.
package smoke

import (
	"context"
	"io"
	"log/slog"
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
