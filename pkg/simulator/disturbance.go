package simulator

import (
	"fmt"
	"strings"
)

// Disturbance temporarily offsets the process target.
type Disturbance struct {
	Name             string
	Amplitude        float64
	RemainingSeconds float64
}

func (d Disturbance) validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidDisturbance)
	}
	if !finite(d.Amplitude) {
		return fmt.Errorf("%w: amplitude must be finite", ErrInvalidDisturbance)
	}
	if !finite(d.RemainingSeconds) || d.RemainingSeconds <= 0 {
		return fmt.Errorf("%w: remaining seconds must be finite and greater than zero", ErrInvalidDisturbance)
	}
	return nil
}
