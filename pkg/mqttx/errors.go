package mqttx

import "errors"

var (
	ErrInvalidConfig  = errors.New("invalid MQTT configuration")
	ErrInvalidCommand = errors.New("invalid MQTT command")
	ErrNotConnected   = errors.New("MQTT client is not connected")
	ErrTimeout        = errors.New("MQTT operation timed out")
)
