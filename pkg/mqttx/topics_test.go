package mqttx

import "testing"

func TestTopics(t *testing.T) {
	base := "/vplc/device-1/"
	tests := map[string]string{
		StatusTopic(base): "vplc/device-1/status", TelemetryTopic(base): "vplc/device-1/telemetry",
		EventsTopic(base): "vplc/device-1/events", CommandsTopic(base): "vplc/device-1/commands",
		ConfigTopic(base): "vplc/device-1/config",
	}
	for got, want := range tests {
		if got != want {
			t.Fatalf("topic = %q, want %q", got, want)
		}
	}
}
