package plc

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

// Config defines one virtual PLC runtime.
type Config struct {
	DeviceID           string
	ScanInterval       time.Duration
	PublishInterval    time.Duration
	UIUpdateInterval   time.Duration
	ScanOverrunWarning time.Duration
	Loops              []LoopConfig
}

// LoopConfig connects one PID controller to one synthetic process.
type LoopConfig struct {
	Name        string
	DisplayName string
	Unit        string
	Enabled     bool
	Mode        LoopMode
	Setpoint    float64
	SetpointMin float64
	SetpointMax float64
	PID         pid.Config
	Process     simulator.Config
}

func (c Config) validate() error {
	if strings.TrimSpace(c.DeviceID) == "" {
		return fmt.Errorf("%w: device ID must not be empty", ErrInvalidConfig)
	}
	if c.ScanInterval <= 0 {
		return fmt.Errorf("%w: scan interval must be greater than zero", ErrInvalidConfig)
	}
	if c.PublishInterval <= 0 {
		return fmt.Errorf("%w: publish interval must be greater than zero", ErrInvalidConfig)
	}
	if c.UIUpdateInterval <= 0 {
		return fmt.Errorf("%w: UI update interval must be greater than zero", ErrInvalidConfig)
	}
	if c.ScanOverrunWarning <= 0 {
		return fmt.Errorf("%w: scan overrun warning must be greater than zero", ErrInvalidConfig)
	}
	if len(c.Loops) == 0 {
		return fmt.Errorf("%w: at least one loop is required", ErrInvalidConfig)
	}

	names := make(map[string]struct{}, len(c.Loops))
	for i, loop := range c.Loops {
		name := strings.TrimSpace(loop.Name)
		if name == "" {
			return fmt.Errorf("%w: loops[%d] name must not be empty", ErrInvalidConfig, i)
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("%w: duplicate loop name %q", ErrInvalidConfig, name)
		}
		names[name] = struct{}{}
		if !loop.Mode.Valid() {
			return fmt.Errorf("%w: loop %q has invalid mode %q", ErrInvalidConfig, name, loop.Mode)
		}
		if !finite(loop.Setpoint) {
			return fmt.Errorf("%w: loop %q setpoint must be finite", ErrInvalidConfig, name)
		}
		if loop.hasSetpointLimits() {
			if !finite(loop.SetpointMin) || !finite(loop.SetpointMax) || loop.SetpointMin >= loop.SetpointMax {
				return fmt.Errorf("%w: loop %q has invalid setpoint limits", ErrInvalidConfig, name)
			}
			if loop.Setpoint < loop.SetpointMin || loop.Setpoint > loop.SetpointMax {
				return fmt.Errorf("%w: loop %q setpoint is outside [%g, %g]", ErrInvalidConfig, name, loop.SetpointMin, loop.SetpointMax)
			}
		}

		pidConfig := loop.PID
		pidConfig.Name = name
		pidConfig.Setpoint = loop.Setpoint
		pidConfig.Mode = loop.Mode.pidMode()
		pidConfig.Enabled = loop.Enabled
		if err := pidConfig.Validate(); err != nil {
			return fmt.Errorf("%w: loop %q PID: %v", ErrInvalidConfig, name, err)
		}

		processConfig := loop.Process
		processConfig.Name = name
		if processConfig.DisplayName == "" {
			processConfig.DisplayName = loop.DisplayName
		}
		if processConfig.Unit == "" {
			processConfig.Unit = loop.Unit
		}
		if err := processConfig.Validate(); err != nil {
			return fmt.Errorf("%w: loop %q process: %v", ErrInvalidConfig, name, err)
		}
	}
	return nil
}

func (c LoopConfig) hasSetpointLimits() bool {
	return c.SetpointMin != 0 || c.SetpointMax != 0
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
