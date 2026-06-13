package simulator

import "errors"

var (
	ErrInvalidConfig        = errors.New("invalid simulator configuration")
	ErrInvalidMV            = errors.New("manipulated value must be finite")
	ErrInvalidDuration      = errors.New("simulation step duration must be greater than zero")
	ErrInvalidDisturbance   = errors.New("invalid disturbance")
	ErrInvalidState         = errors.New("simulator state must be finite")
	ErrNonFiniteCalculation = errors.New("simulation calculation produced a non-finite value")
)
