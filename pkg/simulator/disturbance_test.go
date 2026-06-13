package simulator_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func TestDisturbanceAffectsProcessMovement(t *testing.T) {
	config := validConfig()
	config.InitialPV = 0
	baseline := mustProcess(t, config)
	disturbed := mustProcess(t, config)
	if err := disturbed.InjectDisturbance(simulator.Disturbance{
		Name:             "load step",
		Amplitude:        20,
		RemainingSeconds: 5,
	}); err != nil {
		t.Fatalf("InjectDisturbance() error = %v", err)
	}

	baselineSnapshot, err := baseline.Step(time.Second)
	if err != nil {
		t.Fatalf("baseline Step() error = %v", err)
	}
	disturbedSnapshot, err := disturbed.Step(time.Second)
	if err != nil {
		t.Fatalf("disturbed Step() error = %v", err)
	}
	if disturbedSnapshot.PV <= baselineSnapshot.PV {
		t.Fatalf("disturbed PV = %v, want greater than baseline %v", disturbedSnapshot.PV, baselineSnapshot.PV)
	}
}

func TestDisturbanceExpiresAfterDuration(t *testing.T) {
	process := mustProcess(t, validConfig())
	if err := process.InjectDisturbance(simulator.Disturbance{
		Name:             "short step",
		Amplitude:        10,
		RemainingSeconds: 1.5,
	}); err != nil {
		t.Fatalf("InjectDisturbance() error = %v", err)
	}

	first, err := process.Step(time.Second)
	if err != nil {
		t.Fatalf("first Step() error = %v", err)
	}
	if !first.DisturbanceActive {
		t.Fatal("disturbance expired too early")
	}
	second, err := process.Step(500 * time.Millisecond)
	if err != nil {
		t.Fatalf("second Step() error = %v", err)
	}
	if second.DisturbanceActive {
		t.Fatal("disturbance remained active after duration elapsed")
	}
}

func TestClearDisturbanceRemovesItImmediately(t *testing.T) {
	process := mustProcess(t, validConfig())
	if err := process.InjectDisturbance(simulator.Disturbance{
		Name:             "manual step",
		Amplitude:        10,
		RemainingSeconds: 10,
	}); err != nil {
		t.Fatalf("InjectDisturbance() error = %v", err)
	}
	process.ClearDisturbance()

	snapshot := process.Snapshot()
	if snapshot.DisturbanceActive {
		t.Fatal("Snapshot().DisturbanceActive = true after ClearDisturbance()")
	}
	if snapshot.Target != process.Config().Base+process.Config().Gain*snapshot.MV {
		t.Fatalf("Snapshot().Target = %v after clear, want undisturbed target", snapshot.Target)
	}
}

func TestInjectDisturbanceRejectsInvalidValues(t *testing.T) {
	process := mustProcess(t, validConfig())
	tests := []struct {
		name        string
		disturbance simulator.Disturbance
	}{
		{name: "empty name", disturbance: simulator.Disturbance{Amplitude: 1, RemainingSeconds: 1}},
		{name: "NaN amplitude", disturbance: simulator.Disturbance{Name: "bad", Amplitude: math.NaN(), RemainingSeconds: 1}},
		{name: "infinite amplitude", disturbance: simulator.Disturbance{Name: "bad", Amplitude: math.Inf(1), RemainingSeconds: 1}},
		{name: "zero duration", disturbance: simulator.Disturbance{Name: "bad", Amplitude: 1}},
		{name: "infinite duration", disturbance: simulator.Disturbance{Name: "bad", Amplitude: 1, RemainingSeconds: math.Inf(1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := process.Snapshot()
			err := process.InjectDisturbance(tt.disturbance)
			if !errors.Is(err, simulator.ErrInvalidDisturbance) {
				t.Fatalf("InjectDisturbance() error = %v, want %v", err, simulator.ErrInvalidDisturbance)
			}
			if after := process.Snapshot(); after != before {
				t.Fatalf("Snapshot() changed after invalid disturbance: got %+v, want %+v", after, before)
			}
		})
	}
}
