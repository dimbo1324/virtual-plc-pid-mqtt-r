package simulator

// Snapshot is a point-in-time copy of a simulated process.
type Snapshot struct {
	Name              string
	DisplayName       string
	Unit              string
	PV                float64
	MV                float64
	Target            float64
	Min               float64
	Max               float64
	Quality           Quality
	DisturbanceActive bool
}
