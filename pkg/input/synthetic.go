package input

import (
	"context"
	"time"
)

// SyntheticProvider reads tag values from a snapshot function. It is used in
// demo and test scenarios where the PLC runtime's own simulated output feeds
// back as an input source.
type SyntheticProvider struct {
	name   string
	snapFn func() map[string]float64
}

// NewSyntheticProvider creates a SyntheticProvider that calls snapFn on every
// Read. snapFn must return a map of tag name → current value.
func NewSyntheticProvider(name string, snapFn func() map[string]float64) *SyntheticProvider {
	return &SyntheticProvider{name: name, snapFn: snapFn}
}

// Name returns the provider name supplied at construction.
func (s *SyntheticProvider) Name() string { return s.name }

// Close is a no-op; the synthetic provider holds no external resources.
func (s *SyntheticProvider) Close() error { return nil }

// Read calls snapFn and converts the result to []TagValue with QualityGood.
func (s *SyntheticProvider) Read(_ context.Context) ([]TagValue, error) {
	now := time.Now()
	vals := s.snapFn()
	tags := make([]TagValue, 0, len(vals))
	for name, v := range vals {
		tags = append(tags, TagValue{
			Name:      name,
			Value:     v,
			Quality:   QualityGood,
			Timestamp: now,
		})
	}
	return tags, nil
}
