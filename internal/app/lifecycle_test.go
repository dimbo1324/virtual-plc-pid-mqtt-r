package app

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func TestConfigMappingCreatesValidRuntime(t *testing.T) {
	mapped := mapPLCConfig(config.Default())
	runtime, err := plc.NewRuntime(mapped)
	if err != nil {
		t.Fatalf("NewRuntime(mapped) error = %v", err)
	}
	if len(runtime.Snapshot().Loops) != len(config.Default().Loops) {
		t.Fatalf("mapped loops = %d", len(runtime.Snapshot().Loops))
	}
}

func TestRunRuntimeStopsWithContext(t *testing.T) {
	cfg := config.Default()
	cfg.MQTT.Enabled = false
	cfg.PLC.ScanIntervalMS = 5
	cfg.PLC.PublishIntervalMS = 10
	cfg.PLC.UIUpdateIntervalMS = 10
	cfg.PLC.ScanOverrunWarningMS = 100
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := New(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := application.RunRuntime(ctx); err != nil {
		t.Fatalf("RunRuntime() error = %v", err)
	}
}
