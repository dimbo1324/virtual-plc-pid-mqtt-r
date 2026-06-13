package pid

// State is a point-in-time copy of the controller state.
type State struct {
	Setpoint     float64
	ProcessValue float64
	Output       float64
	Error        float64
	Proportional float64
	Integral     float64
	Derivative   float64
	LastError    float64
	LastPV       float64
	Mode         Mode
	Enabled      bool
}
