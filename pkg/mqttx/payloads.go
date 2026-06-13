package mqttx

import (
	"encoding/json"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

// StatusPayload is retained on the status topic.
type StatusPayload struct {
	DeviceID  string    `json:"device_id"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// EventPayload adds the publishing device to a PLC event.
type EventPayload struct {
	DeviceID  string         `json:"device_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Type      string         `json:"event_type"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
}

func MarshalSnapshot(snapshot plc.Snapshot) ([]byte, error) { return json.Marshal(snapshot) }
func MarshalStatus(status StatusPayload) ([]byte, error)    { return json.Marshal(status) }

func MarshalEvent(deviceID string, event plc.Event) ([]byte, error) {
	return json.Marshal(EventPayload{
		DeviceID: deviceID, Timestamp: event.Timestamp, Level: event.Level,
		Type: event.Type, Message: event.Message, Details: event.Details,
	})
}
