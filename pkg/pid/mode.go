package pid

// Mode defines how the controller produces its output.
type Mode string

const (
	ModeAuto     Mode = "auto"
	ModeManual   Mode = "manual"
	ModeHold     Mode = "hold"
	ModeDisabled Mode = "disabled"
)

// Valid reports whether the mode is supported by the controller.
func (m Mode) Valid() bool {
	switch m {
	case ModeAuto, ModeManual, ModeHold, ModeDisabled:
		return true
	default:
		return false
	}
}
