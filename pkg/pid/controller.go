package pid

import (
	"fmt"
	"time"
)

// Controller calculates a bounded PID output from a process value.
type Controller struct {
	config          Config
	state           State
	manualOutput    float64
	hasProcessValue bool
	bumplessPending bool
}

// New validates config and creates a controller in its configured mode.
func New(config Config) (*Controller, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	mode := config.Mode
	if !config.Enabled {
		mode = ModeDisabled
	}

	initialOutput := config.OutputMin
	controller := &Controller{
		config:       config,
		manualOutput: initialOutput,
		state: State{
			Setpoint: config.Setpoint,
			Output:   initialOutput,
			Mode:     mode,
			Enabled:  config.Enabled,
		},
	}

	return controller, nil
}

// Update calculates the next output for pv over dt.
func (c *Controller) Update(pv float64, dt time.Duration) (float64, error) {
	if dt <= 0 {
		return c.state.Output, ErrInvalidDuration
	}
	if !finite(pv) {
		return c.state.Output, ErrNonFiniteProcessValue
	}
	if !finite(c.config.Setpoint) {
		return c.state.Output, ErrNonFiniteSetpoint
	}

	errorValue := c.config.Setpoint - pv
	if !finite(errorValue) {
		return c.state.Output, ErrNonFiniteCalculation
	}

	next := c.state
	next.Setpoint = c.config.Setpoint
	next.LastError = c.state.Error
	next.LastPV = c.state.ProcessValue
	next.ProcessValue = pv
	next.Error = errorValue
	next.Mode = c.effectiveMode()
	next.Enabled = c.config.Enabled

	switch next.Mode {
	case ModeAuto:
		if err := c.updateAuto(&next, dt); err != nil {
			return c.state.Output, err
		}
	case ModeManual:
		next.Proportional = c.config.Kp * errorValue
		next.Derivative = 0
		next.Output = clamp(c.manualOutput, c.config.OutputMin, c.config.OutputMax)
		if !finite(next.Proportional) {
			return c.state.Output, ErrNonFiniteCalculation
		}
	case ModeHold:
		next.Proportional = c.config.Kp * errorValue
		next.Derivative = 0
		next.Output = clamp(c.state.Output, c.config.OutputMin, c.config.OutputMax)
		if !finite(next.Proportional) {
			return c.state.Output, ErrNonFiniteCalculation
		}
	case ModeDisabled:
		next.Proportional = 0
		next.Derivative = 0
		next.Output = c.config.OutputMin
	default:
		return c.state.Output, fmt.Errorf("%w %q", ErrInvalidMode, next.Mode)
	}

	if !finite(next.Output) || !finite(next.Integral) || !finite(next.Derivative) {
		return c.state.Output, ErrNonFiniteCalculation
	}

	c.state = next
	c.hasProcessValue = true
	c.bumplessPending = false

	return next.Output, nil
}

func (c *Controller) updateAuto(next *State, dt time.Duration) error {
	seconds := dt.Seconds()
	proportional := c.config.Kp * next.Error
	derivative := 0.0
	if c.hasProcessValue && !c.bumplessPending {
		// Derivative-on-measurement avoids a derivative kick when only SP changes.
		derivative = -c.config.Kd * (next.ProcessValue - c.state.ProcessValue) / seconds
	}
	if !finite(proportional) || !finite(derivative) {
		return ErrNonFiniteCalculation
	}

	candidateIntegral := c.state.Integral
	if !c.bumplessPending {
		candidateIntegral += c.config.Ki * next.Error * seconds
	}
	if !finite(candidateIntegral) {
		return ErrNonFiniteCalculation
	}

	candidateOutput := c.config.Bias + proportional + candidateIntegral + derivative
	if !finite(candidateOutput) {
		return ErrNonFiniteCalculation
	}

	integral := c.state.Integral
	// Conditional integration prevents the integral from pushing the output
	// farther into saturation, while still allowing it to unwind toward range.
	if candidateOutput >= c.config.OutputMin && candidateOutput <= c.config.OutputMax ||
		candidateOutput > c.config.OutputMax && next.Error < 0 ||
		candidateOutput < c.config.OutputMin && next.Error > 0 {
		integral = candidateIntegral
	}

	unclampedOutput := c.config.Bias + proportional + integral + derivative
	if !finite(unclampedOutput) {
		return ErrNonFiniteCalculation
	}

	next.Proportional = proportional
	next.Integral = integral
	next.Derivative = derivative
	next.Output = clamp(unclampedOutput, c.config.OutputMin, c.config.OutputMax)

	return nil
}

// State returns a copy of the current controller state.
func (c *Controller) State() State {
	return c.state
}

// Config returns a copy of the current controller configuration.
func (c *Controller) Config() Config {
	return c.config
}

// SetSetpoint changes the target used by subsequent updates.
// Update rejects a non-finite setpoint without changing runtime state.
func (c *Controller) SetSetpoint(setpoint float64) {
	c.config.Setpoint = setpoint
	if finite(setpoint) {
		c.state.Setpoint = setpoint
	}
}

// SetTunings changes the PID gains used by subsequent updates.
func (c *Controller) SetTunings(kp, ki, kd float64) error {
	if !finite(kp) || !finite(ki) || !finite(kd) || kp < 0 || ki < 0 || kd < 0 {
		return ErrInvalidTunings
	}

	c.config.Kp = kp
	c.config.Ki = ki
	c.config.Kd = kd
	return nil
}

// SetMode changes the operating mode.
func (c *Controller) SetMode(mode Mode) error {
	if !mode.Valid() {
		return fmt.Errorf("%w %q", ErrInvalidMode, mode)
	}

	previousMode := c.effectiveMode()
	c.config.Mode = mode
	effectiveMode := c.effectiveMode()
	c.state.Mode = effectiveMode

	switch effectiveMode {
	case ModeManual:
		c.manualOutput = clamp(c.state.Output, c.config.OutputMin, c.config.OutputMax)
		c.state.Derivative = 0
	case ModeAuto:
		if previousMode != ModeAuto {
			// Track the current output through the integral term. The first auto
			// update then skips I and D changes for a simple bumpless transfer.
			trackingIntegral := c.state.Output - c.config.Bias - c.config.Kp*c.state.Error
			if finite(trackingIntegral) {
				c.state.Integral = trackingIntegral
			} else {
				c.state.Integral = 0
			}
			c.state.Derivative = 0
			c.bumplessPending = true
		}
	case ModeDisabled:
		c.state.Proportional = 0
		c.state.Derivative = 0
		c.state.Output = c.config.OutputMin
	case ModeHold:
		c.state.Derivative = 0
	}

	return nil
}

// SetManualOutput sets and clamps the output used in manual mode.
func (c *Controller) SetManualOutput(output float64) error {
	if !finite(output) {
		return ErrNonFiniteManualOutput
	}

	c.manualOutput = clamp(output, c.config.OutputMin, c.config.OutputMax)
	if c.effectiveMode() == ModeManual {
		c.state.Output = c.manualOutput
	}
	return nil
}

// Reset clears dynamic PID state while preserving configuration and mode.
func (c *Controller) Reset() {
	mode := c.effectiveMode()
	c.state = State{
		Setpoint: c.config.Setpoint,
		Output:   c.config.OutputMin,
		Mode:     mode,
		Enabled:  c.config.Enabled,
	}
	c.manualOutput = c.config.OutputMin
	c.hasProcessValue = false
	c.bumplessPending = false
}

func (c *Controller) effectiveMode() Mode {
	if !c.config.Enabled {
		return ModeDisabled
	}
	return c.config.Mode
}
