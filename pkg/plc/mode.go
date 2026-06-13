package plc

import "github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"

// LoopMode describes how a control loop produces its output.
type LoopMode string

const (
	LoopModeAuto     LoopMode = "auto"
	LoopModeManual   LoopMode = "manual"
	LoopModeHold     LoopMode = "hold"
	LoopModeDisabled LoopMode = "disabled"
)

// Valid reports whether the mode is supported.
func (m LoopMode) Valid() bool {
	switch m {
	case LoopModeAuto, LoopModeManual, LoopModeHold, LoopModeDisabled:
		return true
	default:
		return false
	}
}

func (m LoopMode) pidMode() pid.Mode {
	return pid.Mode(m)
}
