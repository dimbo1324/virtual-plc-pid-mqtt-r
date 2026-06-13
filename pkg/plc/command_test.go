package plc

import (
	"context"
	"math"
	"testing"
)

func ptr(value float64) *float64 { return &value }

func TestLoopCommandsUpdateSnapshotAndEmitEvents(t *testing.T) {
	runtime := newTestRuntime(t)

	tests := []struct {
		name      string
		command   Command
		eventType string
		check     func(t *testing.T, snapshot LoopSnapshot)
	}{
		{
			name: "setpoint", command: Command{Command: CommandSetSetpoint, Loop: "pressure", Value: ptr(7)},
			eventType: EventSetpointChanged,
			check: func(t *testing.T, snapshot LoopSnapshot) {
				if snapshot.Setpoint != 7 {
					t.Fatalf("setpoint = %g, want 7", snapshot.Setpoint)
				}
			},
		},
		{
			name: "PID gains", command: Command{Command: CommandSetPIDGains, Loop: "pressure", Kp: ptr(2), Ki: ptr(0.4), Kd: ptr(0.1)},
			eventType: EventPIDTuningChanged,
			check: func(t *testing.T, snapshot LoopSnapshot) {
				if snapshot.Kp != 2 || snapshot.Ki != 0.4 || snapshot.Kd != 0.1 {
					t.Fatalf("gains = %g/%g/%g", snapshot.Kp, snapshot.Ki, snapshot.Kd)
				}
			},
		},
		{
			name: "manual mode", command: Command{Command: CommandSetMode, Loop: "pressure", Mode: LoopModeManual, ManualOutput: ptr(35)},
			eventType: EventModeChanged,
			check: func(t *testing.T, snapshot LoopSnapshot) {
				if snapshot.Mode != string(LoopModeManual) || snapshot.Output != 35 {
					t.Fatalf("mode/output = %q/%g", snapshot.Mode, snapshot.Output)
				}
			},
		},
		{
			name: "manual output", command: Command{Command: CommandSetManualOutput, Loop: "pressure", Value: ptr(42)},
			eventType: EventManualOutputChanged,
			check: func(t *testing.T, snapshot LoopSnapshot) {
				if snapshot.Output != 42 {
					t.Fatalf("manual output = %g, want 42", snapshot.Output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := runtime.ApplyCommand(tt.command)
			if err != nil {
				t.Fatalf("ApplyCommand() error = %v", err)
			}
			if event.Type != tt.eventType {
				t.Fatalf("event type = %q, want %q", event.Type, tt.eventType)
			}
			tt.check(t, runtime.Snapshot().Loops["pressure"])
		})
	}
}

func TestCommandsEmitAppliedEvent(t *testing.T) {
	runtime := newTestRuntime(t)
	command := Command{CommandID: "cmd-1", Command: CommandSetSetpoint, Loop: "pressure", Value: ptr(7)}
	if _, err := runtime.ApplyCommand(command); err != nil {
		t.Fatalf("ApplyCommand() error = %v", err)
	}
	first := <-runtime.Events()
	second := <-runtime.Events()
	if first.Type != EventSetpointChanged || second.Type != EventCommandApplied {
		t.Fatalf("events = %q, %q", first.Type, second.Type)
	}
}

func TestStartAndStopCommands(t *testing.T) {
	runtime := newTestRuntime(t)
	if event, err := runtime.ApplyCommand(Command{Command: CommandStartPLC}); err != nil || event.Type != EventPLCStarted {
		t.Fatalf("start event/error = %q/%v", event.Type, err)
	}
	waitForScans(t, runtime, 1)
	if event, err := runtime.ApplyCommand(Command{Command: CommandStopPLC}); err != nil || event.Type != EventPLCStopped {
		t.Fatalf("stop event/error = %q/%v", event.Type, err)
	}
	if runtime.State() != StateStopped {
		t.Fatalf("state = %q, want stopped", runtime.State())
	}
}

func TestDisturbanceAndResetCommands(t *testing.T) {
	runtime := newTestRuntime(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer runtime.Stop(context.Background())
	waitForScans(t, runtime, 2)

	before := runtime.Snapshot().Loops["pressure"].ProcessValue
	if _, err := runtime.ApplyCommand(Command{Command: CommandInjectDisturbance, Loop: "pressure", Value: ptr(-10), DurationSeconds: ptr(1)}); err != nil {
		t.Fatalf("inject disturbance: %v", err)
	}
	after := waitForScans(t, runtime, runtime.Snapshot().PLC.ScanCounter+3).Loops["pressure"].ProcessValue
	if after >= before {
		t.Fatalf("disturbed PV = %g, want below %g", after, before)
	}

	if _, err := runtime.ApplyCommand(Command{Command: CommandResetLoop, Loop: "pressure"}); err != nil {
		t.Fatalf("reset loop: %v", err)
	}
	reset := runtime.Snapshot().Loops["pressure"]
	if reset.ProcessValue != testConfig().Loops[0].Process.InitialPV || reset.Output != testConfig().Loops[0].PID.OutputMin {
		t.Fatalf("reset snapshot PV/output = %g/%g", reset.ProcessValue, reset.Output)
	}
}

func TestInvalidCommandsAreRejected(t *testing.T) {
	runtime := newTestRuntime(t)
	tests := []Command{
		{Command: CommandSetSetpoint, Loop: "missing", Value: ptr(5)},
		{Command: CommandType("unknown"), Loop: "pressure"},
		{Command: CommandSetSetpoint, Loop: "pressure", Value: ptr(math.NaN())},
		{Command: CommandSetSetpoint, Loop: "pressure", Value: ptr(99)},
		{Command: CommandInjectDisturbance, Loop: "pressure", Value: ptr(1), DurationSeconds: ptr(0)},
		{Command: CommandInjectDisturbance, Loop: "pressure", Value: ptr(math.Inf(1))},
	}
	for _, command := range tests {
		event, err := runtime.ApplyCommand(command)
		if err == nil {
			t.Fatalf("command %+v accepted", command)
		}
		if event.Type != EventCommandRejected {
			t.Fatalf("event type = %q, want rejected", event.Type)
		}
	}
}

func TestRejectedCommandEventIncludesDiagnosticContext(t *testing.T) {
	runtime := newTestRuntime(t)
	event, err := runtime.ApplyCommand(Command{
		CommandID: "bad-1", Command: CommandSetSetpoint, Loop: "missing",
		Value: ptr(5), Source: "mqtt",
	})
	if err == nil {
		t.Fatal("unknown loop command accepted")
	}
	if event.Type != EventCommandRejected || event.Level != "warning" {
		t.Fatalf("rejection event = %+v", event)
	}
	if event.Details["command_id"] != "bad-1" || event.Details["source"] != "mqtt" {
		t.Fatalf("rejection details = %+v", event.Details)
	}
}

func TestDisabledLoopUsesSafeOutput(t *testing.T) {
	config := testConfig()
	config.Loops[0].Enabled = false
	config.Loops[0].Mode = LoopModeDisabled
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer runtime.Stop(context.Background())
	snapshot := waitForScans(t, runtime, 2).Loops["pressure"]
	if snapshot.Output != config.Loops[0].PID.OutputMin || snapshot.Mode != string(LoopModeDisabled) {
		t.Fatalf("disabled output/mode = %g/%q", snapshot.Output, snapshot.Mode)
	}
}
