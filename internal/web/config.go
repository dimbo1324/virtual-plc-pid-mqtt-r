package web

import "fmt"

// Config holds the web server configuration.
type Config struct {
	Enabled bool
	Host    string
	Port    int
}

// Validate returns an error if the config is invalid when enabled.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Host == "" {
		return fmt.Errorf("web.host must not be empty when web is enabled")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("web.port must be between 1 and 65535 when web is enabled")
	}
	return nil
}

func (c Config) addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
