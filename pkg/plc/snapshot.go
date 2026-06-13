package plc

import "time"

// PLCStatus contains scan-cycle status.
type PLCStatus struct {
	State              State   `json:"state"`
	ScanIntervalMS     int     `json:"scan_interval_ms"`
	LastScanDurationMS float64 `json:"last_scan_duration_ms"`
	ScanCounter        uint64  `json:"scan_counter"`
}

// LoopSnapshot contains one control loop's current engineering values.
type LoopSnapshot struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	Unit         string  `json:"unit"`
	Setpoint     float64 `json:"sp"`
	ProcessValue float64 `json:"pv"`
	Output       float64 `json:"mv"`
	Error        float64 `json:"error"`
	Mode         string  `json:"mode"`
	Quality      string  `json:"quality"`
	Enabled      bool    `json:"enabled"`
	Kp           float64 `json:"kp"`
	Ki           float64 `json:"ki"`
	Kd           float64 `json:"kd"`
}

// Snapshot is an immutable point-in-time copy of the virtual PLC.
type Snapshot struct {
	Timestamp time.Time               `json:"timestamp"`
	DeviceID  string                  `json:"device_id"`
	PLC       PLCStatus               `json:"plc"`
	Loops     map[string]LoopSnapshot `json:"loops"`
}

func cloneSnapshot(snapshot Snapshot) Snapshot {
	copySnapshot := snapshot
	copySnapshot.Loops = make(map[string]LoopSnapshot, len(snapshot.Loops))
	for name, loop := range snapshot.Loops {
		copySnapshot.Loops[name] = loop
	}
	return copySnapshot
}
