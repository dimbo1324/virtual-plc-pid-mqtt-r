package plc

import (
	"context"
	"fmt"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

// CommandType identifies a supported runtime operation.
type CommandType string

const (
	CommandSetSetpoint       CommandType = "set_setpoint"
	CommandSetPIDGains       CommandType = "set_pid_gains"
	CommandSetMode           CommandType = "set_mode"
	CommandSetManualOutput   CommandType = "set_manual_output"
	CommandInjectDisturbance CommandType = "inject_disturbance"
	CommandResetLoop         CommandType = "reset_loop"
	CommandStartPLC          CommandType = "start_plc"
	CommandStopPLC           CommandType = "stop_plc"
)

// Command is the common typed command used by external interfaces.
type Command struct {
	CommandID       string
	Command         CommandType
	Loop            string
	Value           *float64
	Kp              *float64
	Ki              *float64
	Kd              *float64
	Mode            LoopMode
	ManualOutput    *float64
	DurationSeconds *float64
	Source          string
	ReceivedAt      time.Time
}

func (r *Runtime) ApplyCommand(command Command) (Event, error) {
	var event Event
	var err error

	switch command.Command {
	case CommandStartPLC:
		err = r.Start(context.Background())
		event = newEvent("info", EventPLCStarted, "PLC runtime started", commandDetails(command))
	case CommandStopPLC:
		err = r.Stop(context.Background())
		event = newEvent("info", EventPLCStopped, "PLC runtime stopped", commandDetails(command))
	default:
		r.mu.Lock()
		event, err = r.applyLoopCommandLocked(command)
		if err == nil {
			r.snapshot = r.buildSnapshotLocked(time.Now().UTC())
		}
		r.mu.Unlock()
	}

	if err != nil {
		rejected := newEvent("warning", EventCommandRejected, err.Error(), commandDetails(command))
		r.emitEvent(rejected)
		return rejected, err
	}

	if command.Command != CommandStartPLC && command.Command != CommandStopPLC {
		r.emitEvent(event)
	}
	r.emitEvent(newEvent("info", EventCommandApplied, "command applied", commandDetails(command)))
	return event, nil
}

func (r *Runtime) applyLoopCommandLocked(command Command) (Event, error) {
	loop, ok := r.loops[command.Loop]
	if !ok {
		return Event{}, fmt.Errorf("%w: %q", ErrUnknownLoop, command.Loop)
	}
	details := commandDetails(command)
	details["loop"] = loop.name

	switch command.Command {
	case CommandSetSetpoint:
		if command.Value == nil || !finite(*command.Value) {
			return Event{}, fmt.Errorf("%w: setpoint value must be finite", ErrInvalidCommand)
		}
		if loop.hasSetpointLimits && (*command.Value < loop.setpointMin || *command.Value > loop.setpointMax) {
			return Event{}, fmt.Errorf("%w: setpoint must be within [%g, %g]", ErrInvalidCommand, loop.setpointMin, loop.setpointMax)
		}
		old := loop.controller.State().Setpoint
		loop.controller.SetSetpoint(*command.Value)
		details["old_value"] = old
		details["new_value"] = *command.Value
		return newEvent("info", EventSetpointChanged, "setpoint changed", details), nil

	case CommandSetPIDGains:
		if command.Kp == nil || command.Ki == nil || command.Kd == nil {
			return Event{}, fmt.Errorf("%w: kp, ki and kd are required", ErrInvalidCommand)
		}
		if err := loop.controller.SetTunings(*command.Kp, *command.Ki, *command.Kd); err != nil {
			return Event{}, fmt.Errorf("%w: %v", ErrInvalidCommand, err)
		}
		details["kp"] = *command.Kp
		details["ki"] = *command.Ki
		details["kd"] = *command.Kd
		return newEvent("info", EventPIDTuningChanged, "PID tunings changed", details), nil

	case CommandSetMode:
		if !command.Mode.Valid() {
			return Event{}, fmt.Errorf("%w: invalid loop mode %q", ErrInvalidCommand, command.Mode)
		}
		if command.ManualOutput != nil && !finite(*command.ManualOutput) {
			return Event{}, fmt.Errorf("%w: manual output must be finite", ErrInvalidCommand)
		}
		if err := loop.controller.SetMode(command.Mode.pidMode()); err != nil {
			return Event{}, fmt.Errorf("%w: %v", ErrInvalidCommand, err)
		}
		if command.ManualOutput != nil {
			if err := loop.controller.SetManualOutput(*command.ManualOutput); err != nil {
				return Event{}, fmt.Errorf("%w: %v", ErrInvalidCommand, err)
			}
		}
		details["mode"] = command.Mode
		return newEvent("info", EventModeChanged, "loop mode changed", details), nil

	case CommandSetManualOutput:
		if command.Value == nil {
			return Event{}, fmt.Errorf("%w: manual output value is required", ErrInvalidCommand)
		}
		if err := loop.controller.SetManualOutput(*command.Value); err != nil {
			return Event{}, fmt.Errorf("%w: %v", ErrInvalidCommand, err)
		}
		details["value"] = *command.Value
		return newEvent("info", EventManualOutputChanged, "manual output changed", details), nil

	case CommandInjectDisturbance:
		if command.Value == nil || !finite(*command.Value) {
			return Event{}, fmt.Errorf("%w: disturbance amplitude must be finite", ErrInvalidCommand)
		}
		duration := 30.0
		if command.DurationSeconds != nil {
			duration = *command.DurationSeconds
		}
		if !finite(duration) || duration <= 0 {
			return Event{}, fmt.Errorf("%w: disturbance duration must be greater than zero", ErrInvalidCommand)
		}
		if err := loop.process.InjectDisturbance(simulator.Disturbance{Name: "command", Amplitude: *command.Value, RemainingSeconds: duration}); err != nil {
			return Event{}, fmt.Errorf("%w: %v", ErrInvalidCommand, err)
		}
		details["amplitude"] = *command.Value
		details["duration_seconds"] = duration
		return newEvent("info", EventDisturbanceInjected, "process disturbance injected", details), nil

	case CommandResetLoop:
		loop.controller.Reset()
		loop.process.Reset()
		return newEvent("info", EventLoopReset, "loop reset", details), nil

	default:
		return Event{}, fmt.Errorf("%w: unsupported command %q", ErrInvalidCommand, command.Command)
	}
}

func commandDetails(command Command) map[string]any {
	details := map[string]any{"command": command.Command}
	if command.CommandID != "" {
		details["command_id"] = command.CommandID
	}
	if command.Source != "" {
		details["source"] = command.Source
	}
	return details
}
