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

func TestParseCommandAllowsOptionalCommandID(t *testing.T) {
	command, err := ParseCommand([]byte(`{"command":"start_plc"}`))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}
	if command.CommandID != "" || command.Command != plc.CommandStartPLC {
		t.Fatalf("command = %+v", command)
	}
}

func TestParseCommand_ManualOutput(t *testing.T) {
	payload := `{"command":"set_manual_output","loop":"pressure","value":55.0}`
	command, err := ParseCommand([]byte(payload))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}
	if command.Command != plc.CommandSetManualOutput || command.Loop != "pressure" {
		t.Fatalf("command = %+v", command)
	}
	if command.Value == nil || *command.Value != 55.0 {
		t.Fatalf("value = %v, want 55.0", command.Value)
	}
}

func TestParseCommand_InjectDisturbance(t *testing.T) {
	payload := `{"command":"inject_disturbance","loop":"temperature","value":3.5,"duration_seconds":10.0}`
	command, err := ParseCommand([]byte(payload))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}
	if command.Command != plc.CommandInjectDisturbance || command.Loop != "temperature" {
		t.Fatalf("command = %+v", command)
	}
	if command.Value == nil || *command.Value != 3.5 {
		t.Fatalf("value = %v, want 3.5", command.Value)
	}
	if command.DurationSeconds == nil || *command.DurationSeconds != 10.0 {
		t.Fatalf("duration_seconds = %v, want 10.0", command.DurationSeconds)
	}
}

func TestParseCommand_ResetLoop(t *testing.T) {
	payload := `{"command":"reset_loop","loop":"level"}`
	command, err := ParseCommand([]byte(payload))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}
	if command.Command != plc.CommandResetLoop || command.Loop != "level" {
		t.Fatalf("command = %+v", command)
	}
}

func TestCommandIDFromPayload(t *testing.T) {
	cases := []struct {
		payload string
		want    string
	}{
		{`{"command_id":"abc-123","command":"start_plc"}`, "abc-123"},
		{`{"command":"start_plc"}`, ""},
		{`not json`, ""},
	}
	for _, tc := range cases {
		got := commandIDFromPayload([]byte(tc.payload))
		if got != tc.want {
			t.Errorf("commandIDFromPayload(%q) = %q, want %q", tc.payload, got, tc.want)
		}
	}
}
