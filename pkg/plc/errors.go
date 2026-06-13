package plc

import "errors"

var (
	ErrInvalidConfig  = errors.New("invalid PLC configuration")
	ErrInvalidCommand = errors.New("invalid PLC command")
	ErrUnknownLoop    = errors.New("unknown PLC loop")
	ErrInvalidState   = errors.New("invalid PLC runtime state")
)
