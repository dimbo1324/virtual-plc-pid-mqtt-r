package app

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/storage"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func TestBuildCommandPayloadJSON_MinimalCommand(t *testing.T) {
	out := buildCommandPayloadJSON(plc.Command{Command: plc.CommandStartPLC})
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["command"] != string(plc.CommandStartPLC) {
		t.Errorf("command = %v, want %q", m["command"], plc.CommandStartPLC)
	}
	if _, ok := m["loop"]; ok {
		t.Error("empty loop should not appear in payload")
	}
	if _, ok := m["command_id"]; ok {
		t.Error("empty command_id should not appear in payload")
	}
}

func TestBuildCommandPayloadJSON_AllOptionalFields(t *testing.T) {
	v := 7.5
	kp, ki, kd := 2.0, 0.2, 0.05
	mo := 50.0
	cmd := plc.Command{
		Command:      plc.CommandSetPIDGains,
		CommandID:    "cmd-42",
		Loop:         "pressure",
		Value:        &v,
		Kp:           &kp,
		Ki:           &ki,
		Kd:           &kd,
		Mode:         plc.LoopModeManual,
		ManualOutput: &mo,
	}
	out := buildCommandPayloadJSON(cmd)
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	checks := map[string]any{
		"command":    string(plc.CommandSetPIDGains),
		"command_id": "cmd-42",
		"loop":       "pressure",
	}
	for key, want := range checks {
		if m[key] != want {
			t.Errorf("payload[%q] = %v, want %v", key, m[key], want)
		}
	}
	if m["kp"] == nil || m["ki"] == nil || m["kd"] == nil {
		t.Error("kp/ki/kd should appear in payload when set")
	}
	if m["value"] == nil {
		t.Error("value should appear in payload when set")
	}
	if m["manual_output"] == nil {
		t.Error("manual_output should appear in payload when set")
	}
}

func TestPlcEventToStorageRecord_FieldMapping(t *testing.T) {
	ts := time.Now().UTC()
	details := map[string]any{"loop": "pressure", "pv": 6.2}
	event := plc.Event{
		Timestamp: ts,
		Level:     "info",
		Type:      "setpoint_changed",
		Message:   "setpoint applied",
		Details:   details,
	}
	rec := plcEventToStorageRecord(event)
	if rec.Timestamp != ts {
		t.Errorf("Timestamp = %v, want %v", rec.Timestamp, ts)
	}
	if rec.Level != "info" || rec.Type != "setpoint_changed" || rec.Message != "setpoint applied" {
		t.Fatalf("record = %+v", rec)
	}
	if rec.Details["loop"] != "pressure" {
		t.Fatalf("details = %v", rec.Details)
	}
}

func TestPlcEventToStorageRecord_NilDetails(t *testing.T) {
	rec := plcEventToStorageRecord(plc.Event{Level: "info", Type: "plc_started"})
	if rec.Details != nil {
		t.Errorf("details should be nil when event has no details, got %v", rec.Details)
	}
}

func TestMapStorageConfig_DefaultQueueSize(t *testing.T) {
	cfg := config.Default()
	cfg.Storage.WriteQueueSize = 0
	mapped := mapStorageConfig(cfg)
	if mapped.WriteQueueSize != storage.DefaultWriteQueueSize {
		t.Fatalf("WriteQueueSize = %d, want %d", mapped.WriteQueueSize, storage.DefaultWriteQueueSize)
	}
}

func TestMapStorageConfig_CustomQueueSize(t *testing.T) {
	cfg := config.Default()
	cfg.Storage.WriteQueueSize = 512
	mapped := mapStorageConfig(cfg)
	if mapped.WriteQueueSize != 512 {
		t.Fatalf("WriteQueueSize = %d, want 512", mapped.WriteQueueSize)
	}
}

func TestMapWebConfig_FieldMapping(t *testing.T) {
	cfg := config.Default()
	cfg.Web.Enabled = true
	cfg.Web.Host = "0.0.0.0"
	cfg.Web.Port = 9090
	mapped := mapWebConfig(cfg)
	if !mapped.Enabled {
		t.Error("Enabled should be true")
	}
	if mapped.Host != "0.0.0.0" || mapped.Port != 9090 {
		t.Errorf("addr = %s:%d, want 0.0.0.0:9090", mapped.Host, mapped.Port)
	}
}
