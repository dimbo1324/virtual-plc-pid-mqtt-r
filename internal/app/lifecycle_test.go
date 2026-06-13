package app

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func TestConfigMappingCreatesValidRuntime(t *testing.T) {
	cfg := config.Default()
	mapped := mapPLCConfig(cfg)
	runtime, err := plc.NewRuntime(mapped)
	if err != nil {
		t.Fatalf("NewRuntime(mapped) error = %v", err)
	}
	if len(runtime.Snapshot().Loops) != len(cfg.Loops) {
		t.Fatalf("mapped loops = %d", len(runtime.Snapshot().Loops))
	}
	if mapped.DeviceID != cfg.App.DeviceID || mapped.ScanInterval != 500*time.Millisecond {
		t.Fatalf("mapped PLC config = %+v", mapped)
	}
	pressure := mapped.Loops[0]
	if pressure.Name != "pressure" || pressure.PID.Kp != cfg.Loops[0].PID.Kp || pressure.Process.RandomSeed != 1 {
		t.Fatalf("mapped pressure loop = %+v", pressure)
	}
}

func TestConfigMappingRejectsInvalidLoopMode(t *testing.T) {
	cfg := config.Default()
	cfg.Loops[0].Mode = "cascade"
	if _, err := plc.NewRuntime(mapPLCConfig(cfg)); err == nil {
		t.Fatal("mapped invalid loop mode was accepted")
	}
}

func TestMQTTConfigMapping(t *testing.T) {
	cfg := config.Default()
	mapped := mapMQTTConfig(cfg)
	if mapped.Enabled != cfg.MQTT.Enabled || mapped.BrokerURL != cfg.MQTT.BrokerURL || mapped.BaseTopic != cfg.MQTT.BaseTopic {
		t.Fatalf("mapped MQTT config = %+v", mapped)
	}
	if mapped.ConnectTimeout != 5*time.Second || mapped.ReconnectInterval != 3*time.Second {
		t.Fatalf("mapped MQTT durations = %v/%v", mapped.ConnectTimeout, mapped.ReconnectInterval)
	}
}

func TestRunIsShortFoundationInitialization(t *testing.T) {
	var output bytes.Buffer
	application := New(config.Default(), slog.New(slog.NewTextHandler(&output, nil)))
	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !bytes.Contains(output.Bytes(), []byte("runtime foundation initialized")) {
		t.Fatalf("Run() output = %q", output.String())
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

func TestUnavailableBrokerDoesNotFailRuntime(t *testing.T) {
	cfg := config.Default()
	cfg.MQTT.BrokerURL = "tcp://127.0.0.1:1"
	cfg.MQTT.ConnectTimeoutSeconds = 1
	cfg.MQTT.ReconnectIntervalSeconds = 1
	cfg.PLC.ScanIntervalMS = 5
	cfg.PLC.PublishIntervalMS = 10
	cfg.PLC.UIUpdateIntervalMS = 10
	cfg.PLC.ScanOverrunWarningMS = 100
	application := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if err := application.RunRuntime(ctx); err != nil {
		t.Fatalf("RunRuntime() with unavailable broker error = %v", err)
	}
}

func TestMQTTOriginatedRuntimeEventsAreNotRepublished(t *testing.T) {
	if shouldPublishRuntimeEventToMQTT(plc.Event{}) != true {
		t.Fatal("runtime event without source should be published")
	}
	if shouldPublishRuntimeEventToMQTT(plc.Event{Details: map[string]any{"source": "mqtt"}}) {
		t.Fatal("MQTT-originated runtime event should not be republished")
	}
	if !shouldPublishRuntimeEventToMQTT(plc.Event{Details: map[string]any{"source": "local"}}) {
		t.Fatal("non-MQTT runtime event should be published")
	}
}
