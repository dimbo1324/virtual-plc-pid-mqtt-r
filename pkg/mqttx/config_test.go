package mqttx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func validConfig() Config {
	return Config{Enabled: true, BrokerURL: "tcp://localhost:1883", ClientID: "test", BaseTopic: "vplc/test", QoS: 1, ConnectTimeout: time.Second, ReconnectInterval: time.Second}
}

func TestConfigValidation(t *testing.T) {
	if err := (Config{}).Validate(); err != nil {
		t.Fatalf("disabled config rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"empty broker", func(c *Config) { c.BrokerURL = "" }},
		{"empty client ID", func(c *Config) { c.ClientID = "" }},
		{"empty base topic", func(c *Config) { c.BaseTopic = "" }},
		{"wildcard base topic", func(c *Config) { c.BaseTopic = "vplc/#" }},
		{"empty base topic level", func(c *Config) { c.BaseTopic = "vplc//test" }},
		{"bad QoS", func(c *Config) { c.QoS = 3 }},
		{"zero connect timeout", func(c *Config) { c.ConnectTimeout = 0 }},
		{"zero reconnect interval", func(c *Config) { c.ReconnectInterval = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.mutate(&config)
			if err := config.Validate(); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func TestDisabledClientIsNoOp(t *testing.T) {
	client, err := New(Config{}, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := client.Connect(nil); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if client.IsConnected() {
		t.Fatal("disabled client reports connected")
	}
	if err := client.PublishStatus(context.Background(), StatusPayload{}); err != nil {
		t.Fatalf("PublishStatus() error = %v", err)
	}
	if err := client.PublishSnapshot(context.Background(), plc.Snapshot{}); err != nil {
		t.Fatalf("PublishSnapshot() error = %v", err)
	}
	if err := client.PublishEvent(context.Background(), plc.Event{}); err != nil {
		t.Fatalf("PublishEvent() error = %v", err)
	}
	client.Disconnect(0)
}

func TestNewNormalizesConfigWithoutConnecting(t *testing.T) {
	config := validConfig()
	config.BrokerURL = " tcp://localhost:1883 "
	config.ClientID = " test-client "
	config.BaseTopic = " /vplc/device-1/ "
	client, err := New(config, func(context.Context, plc.Command) (plc.Event, error) {
		return plc.Event{}, nil
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client.config.BrokerURL != "tcp://localhost:1883" || client.config.ClientID != "test-client" || client.config.BaseTopic != "vplc/device-1" {
		t.Fatalf("normalized config = %+v", client.config)
	}
	if client.deviceID != "device-1" {
		t.Fatalf("device ID = %q", client.deviceID)
	}
	if err := client.PublishEvent(context.Background(), plc.Event{}); !errors.Is(err, ErrNotConnected) {
		t.Fatalf("PublishEvent() error = %v, want ErrNotConnected", err)
	}
}

func TestNewEnabledClientRequiresCommandHandler(t *testing.T) {
	if _, err := New(validConfig(), nil); err == nil {
		t.Fatal("New() accepted enabled config without command handler")
	}
}
