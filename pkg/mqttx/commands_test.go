package mqttx

import (
	"testing"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name, payload string
		command       plc.CommandType
		loop          string
	}{
		{"setpoint", `{"command_id":"1","command":"set_setpoint","loop":"pressure","value":7.5}`, plc.CommandSetSetpoint, "pressure"},
		{"gains", `{"command_id":"2","command":"set_pid_gains","loop":"pressure","kp":2,"ki":0.4,"kd":0.05}`, plc.CommandSetPIDGains, "pressure"},
		{"mode", `{"command_id":"3","command":"set_mode","loop":"temperature","mode":"manual","manual_output":40}`, plc.CommandSetMode, "temperature"},
		{"start", `{"command_id":"4","command":"start_plc"}`, plc.CommandStartPLC, ""},
		{"stop", `{"command_id":"5","command":"stop_plc"}`, plc.CommandStopPLC, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := ParseCommand([]byte(tt.payload))
			if err != nil {
				t.Fatalf("ParseCommand() error = %v", err)
			}
			if command.Command != tt.command || command.Loop != tt.loop || command.Source != "mqtt" || command.ReceivedAt.IsZero() {
				t.Fatalf("command = %+v", command)
			}
		})
	}
}

func TestParseCommandRejectsInvalidPayloads(t *testing.T) {
	values := []string{
		`{`,
		`{"command":"unknown"}`,
		`{"command":"set_setpoint","loop":"pressure"}`,
		`{"command":"set_pid_gains","loop":"pressure","kp":1,"ki":2}`,
		`{"command":"set_mode","loop":"pressure","mode":"cascade"}`,
		`{"command":"start_plc","unexpected":true}`,
	}
	for _, payload := range values {
		if _, err := ParseCommand([]byte(payload)); err == nil {
			t.Fatalf("payload accepted: %s", payload)
		}
	}
}
