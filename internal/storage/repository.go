package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// TelemetrySample is one control-loop measurement persisted per scan.
type TelemetrySample struct {
	Timestamp   time.Time `json:"timestamp"`
	DeviceID    string    `json:"device_id"`
	ScanCounter uint64    `json:"scan_counter"`
	LoopName    string    `json:"loop_name"`
	SP          float64   `json:"sp"`
	PV          float64   `json:"pv"`
	MV          float64   `json:"mv"`
	Error       float64   `json:"error"`
	Mode        string    `json:"mode"`
	Quality     string    `json:"quality"`
	Unit        string    `json:"unit"`
}

// EventRecord is a single application or runtime event.
type EventRecord struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
}

// CommandRecord captures a command that was received and its outcome.
type CommandRecord struct {
	Timestamp    time.Time
	CommandID    string
	Source       string
	CommandType  string
	LoopName     string
	PayloadJSON  string
	Status       string
	ErrorMessage string
}

// PIDChangeRecord tracks changes to PID tuning parameters.
type PIDChangeRecord struct {
	Timestamp time.Time
	LoopName  string
	OldKp     *float64
	OldKi     *float64
	OldKd     *float64
	NewKp     float64
	NewKi     float64
	NewKd     float64
	Source    string
	CommandID string
}

func validateTelemetrySample(s TelemetrySample) error {
	if s.Timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp is zero", ErrInvalidRecord)
	}
	if s.LoopName == "" {
		return fmt.Errorf("%w: loop_name is empty", ErrInvalidRecord)
	}
	if !isFinite(s.SP) || !isFinite(s.PV) || !isFinite(s.MV) || !isFinite(s.Error) {
		return fmt.Errorf("%w: non-finite telemetry value", ErrInvalidRecord)
	}
	return nil
}

func validateEventRecord(e EventRecord) error {
	if e.Timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp is zero", ErrInvalidRecord)
	}
	if e.Type == "" {
		return fmt.Errorf("%w: event type is empty", ErrInvalidRecord)
	}
	if e.Message == "" {
		return fmt.Errorf("%w: event message is empty", ErrInvalidRecord)
	}
	return nil
}

func validateCommandRecord(c CommandRecord) error {
	if c.Timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp is zero", ErrInvalidRecord)
	}
	if c.Status == "" {
		return fmt.Errorf("%w: command status is empty", ErrInvalidRecord)
	}
	if c.Source == "" {
		return fmt.Errorf("%w: command source is empty", ErrInvalidRecord)
	}
	return nil
}

func marshalDetails(details map[string]any) (string, error) {
	if len(details) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(details)
	if err != nil {
		return "", fmt.Errorf("marshal event details: %w", err)
	}
	return string(b), nil
}

func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
