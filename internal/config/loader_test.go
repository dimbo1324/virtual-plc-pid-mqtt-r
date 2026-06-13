package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfiguration(t *testing.T) {
	payload, err := json.Marshal(Default())
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	path := writeTestConfig(t, payload)
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.App.DeviceID != Default().App.DeviceID || len(loaded.Loops) != len(Default().Loops) {
		t.Fatalf("loaded config = %+v", loaded)
	}
}

func TestLoadRejectsUnknownFieldsAndMultipleValues(t *testing.T) {
	tests := map[string][]byte{
		"unknown field":   []byte(`{"unexpected":true}`),
		"multiple values": []byte(`{} {}`),
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(writeTestConfig(t, payload)); err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}

func writeTestConfig(t *testing.T, payload []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	return path
}
