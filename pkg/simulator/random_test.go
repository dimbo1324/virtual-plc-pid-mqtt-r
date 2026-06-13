package simulator_test

import (
	"testing"
	"time"
)

func TestNoNoiseProcessesAreDeterministic(t *testing.T) {
	config := validConfig()
	first := mustProcess(t, config)
	second := mustProcess(t, config)
	if err := first.ApplyMV(50); err != nil {
		t.Fatalf("first ApplyMV() error = %v", err)
	}
	if err := second.ApplyMV(50); err != nil {
		t.Fatalf("second ApplyMV() error = %v", err)
	}

	for step := 0; step < 10; step++ {
		firstSnapshot, err := first.Step(500 * time.Millisecond)
		if err != nil {
			t.Fatalf("first Step() error = %v", err)
		}
		secondSnapshot, err := second.Step(500 * time.Millisecond)
		if err != nil {
			t.Fatalf("second Step() error = %v", err)
		}
		if firstSnapshot != secondSnapshot {
			t.Fatalf("step %d snapshots differ: first %+v, second %+v", step, firstSnapshot, secondSnapshot)
		}
	}
}

func TestSameSeedProducesSameNoisySequence(t *testing.T) {
	config := validConfig()
	config.NoiseStddev = 0.5
	config.RandomSeed = 0
	first := mustProcess(t, config)
	second := mustProcess(t, config)

	for step := 0; step < 20; step++ {
		firstSnapshot, err := first.Step(time.Second)
		if err != nil {
			t.Fatalf("first Step() error = %v", err)
		}
		secondSnapshot, err := second.Step(time.Second)
		if err != nil {
			t.Fatalf("second Step() error = %v", err)
		}
		if firstSnapshot.PV != secondSnapshot.PV {
			t.Fatalf("step %d PV differs: first %v, second %v", step, firstSnapshot.PV, secondSnapshot.PV)
		}
	}
}

func TestDifferentSeedsProduceDifferentNoisySequence(t *testing.T) {
	config := validConfig()
	config.NoiseStddev = 0.5
	first := mustProcess(t, config)
	config.RandomSeed++
	second := mustProcess(t, config)

	different := false
	for range 20 {
		firstSnapshot, err := first.Step(time.Second)
		if err != nil {
			t.Fatalf("first Step() error = %v", err)
		}
		secondSnapshot, err := second.Step(time.Second)
		if err != nil {
			t.Fatalf("second Step() error = %v", err)
		}
		if firstSnapshot.PV != secondSnapshot.PV {
			different = true
			break
		}
	}
	if !different {
		t.Fatal("different seeds produced identical noisy sequences")
	}
}

func TestResetRestartsDeterministicNoiseSequence(t *testing.T) {
	config := validConfig()
	config.NoiseStddev = 0.5
	process := mustProcess(t, config)

	first, err := process.Step(time.Second)
	if err != nil {
		t.Fatalf("first Step() error = %v", err)
	}
	process.Reset()
	second, err := process.Step(time.Second)
	if err != nil {
		t.Fatalf("Step() after Reset() error = %v", err)
	}
	if first.PV != second.PV {
		t.Fatalf("PV after Reset() = %v, want deterministic replay %v", second.PV, first.PV)
	}
}
