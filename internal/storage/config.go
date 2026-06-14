package storage

import "fmt"

// Config holds all parameters needed to open and operate storage.
type Config struct {
	Enabled             bool
	Type                string
	SQLitePath          string
	EventsJSONLPath     string
	AppLogPath          string
	RetentionMaxSamples int
	WriteQueueSize      int
	FallbackOnError     bool
	FallbackType        string // "jsonl" or "noop"
}

// DefaultWriteQueueSize is used when WriteQueueSize is not set in the config.
const DefaultWriteQueueSize = 256

// Validate checks the configuration for consistency.
// A disabled config is always valid. An enabled config must be complete.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Type != "sqlite" && c.Type != "jsonl" {
		return fmt.Errorf("storage type %q is not supported; only \"sqlite\" is allowed", c.Type)
	}
	if c.SQLitePath == "" {
		return fmt.Errorf("storage sqlite_path must not be empty when storage is enabled")
	}
	if c.EventsJSONLPath == "" {
		return fmt.Errorf("storage events_jsonl_path must not be empty when storage is enabled")
	}
	if c.RetentionMaxSamples <= 0 {
		return fmt.Errorf("storage retention_max_samples must be > 0, got %d", c.RetentionMaxSamples)
	}
	if c.WriteQueueSize <= 0 {
		return fmt.Errorf("storage write_queue_size must be > 0, got %d", c.WriteQueueSize)
	}
	return nil
}
