package pid

import "errors"

var (
	ErrInvalidConfig         = errors.New("invalid PID configuration")
	ErrInvalidMode           = errors.New("invalid PID mode")
	ErrInvalidDuration       = errors.New("PID update duration must be greater than zero")
	ErrNonFiniteProcessValue = errors.New("process value must be finite")
	ErrNonFiniteSetpoint     = errors.New("setpoint must be finite")
	ErrInvalidTunings        = errors.New("PID tunings must be finite and non-negative")
	ErrNonFiniteManualOutput = errors.New("manual output must be finite")
	ErrNonFiniteCalculation  = errors.New("PID calculation produced a non-finite value")
)
