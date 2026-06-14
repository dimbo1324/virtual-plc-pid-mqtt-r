package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// TelemetrySample is one control-loop measurement persisted per scan.
type TelemetrySample struct {
	Timestamp   time.Time
	DeviceID    string
	ScanCounter uint64
	LoopName    string
	SP          float64
	PV          float64
	MV          float64
	Error       float64
	Mode        string
	Quality     string
	Unit        string
}

// EventRecord is a single application or runtime event.
type EventRecord struct {
	Timestamp time.Time
	Level     string
	Type      string
	Message   string
	Details   map[string]any
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
