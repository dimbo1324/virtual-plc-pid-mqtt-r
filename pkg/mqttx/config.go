package mqttx

import (
	"fmt"
	"strings"
	"time"
)

// Config defines one MQTT client connection and topic namespace.
type Config struct {
	Enabled           bool
	BrokerURL         string
	ClientID          string
	Username          string
	Password          string
	BaseTopic         string
	QoS               byte
	ConnectTimeout    time.Duration
	ReconnectInterval time.Duration
}

// Validate checks connection settings without exposing credentials.
func (c Config) Validate() error {
	if c.QoS > 2 {
		return fmt.Errorf("%w: QoS must be 0, 1, or 2", ErrInvalidConfig)
	}
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.BrokerURL) == "" {
		return fmt.Errorf("%w: broker URL must not be empty", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.ClientID) == "" {
		return fmt.Errorf("%w: client ID must not be empty", ErrInvalidConfig)
	}
	baseTopic := strings.Trim(strings.TrimSpace(c.BaseTopic), "/")
	if baseTopic == "" {
		return fmt.Errorf("%w: base topic must not be empty", ErrInvalidConfig)
	}
	if strings.ContainsAny(baseTopic, "+#\x00") {
		return fmt.Errorf("%w: base topic must not contain MQTT wildcards or null bytes", ErrInvalidConfig)
	}
	for _, level := range strings.Split(baseTopic, "/") {
		if level == "" {
			return fmt.Errorf("%w: base topic must not contain empty topic levels", ErrInvalidConfig)
		}
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("%w: connect timeout must be greater than zero", ErrInvalidConfig)
	}
	if c.ReconnectInterval <= 0 {
		return fmt.Errorf("%w: reconnect interval must be greater than zero", ErrInvalidConfig)
	}
	return nil
}

func (c Config) normalized() Config {
	c.BrokerURL = strings.TrimSpace(c.BrokerURL)
	c.ClientID = strings.TrimSpace(c.ClientID)
	c.BaseTopic = strings.Trim(strings.TrimSpace(c.BaseTopic), "/")
	return c
}
