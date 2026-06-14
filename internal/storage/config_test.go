package storage

import "testing"

func TestConfigValidate_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should be valid, got: %v", err)
	}
}

func TestConfigValidate_ValidSQLite(t *testing.T) {
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          "data/history.db",
		EventsJSONLPath:     "logs/events.jsonl",
		RetentionMaxSamples: 100000,
		WriteQueueSize:      256,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid sqlite config should pass, got: %v", err)
	}
}

func TestConfigValidate_EmptySQLitePath(t *testing.T) {
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          "",
		EventsJSONLPath:     "logs/events.jsonl",
		RetentionMaxSamples: 100000,
		WriteQueueSize:      256,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("empty sqlite_path should be rejected")
	}
}

func TestConfigValidate_InvalidType(t *testing.T) {
	cfg := Config{
		Enabled:             true,
		Type:                "postgres",
		SQLitePath:          "data/history.db",
		EventsJSONLPath:     "logs/events.jsonl",
		RetentionMaxSamples: 100000,
		WriteQueueSize:      256,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("non-sqlite type should be rejected")
	}
}

func TestConfigValidate_ZeroRetention(t *testing.T) {
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          "data/history.db",
		EventsJSONLPath:     "logs/events.jsonl",
		RetentionMaxSamples: 0,
		WriteQueueSize:      256,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("zero retention_max_samples should be rejected")
	}
}

func TestConfigValidate_ZeroQueueSize(t *testing.T) {
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          "data/history.db",
		EventsJSONLPath:     "logs/events.jsonl",
		RetentionMaxSamples: 100,
		WriteQueueSize:      0,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("zero write_queue_size should be rejected")
	}
}
