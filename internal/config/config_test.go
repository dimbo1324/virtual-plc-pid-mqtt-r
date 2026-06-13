package config

import "testing"

func TestDefaultIsValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("Default().Validate() error = %v", err)
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
