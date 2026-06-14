// Package input defines the Provider interface for reading tag values from
// external data sources. Implementations can wrap OPC-UA, Modbus, REST APIs,
// or the built-in synthetic simulator. The interface is intentionally minimal
// so real-data adapters can be added without touching the PLC runtime.
package input

import (
	"context"
	"time"
)

// Provider is the interface that all input data sources must implement.
type Provider interface {
	// Name returns a human-readable identifier for this source.
	Name() string
	// Read fetches the current values of all tags from the source.
	Read(ctx context.Context) ([]TagValue, error)
	// Close releases any resources held by the provider.
	Close() error
}

// TagValue is a single tag reading from a Provider.
type TagValue struct {
	Name      string
	Value     float64
	Quality   Quality
	Timestamp time.Time
}

// Quality indicates the reliability of a TagValue.
type Quality uint8

const (
	QualityGood      Quality = 0
	QualityUncertain Quality = 1
	QualityBad       Quality = 2
)

func (q Quality) String() string {
	switch q {
	case QualityGood:
		return "good"
	case QualityUncertain:
		return "uncertain"
	default:
		return "bad"
	}
}
