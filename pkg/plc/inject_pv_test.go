package plc_test

import (
	"context"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func buildTestRuntime(t *testing.T, scanInterval time.Duration) *plc.Runtime {
	t.Helper()
	cfg := plc.Config{
		DeviceID:           "test",
		ScanInterval:       scanInterval,
		PublishInterval:    time.Second,
		UIUpdateInterval:   time.Second,
		ScanOverrunWarning: time.Second,
		Loops: []plc.LoopConfig{{
			Name: "loop1", DisplayName: "Loop 1", Unit: "bar",
			Enabled: true, Mode: plc.LoopModeAuto, Setpoint: 50,
			PID: pid.Config{
				Name: "loop1", Kp: 1, Ki: 0, Kd: 0,
				OutputMin: 0, OutputMax: 100, Setpoint: 50, Mode: pid.ModeAuto, Enabled: true,
			},
			Process: simulator.Config{
				Name: "loop1", Min: 0, Max: 100, InitialPV: 10,
				Gain: 1, TauSeconds: 100,
			},
		}},
	}
	rt, err := plc.NewRuntime(cfg)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	return rt
}

func TestInjectPV_OverridesSimulator(t *testing.T) {
	rt := buildTestRuntime(t, 10*time.Millisecond)

	const injectedPV = 42.0
	rt.InjectPV("loop1", injectedPV)

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = rt.Stop(context.Background()) })

	time.Sleep(60 * time.Millisecond)

	snap := rt.Snapshot()
	loop, ok := snap.Loops["loop1"]
	if !ok {
		t.Fatal("loop1 not in snapshot")
	}
	if loop.ProcessValue != injectedPV {
		t.Errorf("ProcessValue = %v, want %v (injected PV)", loop.ProcessValue, injectedPV)
	}
}

func TestInjectPV_UnknownLoop(t *testing.T) {
	rt := buildTestRuntime(t, 10*time.Millisecond)
	if rt.InjectPV("nonexistent", 1.0) {
		t.Error("InjectPV returned true for nonexistent loop")
	}
}

func TestClearPV_RestoresSimulator(t *testing.T) {
	rt := buildTestRuntime(t, 10*time.Millisecond)

	rt.InjectPV("loop1", 42.0)

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = rt.Stop(context.Background()) })

	time.Sleep(40 * time.Millisecond)

	snap := rt.Snapshot()
	if snap.Loops["loop1"].ProcessValue != 42.0 {
		t.Skip("race: injected PV not yet visible in snapshot, skipping")
	}

	rt.ClearPV("loop1")

	time.Sleep(40 * time.Millisecond)

	snap = rt.Snapshot()
	pv := snap.Loops["loop1"].ProcessValue
	if pv == 42.0 {
		t.Error("ClearPV did not remove injected PV; simulator should have taken over")
	}
}

func TestScanCounter_Advances(t *testing.T) {
	rt := buildTestRuntime(t, 5*time.Millisecond)

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = rt.Stop(context.Background()) })

	time.Sleep(80 * time.Millisecond)

	snap := rt.Snapshot()
	if snap.PLC.ScanCounter < 5 {
		t.Errorf("ScanCounter = %d, want >= 5 after 80ms with 5ms scan interval", snap.PLC.ScanCounter)
	}
}
