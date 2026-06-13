package mqttx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

const maxCommandPayloadBytes = 64 * 1024

type commandPayload struct {
	CommandID       string          `json:"command_id"`
	Command         plc.CommandType `json:"command"`
	Loop            string          `json:"loop,omitempty"`
	Value           *float64        `json:"value,omitempty"`
	Kp              *float64        `json:"kp,omitempty"`
	Ki              *float64        `json:"ki,omitempty"`
	Kd              *float64        `json:"kd,omitempty"`
	Mode            plc.LoopMode    `json:"mode,omitempty"`
	ManualOutput    *float64        `json:"manual_output,omitempty"`
	DurationSeconds *float64        `json:"duration_seconds,omitempty"`
}

// ParseCommand validates JSON and converts it to the shared PLC command model.
func ParseCommand(payload []byte) (plc.Command, error) {
	if len(payload) == 0 || len(payload) > maxCommandPayloadBytes {
		return plc.Command{}, fmt.Errorf("%w: payload size is invalid", ErrInvalidCommand)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var decoded commandPayload
	if err := decoder.Decode(&decoded); err != nil {
		return plc.Command{}, fmt.Errorf("%w: decode JSON: %v", ErrInvalidCommand, err)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return plc.Command{}, fmt.Errorf("%w: %v", ErrInvalidCommand, err)
	}
	if err := decoded.validate(); err != nil {
		return plc.Command{}, err
	}
	return plc.Command{
		CommandID: decoded.CommandID, Command: decoded.Command, Loop: strings.TrimSpace(decoded.Loop),
		Value: decoded.Value, Kp: decoded.Kp, Ki: decoded.Ki, Kd: decoded.Kd,
		Mode: decoded.Mode, ManualOutput: decoded.ManualOutput, DurationSeconds: decoded.DurationSeconds,
		Source: "mqtt", ReceivedAt: time.Now().UTC(),
	}, nil
}

func (p commandPayload) validate() error {
	requireLoop := func() error {
		if strings.TrimSpace(p.Loop) == "" {
			return fmt.Errorf("%w: loop is required for %s", ErrInvalidCommand, p.Command)
		}
		return nil
	}
	finite := func(value *float64) bool {
		return value != nil && !math.IsNaN(*value) && !math.IsInf(*value, 0)
	}

	switch p.Command {
	case plc.CommandSetSetpoint, plc.CommandSetManualOutput, plc.CommandInjectDisturbance:
		if err := requireLoop(); err != nil {
			return err
		}
		if !finite(p.Value) {
			return fmt.Errorf("%w: finite value is required for %s", ErrInvalidCommand, p.Command)
		}
		if p.Command == plc.CommandInjectDisturbance && p.DurationSeconds != nil && (!finite(p.DurationSeconds) || *p.DurationSeconds <= 0) {
			return fmt.Errorf("%w: disturbance duration must be greater than zero", ErrInvalidCommand)
		}
	case plc.CommandSetPIDGains:
		if err := requireLoop(); err != nil {
			return err
		}
		if !finite(p.Kp) || !finite(p.Ki) || !finite(p.Kd) || *p.Kp < 0 || *p.Ki < 0 || *p.Kd < 0 {
			return fmt.Errorf("%w: finite non-negative kp, ki and kd are required", ErrInvalidCommand)
		}
	case plc.CommandSetMode:
		if err := requireLoop(); err != nil {
			return err
		}
		if !p.Mode.Valid() {
			return fmt.Errorf("%w: invalid mode %q", ErrInvalidCommand, p.Mode)
		}
		if p.ManualOutput != nil && !finite(p.ManualOutput) {
			return fmt.Errorf("%w: manual output must be finite", ErrInvalidCommand)
		}
	case plc.CommandResetLoop:
		return requireLoop()
	case plc.CommandStartPLC, plc.CommandStopPLC:
		return nil
	default:
		return fmt.Errorf("%w: unsupported command %q", ErrInvalidCommand, p.Command)
	}
	return nil
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func commandIDFromPayload(payload []byte) string {
	var value struct {
		CommandID string `json:"command_id"`
	}
	_ = json.Unmarshal(payload, &value)
	return value.CommandID
}
