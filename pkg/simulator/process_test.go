package simulator_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func TestProcessValueMovesTowardTarget(t *testing.T) {
	tests := []struct {
		name      string
		initialPV float64
		base      float64
		mv        float64
		wantMove  func(before, after float64) bool
	}{
		{
			name:      "increases when target is above PV",
			initialPV: 10,
			base:      0,
			mv:        50,
			wantMove:  func(before, after float64) bool { return after > before },
		},
		{
			name:      "decreases when target is below PV",
			initialPV: 50,
			base:      0,
			mv:        10,
			wantMove:  func(before, after float64) bool { return after < before },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			config.InitialPV = tt.initialPV
			config.Base = tt.base
			process := mustProcess(t, config)
			if err := process.ApplyMV(tt.mv); err != nil {
				t.Fatalf("ApplyMV() error = %v", err)
			}
			before := process.Snapshot().PV
			after, err := process.Step(time.Second)
			if err != nil {
				t.Fatalf("Step() error = %v", err)
			}
			if !tt.wantMove(before, after.PV) {
				t.Fatalf("PV moved from %v to %v, unexpected direction toward target %v", before, after.PV, after.Target)
			}
			if math.Abs(after.Target-after.PV) >= math.Abs(after.Target-before) {
				t.Fatalf("PV %v did not move closer to target %v from %v", after.PV, after.Target, before)
			}
		})
	}
}

func TestProcessValueIsClampedToPhysicalLimits(t *testing.T) {
	tests := []struct {
		name string
		mv   float64
		want float64
	}{
		{name: "maximum", mv: 1000, want: 100},
		{name: "minimum", mv: -1000, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			config.TauSeconds = 0.1
			process := mustProcess(t, config)
			if err := process.ApplyMV(tt.mv); err != nil {
				t.Fatalf("ApplyMV() error = %v", err)
			}
			snapshot, err := process.Step(time.Second)
			if err != nil {
				t.Fatalf("Step() error = %v", err)
			}
			if snapshot.PV != tt.want {
				t.Fatalf("Snapshot().PV = %v, want %v", snapshot.PV, tt.want)
			}
			if snapshot.Quality != simulator.QualityUncertain {
				t.Fatalf("Snapshot().Quality = %q, want %q", snapshot.Quality, simulator.QualityUncertain)
			}
			if math.IsNaN(snapshot.PV) || math.IsInf(snapshot.PV, 0) {
				t.Fatalf("Snapshot().PV = %v, want finite clamped value", snapshot.PV)
			}
		})
	}
}

func TestApplyMVRejectsNonFiniteValuesWithoutChangingState(t *testing.T) {
	process := mustProcess(t, validConfig())
	tests := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, value := range tests {
		before := process.Snapshot()
		err := process.ApplyMV(value)
		if !errors.Is(err, simulator.ErrInvalidMV) {
			t.Fatalf("ApplyMV(%v) error = %v, want %v", value, err, simulator.ErrInvalidMV)
		}
		if after := process.Snapshot(); after != before {
			t.Fatalf("Snapshot() changed after ApplyMV(%v): got %+v, want %+v", value, after, before)
		}
	}
}

func TestStepRejectsInvalidDurationWithoutChangingState(t *testing.T) {
	process := mustProcess(t, validConfig())
	for _, duration := range []time.Duration{0, -time.Second} {
		before := process.Snapshot()
		snapshot, err := process.Step(duration)
		if !errors.Is(err, simulator.ErrInvalidDuration) {
			t.Fatalf("Step(%v) error = %v, want %v", duration, err, simulator.ErrInvalidDuration)
		}
		if snapshot != before || process.Snapshot() != before {
			t.Fatalf("state changed after Step(%v): got %+v, want %+v", duration, process.Snapshot(), before)
		}
	}
}

