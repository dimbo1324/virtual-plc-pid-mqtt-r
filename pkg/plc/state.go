package plc

// State describes the lifecycle of a Runtime.
type State string

const (
	StateStopped  State = "stopped"
	StateStarting State = "starting"
	StateRunning  State = "running"
	StateStopping State = "stopping"
	StateFaulted  State = "faulted"
)
