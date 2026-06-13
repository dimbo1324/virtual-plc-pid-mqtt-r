package simulator

import (
	"fmt"
	"math"
	"strings"
)

// Config defines one independent first-order process model.
type Config struct {
	Name               string
	DisplayName        string
	Unit               string
	InitialPV          float64
	Min                float64
	Max                float64
	Base               float64
	Gain               float64
	TauSeconds         float64
	NoiseStddev        float64
	RandomSeed         int64
	RandomDisturbances bool
}

// Validate checks whether the configuration can initialize a safe process.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidConfig)
	}
	if !finite(c.InitialPV) || !finite(c.Min) || !finite(c.Max) ||
		!finite(c.Base) || !finite(c.Gain) || !finite(c.TauSeconds) || !finite(c.NoiseStddev) {
		return fmt.Errorf("%w: numeric values must be finite", ErrInvalidConfig)
	}
	if c.Min >= c.Max {
		return fmt.Errorf("%w: minimum must be less than maximum", ErrInvalidConfig)
	}
	if c.InitialPV < c.Min || c.InitialPV > c.Max {
		return fmt.Errorf("%w: initial process value must be within [%g, %g]", ErrInvalidConfig, c.Min, c.Max)
	}
	if c.TauSeconds <= 0 {
		return fmt.Errorf("%w: tau seconds must be greater than zero", ErrInvalidConfig)
	}
	if c.NoiseStddev < 0 {
		return fmt.Errorf("%w: noise standard deviation must not be negative", ErrInvalidConfig)
	}
	return nil
}

func (c Config) normalized() Config {
	c.Name = strings.TrimSpace(c.Name)
	if strings.TrimSpace(c.DisplayName) == "" {
		c.DisplayName = c.Name
	}
	return c
}

// DefaultPressureConfig returns the baseline pressure process from the project specification.
func DefaultPressureConfig() Config {
	return Config{
		Name:               "pressure",
		DisplayName:        "Pressure",
		Unit:               "bar",
		InitialPV:          4,
		Min:                0,
		Max:                12,
		Base:               0,
		Gain:               0.10,
		TauSeconds:         15,
		NoiseStddev:        0.03,
		RandomSeed:         1,
		RandomDisturbances: true,
	}
}

// DefaultTemperatureConfig returns the baseline temperature process from the project specification.
func DefaultTemperatureConfig() Config {
	return Config{
		Name:               "temperature",
		DisplayName:        "Temperature",
		Unit:               "C",
		InitialPV:          80,
		Min:                0,
		Max:                250,
		Base:               20,
		Gain:               2,
		TauSeconds:         60,
		NoiseStddev:        0.2,
		RandomSeed:         2,
		RandomDisturbances: true,
	}
}

// DefaultLevelConfig returns the baseline level process from the project specification.
func DefaultLevelConfig() Config {
	return Config{
		Name:               "level",
		DisplayName:        "Level",
		Unit:               "%",
		InitialPV:          45,
		Min:                0,
		Max:                100,
		Base:               0,
		Gain:               1,
		TauSeconds:         25,
		NoiseStddev:        0.1,
		RandomSeed:         3,
		RandomDisturbances: true,
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func clamp(value, minimum, maximum float64) (float64, bool) {
	if value < minimum {
		return minimum, true
	}
	if value > maximum {
		return maximum, true
	}
	return value, false
}
