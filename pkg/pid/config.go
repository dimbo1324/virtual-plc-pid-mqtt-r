package pid

import (
	"fmt"
	"math"
	"strings"
)

// Config defines the controller tuning, limits, and initial operating mode.
type Config struct {
	Name      string
	Kp        float64
	Ki        float64
	Kd        float64
	Bias      float64
	OutputMin float64
	OutputMax float64
	Setpoint  float64
	Mode      Mode
	Enabled   bool
}

// Validate checks whether the configuration can safely initialize a controller.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidConfig)
	}
	if !finite(c.Kp) || !finite(c.Ki) || !finite(c.Kd) || c.Kp < 0 || c.Ki < 0 || c.Kd < 0 {
		return fmt.Errorf("%w: %w", ErrInvalidConfig, ErrInvalidTunings)
	}
	if !finite(c.Bias) {
		return fmt.Errorf("%w: bias must be finite", ErrInvalidConfig)
	}
	if !finite(c.Setpoint) {
		return fmt.Errorf("%w: %w", ErrInvalidConfig, ErrNonFiniteSetpoint)
	}
	if !finite(c.OutputMin) || !finite(c.OutputMax) {
		return fmt.Errorf("%w: output limits must be finite", ErrInvalidConfig)
	}
	if c.OutputMin >= c.OutputMax {
		return fmt.Errorf("%w: output minimum must be less than output maximum", ErrInvalidConfig)
	}
	if !c.Mode.Valid() {
		return fmt.Errorf("%w: %w %q", ErrInvalidConfig, ErrInvalidMode, c.Mode)
	}

	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
