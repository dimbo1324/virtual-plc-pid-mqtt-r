package mqttx

import (
	"testing"
	"time"
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
	client.Disconnect(0)
}
