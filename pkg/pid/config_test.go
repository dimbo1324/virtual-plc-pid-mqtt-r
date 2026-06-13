package pid_test

import (
	"errors"
	"math"
	"testing"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
)

func TestNewAcceptsValidConfig(t *testing.T) {
	controller, err := pid.New(validConfig())
	if err != nil {
		t.Fatalf("pid.New() error = %v", err)
	}
	if controller == nil {
		t.Fatal("pid.New() controller = nil")
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*pid.Config)
		want   error
	}{
		{
			name: "empty name",
			mutate: func(config *pid.Config) {
				config.Name = " "
			},
			want: pid.ErrInvalidConfig,
		},
		{
			name: "equal output limits",
			mutate: func(config *pid.Config) {
				config.OutputMin = config.OutputMax
			},
			want: pid.ErrInvalidConfig,
		},
		{
			name: "negative proportional gain",
			mutate: func(config *pid.Config) {
				config.Kp = -1
			},
			want: pid.ErrInvalidTunings,
		},
		{
			name: "non-finite integral gain",
			mutate: func(config *pid.Config) {
				config.Ki = math.Inf(1)
			},
			want: pid.ErrInvalidTunings,
		},
		{
			name: "non-finite bias",
			mutate: func(config *pid.Config) {
				config.Bias = math.NaN()
			},
			want: pid.ErrInvalidConfig,
		},
		{
			name: "non-finite setpoint",
			mutate: func(config *pid.Config) {
				config.Setpoint = math.Inf(-1)
			},
			want: pid.ErrNonFiniteSetpoint,
		},
		{
			name: "non-finite output limit",
			mutate: func(config *pid.Config) {
				config.OutputMax = math.Inf(1)
			},
			want: pid.ErrInvalidConfig,
		},
		{
			name: "unknown mode",
			mutate: func(config *pid.Config) {
				config.Mode = pid.Mode("cascade")
			},
			want: pid.ErrInvalidMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.mutate(&config)

			_, err := pid.New(config)
			if !errors.Is(err, tt.want) {
				t.Fatalf("pid.New() error = %v, want error matching %v", err, tt.want)
			}
		})
	}
}

func TestSetTuningsRejectsInvalidValues(t *testing.T) {
	controller := mustController(t, validConfig())

	tests := []struct {
		name       string
		kp, ki, kd float64
	}{
		{name: "negative", kp: -1, ki: 0, kd: 0},
		{name: "NaN", kp: math.NaN(), ki: 0, kd: 0},
		{name: "infinity", kp: 1, ki: math.Inf(1), kd: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := controller.SetTunings(tt.kp, tt.ki, tt.kd); !errors.Is(err, pid.ErrInvalidTunings) {
				t.Fatalf("SetTunings() error = %v, want %v", err, pid.ErrInvalidTunings)
			}
		})
	}
}

func validConfig() pid.Config {
	return pid.Config{
		Name:      "pressure",
		Kp:        2,
		Ki:        1,
		Kd:        0.5,
		Bias:      0,
		OutputMin: -100,
		OutputMax: 100,
		Setpoint:  10,
		Mode:      pid.ModeAuto,
		Enabled:   true,
	}
}

func mustController(t *testing.T, config pid.Config) *pid.Controller {
	t.Helper()
	controller, err := pid.New(config)
	if err != nil {
		t.Fatalf("pid.New() error = %v", err)
	}
	return controller
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
