package mqttx

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func TestPayloadsMarshalValidJSON(t *testing.T) {
	now := time.Now().UTC()
	payloads := []struct {
		name    string
		marshal func() ([]byte, error)
	}{
		{"snapshot", func() ([]byte, error) {
			return MarshalSnapshot(plc.Snapshot{Timestamp: now, DeviceID: "device", Loops: map[string]plc.LoopSnapshot{"pressure": {Name: "pressure", ProcessValue: 5}}})
		}},
		{"status", func() ([]byte, error) {
			return MarshalStatus(StatusPayload{DeviceID: "device", Status: "online", Timestamp: now})
		}},
		{"event", func() ([]byte, error) {
			return MarshalEvent("device", plc.Event{Timestamp: now, Level: "info", Type: "test", Message: "ok"})
		}},
	}
	for _, tt := range payloads {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := tt.marshal()
			if err != nil {
				t.Fatalf("marshal error = %v", err)
			}
			if !json.Valid(payload) {
				t.Fatalf("invalid JSON: %s", payload)
			}
		})
	}
}