func TestResetRestoresInitialStateAndClearsDisturbance(t *testing.T) {
	config := validConfig()
	process := mustProcess(t, config)
	if err := process.ApplyMV(50); err != nil {
		t.Fatalf("ApplyMV() error = %v", err)
	}
	if err := process.InjectDisturbance(simulator.Disturbance{
		Name:             "load",
		Amplitude:        10,
		RemainingSeconds: 5,
	}); err != nil {
		t.Fatalf("InjectDisturbance() error = %v", err)
	}
	if _, err := process.Step(time.Second); err != nil {
		t.Fatalf("Step() error = %v", err)
	}

	process.Reset()
	snapshot := process.Snapshot()
	if snapshot.PV != config.InitialPV || snapshot.MV != 0 || snapshot.Target != config.Base {
		t.Fatalf("Snapshot() after Reset() = %+v, want initial process state", snapshot)
	}
	if snapshot.DisturbanceActive || snapshot.Quality != simulator.QualityGood {
		t.Fatalf("Snapshot() after Reset() = %+v, want no disturbance and good quality", snapshot)
	}
}

func TestSnapshotAndConfigReturnCopies(t *testing.T) {
	process := mustProcess(t, validConfig())
	snapshot := process.Snapshot()
	snapshot.PV = 999
	config := process.Config()
	config.InitialPV = 999

	if actual := process.Snapshot(); actual.PV == 999 {
		t.Fatalf("Snapshot() exposed mutable internal state: %+v", actual)
	}
	if actual := process.Config(); actual.InitialPV == 999 {
		t.Fatalf("Config() exposed mutable internal state: %+v", actual)
	}
}

func TestApplyMVRejectsTargetOverflowWithoutChangingState(t *testing.T) {
	config := validConfig()
	config.Gain = math.MaxFloat64
	process := mustProcess(t, config)
	before := process.Snapshot()

	err := process.ApplyMV(2)
	if !errors.Is(err, simulator.ErrNonFiniteCalculation) {
		t.Fatalf("ApplyMV() error = %v, want %v", err, simulator.ErrNonFiniteCalculation)
	}
	if after := process.Snapshot(); after != before {
		t.Fatalf("Snapshot() changed after overflow: got %+v, want %+v", after, before)
	}
}

func TestStepRejectsArithmeticOverflowWithoutChangingState(t *testing.T) {
	config := validConfig()
	config.InitialPV = math.MaxFloat64
	config.Min = -math.MaxFloat64
	config.Max = math.MaxFloat64
	config.Base = -math.MaxFloat64
	config.Gain = 0
	process := mustProcess(t, config)
	before := process.Snapshot()

	snapshot, err := process.Step(time.Second)
	if !errors.Is(err, simulator.ErrNonFiniteCalculation) {
		t.Fatalf("Step() error = %v, want %v", err, simulator.ErrNonFiniteCalculation)
	}
	if snapshot != before || process.Snapshot() != before {
		t.Fatalf("state changed after overflow: got %+v, want %+v", process.Snapshot(), before)
	}
}

func TestValidStepNeverReturnsNonFiniteValues(t *testing.T) {
	process := mustProcess(t, validConfig())
	if err := process.ApplyMV(50); err != nil {
		t.Fatalf("ApplyMV() error = %v", err)
	}
	for range 100 {
		snapshot, err := process.Step(100 * time.Millisecond)
		if err != nil {
			t.Fatalf("Step() error = %v", err)
		}
		for name, value := range map[string]float64{
			"PV": snapshot.PV, "MV": snapshot.MV, "Target": snapshot.Target,
		} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("Snapshot().%s = %v, want finite value", name, value)
			}
		}
	}
}

func TestFirstOrderStepMatchesExpectedValue(t *testing.T) {
	config := validConfig()
	process := mustProcess(t, config)
	if err := process.ApplyMV(50); err != nil {
		t.Fatalf("ApplyMV() error = %v", err)
	}

	snapshot, err := process.Step(time.Second)
	if err != nil {
		t.Fatalf("Step() error = %v", err)
	}
	want := config.InitialPV + 1/config.TauSeconds*(50-config.InitialPV)
	if !almostEqual(snapshot.PV, want, 1e-12) {
		t.Fatalf("Snapshot().PV = %v, want %v", snapshot.PV, want)
	}
}
