package config

import "testing"

func TestDefaultIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Default().Validate() error = %v", err)
	}
	if cfg.Web.Enabled || cfg.Storage.Enabled {
		t.Fatal("default config enables web or storage before those subsystems are implemented")
	}
}

func TestDisabledMQTTAllowsEmptyConnectionSettings(t *testing.T) {
	cfg := Default()
	cfg.MQTT.Enabled = false
	cfg.MQTT.BrokerURL = ""
	cfg.MQTT.ClientID = ""
	cfg.MQTT.BaseTopic = ""
	cfg.MQTT.ConnectTimeoutSeconds = 0
	cfg.MQTT.ReconnectIntervalSeconds = 0
	if err := cfg.Validate(); err != nil {
		t.Fatalf("disabled MQTT config rejected: %v", err)
	}
}

func TestValidationRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{
			name: "empty device ID",
			mutate: func(cfg *Config) {
				cfg.App.DeviceID = ""
			},
		},
		{
			name: "zero scan interval",
			mutate: func(cfg *Config) {
				cfg.PLC.ScanIntervalMS = 0
			},
		},
		{
			name: "zero scan overrun warning",
			mutate: func(cfg *Config) {
				cfg.PLC.ScanOverrunWarningMS = 0
			},
		},
		{
			name: "invalid enabled web port",
			mutate: func(cfg *Config) {
				cfg.Web.Enabled = true
				cfg.Web.Port = 70000
			},
		},
		{
			name: "enabled MQTT without broker URL",
			mutate: func(cfg *Config) {
				cfg.MQTT.Enabled = true
				cfg.MQTT.BrokerURL = ""
			},
		},
		{
			name: "enabled MQTT with bad QoS",
			mutate: func(cfg *Config) {
				cfg.MQTT.QoS = 3
			},
		},
		{
			name: "enabled MQTT with wildcard base topic",
			mutate: func(cfg *Config) {
				cfg.MQTT.BaseTopic = "vplc/+"
			},
		},
		{
			name: "enabled MQTT with empty topic level",
			mutate: func(cfg *Config) {
				cfg.MQTT.BaseTopic = "vplc//device"
			},
		},
		{
			name: "enabled storage without SQLite path",
			mutate: func(cfg *Config) {
				cfg.Storage.Enabled = true
				cfg.Storage.SQLitePath = ""
			},
		},
		{
			name: "enabled storage without events path",
			mutate: func(cfg *Config) {
				cfg.Storage.Enabled = true
				cfg.Storage.EventsJSONLPath = ""
			},
		},
		{
			name: "enabled storage without app log path",
			mutate: func(cfg *Config) {
				cfg.Storage.Enabled = true
				cfg.Storage.AppLogPath = ""
			},
		},
		{
			name: "loop name with surrounding whitespace",
			mutate: func(cfg *Config) {
				cfg.Loops[0].Name = " pressure "
			},
		},
		{
			name: "duplicate loop names",
			mutate: func(cfg *Config) {
				duplicate := cfg.Loops[0]
				cfg.Loops = append(cfg.Loops, duplicate)
			},
		},
		{
			name: "setpoint outside range",
			mutate: func(cfg *Config) {
				cfg.Loops[0].Setpoint = cfg.Loops[0].SetpointMax + 1
			},
		},
		{
			name: "invalid PID output range",
			mutate: func(cfg *Config) {
				cfg.Loops[0].PID.OutputMin = cfg.Loops[0].PID.OutputMax
			},
		},
		{
			name: "non-positive process tau",
			mutate: func(cfg *Config) {
				cfg.Loops[0].Process.TauSeconds = 0
			},
		},
		{
			name: "unknown loop mode",
			mutate: func(cfg *Config) {
				cfg.Loops[0].Mode = "cascade"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.mutate(&cfg)

			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want validation error")
			}
		})
	}
}

func TestSupportedLoopModesMatchRuntimeModes(t *testing.T) {
	for _, mode := range []string{"auto", "manual", "hold", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			cfg := Default()
			cfg.Loops[0].Mode = mode
			if err := cfg.Validate(); err != nil {
				t.Fatalf("mode %q rejected: %v", mode, err)
			}
		})
	}
}
