package input_test

import (
	"context"
	"testing"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/input"
)

func TestQuality_String(t *testing.T) {
	cases := []struct {
		q    input.Quality
		want string
	}{
		{input.QualityGood, "good"},
		{input.QualityUncertain, "uncertain"},
		{input.QualityBad, "bad"},
	}
	for _, tc := range cases {
		if got := tc.q.String(); got != tc.want {
			t.Errorf("Quality(%d).String() = %q, want %q", tc.q, got, tc.want)
		}
	}
}

func TestSyntheticProvider_Name(t *testing.T) {
	p := input.NewSyntheticProvider("test", func() map[string]float64 { return nil })
	if p.Name() != "test" {
		t.Errorf("Name() = %q, want %q", p.Name(), "test")
	}
}

func TestSyntheticProvider_Close(t *testing.T) {
	p := input.NewSyntheticProvider("test", func() map[string]float64 { return nil })
	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestSyntheticProvider_Read(t *testing.T) {
	snapFn := func() map[string]float64 {
		return map[string]float64{"pv": 42.0, "sp": 50.0}
	}
	p := input.NewSyntheticProvider("sim", snapFn)
	tags, err := p.Read(context.Background())
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("Read() len = %d, want 2", len(tags))
	}
	for _, tv := range tags {
		if tv.Quality != input.QualityGood {
			t.Errorf("tag %q quality = %v, want Good", tv.Name, tv.Quality)
		}
		if tv.Timestamp.IsZero() {
			t.Errorf("tag %q timestamp is zero", tv.Name)
		}
	}
}

func TestSyntheticProvider_ImplementsProvider(t *testing.T) {
	var _ input.Provider = (*input.SyntheticProvider)(nil)
}
