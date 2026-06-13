package simulator

// Quality describes the current confidence in a simulated process value.
type Quality string

const (
	QualityGood      Quality = "good"
	QualityUncertain Quality = "uncertain"
	QualityBad       Quality = "bad"
)

// Valid reports whether quality is a supported value.
func (q Quality) Valid() bool {
	switch q {
	case QualityGood, QualityUncertain, QualityBad:
		return true
	default:
		return false
	}
}
