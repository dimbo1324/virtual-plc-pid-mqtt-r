package plc

import "time"

const (
	EventPLCStarted          = "plc_started"
	EventPLCStopped          = "plc_stopped"
	EventScanOverrun         = "plc_scan_overrun"
	EventCommandApplied      = "command_applied"
	EventCommandRejected     = "command_rejected"
	EventSetpointChanged     = "setpoint_changed"
	EventPIDTuningChanged    = "pid_tuning_changed"
	EventModeChanged         = "mode_changed"
	EventManualOutputChanged = "manual_output_changed"
	EventDisturbanceInjected = "disturbance_injected"
	EventLoopReset           = "loop_reset"
	EventRuntimeFaulted      = "runtime_faulted"
)

// Event describes an important runtime or command action.
type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Type      string         `json:"event_type"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
}

func newEvent(level, eventType, message string, details map[string]any) Event {
	return Event{Timestamp: time.Now().UTC(), Level: level, Type: eventType, Message: message, Details: details}
}
