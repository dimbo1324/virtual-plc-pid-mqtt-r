package simulator_test

import (
	"errors"
	"math"
	"testing"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func TestNewProcessAcceptsValidConfig(t *testing.T) {
	config := validConfig()
	config.DisplayName = ""

	process, err := simulator.NewProcess(config)
	if err != nil {
		t.Fatalf("NewProcess() error = %v", err)
	}
	if process == nil {
		t.Fatal("NewProcess() process = nil")
	}
	if got := process.Config().DisplayName; got != config.Name {
		t.Fatalf("Config().DisplayName = %q, want %q", got, config.Name)
	}
}

func TestNewProcessRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*simulator.Config)
	}{
		{
			name: "empty name",
			mutate: func(config *simulator.Config) {
				config.Name = " "
			},
		},
		{
			name: "equal limits",
			mutate: func(config *simulator.Config) {
				config.Min = config.Max
			},
		},
		{
			name: "initial PV below minimum",
			mutate: func(config *simulator.Config) {
				config.InitialPV = config.Min - 1
			},
		},
		{
			name: "initial PV above maximum",
			mutate: func(config *simulator.Config) {
				config.InitialPV = config.Max + 1
			},
		},
		{
			name: "zero tau",
			mutate: func(config *simulator.Config) {
				config.TauSeconds = 0
			},
		},
		{
			name: "negative noise",
			mutate: func(config *simulator.Config) {
				config.NoiseStddev = -0.1
			},
		},
		{
			name: "NaN base",
			mutate: func(config *simulator.Config) {
				config.Base = math.NaN()
			},
		},
		{
			name: "infinite gain",
			mutate: func(config *simulator.Config) {
				config.Gain = math.Inf(1)
			},
		},
		{
			name: "infinite initial PV",
			mutate: func(config *simulator.Config) {
				config.InitialPV = math.Inf(-1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.mutate(&config)

			_, err := simulator.NewProcess(config)
			if !errors.Is(err, simulator.ErrInvalidConfig) {
				t.Fatalf("NewProcess() error = %v, want %v", err, simulator.ErrInvalidConfig)
			}
		})
	}
}

func TestDefaultConfigsValidate(t *testing.T) {
	tests := []struct {
		name   string
		config simulator.Config
	}{
		{name: "pressure", config: simulator.DefaultPressureConfig()},
		{name: "temperature", config: simulator.DefaultTemperatureConfig()},
		{name: "level", config: simulator.DefaultLevelConfig()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); err != nil {
				t.Fatalf("Config.Validate() error = %v", err)
			}
		})
	}
}

func validConfig() simulator.Config {
	return simulator.Config{
		Name:        "test_process",
		DisplayName: "Test Process",
		Unit:        "unit",
		InitialPV:   10,
		Min:         0,
		Max:         100,
		Base:        0,
		Gain:        1,
		TauSeconds:  10,
		NoiseStddev: 0,
		RandomSeed:  42,
	}
}

func mustProcess(t *testing.T, config simulator.Config) *simulator.Process {
	t.Helper()
	process, err := simulator.NewProcess(config)
	if err != nil {
		t.Fatalf("NewProcess() error = %v", err)
	}
	return process
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
