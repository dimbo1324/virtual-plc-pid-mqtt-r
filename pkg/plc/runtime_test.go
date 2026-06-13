package plc

import (
	"context"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func testConfig() Config {
	return Config{
		DeviceID: "test-plc", ScanInterval: 5 * time.Millisecond,
		PublishInterval: 10 * time.Millisecond, UIUpdateInterval: 10 * time.Millisecond,
		ScanOverrunWarning: time.Second,
		Loops: []LoopConfig{{
			Name: "pressure", DisplayName: "Pressure", Unit: "bar", Enabled: true,
			Mode: LoopModeAuto, Setpoint: 8, SetpointMin: 0, SetpointMax: 12,
			PID: pid.Config{Kp: 8, Ki: 1, OutputMin: 0, OutputMax: 100},
			Process: simulator.Config{
				InitialPV: 2, Min: 0, Max: 12, Base: 0, Gain: 0.12,
				TauSeconds: 0.1, RandomSeed: 1,
			},
		}},
	}
}

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()
	runtime, err := NewRuntime(testConfig())
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	return runtime
}

func waitForScans(t *testing.T, runtime *Runtime, count uint64) Snapshot {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		snapshot := runtime.Snapshot()
		if snapshot.PLC.ScanCounter >= count {
			return snapshot
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("scan counter did not reach %d; got %d", count, runtime.Snapshot().PLC.ScanCounter)
	return Snapshot{}
}

func TestNewRuntimeValidation(t *testing.T) {
	if _, err := NewRuntime(testConfig()); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	emptyDevice := testConfig()
	emptyDevice.DeviceID = ""
	if _, err := NewRuntime(emptyDevice); err == nil {
		t.Fatal("empty device ID accepted")
	}

	duplicate := testConfig()
	duplicate.Loops = append(duplicate.Loops, duplicate.Loops[0])
	if _, err := NewRuntime(duplicate); err == nil {
		t.Fatal("duplicate loop name accepted")
	}
}

func TestRuntimeStartsScansAndStopsCleanly(t *testing.T) {
	runtime := newTestRuntime(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("idempotent Start() error = %v", err)
	}
	snapshot := waitForScans(t, runtime, 4)
	if snapshot.PLC.State != StateRunning {
		t.Fatalf("state = %q, want running", snapshot.PLC.State)
	}
	if len(snapshot.Loops) != len(testConfig().Loops) {
		t.Fatalf("snapshot loops = %d, want %d", len(snapshot.Loops), len(testConfig().Loops))
	}
	if snapshot.Loops["pressure"].ProcessValue <= testConfig().Loops[0].Process.InitialPV {
		t.Fatalf("PV = %g, want movement above initial PV", snapshot.Loops["pressure"].ProcessValue)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := runtime.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := runtime.Stop(stopCtx); err != nil {
		t.Fatalf("idempotent Stop() error = %v", err)
	}
	if runtime.State() != StateStopped {
		t.Fatalf("state = %q, want stopped", runtime.State())
	}
}

func TestRuntimeEmitsScanOverrun(t *testing.T) {
	config := testConfig()
	config.ScanOverrunWarning = time.Nanosecond
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer runtime.Stop(context.Background())

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case event := <-runtime.Events():
			if event.Type == EventScanOverrun {
				return
			}
		case <-deadline:
			t.Fatal("scan overrun event was not emitted")
		}
	}
}

func TestFullOutputChannelsDoNotBlockRuntime(t *testing.T) {
	config := testConfig()
	config.ScanInterval = time.Millisecond
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	for i := 0; i < defaultEventBuffer+10; i++ {
		runtime.emitEvent(newEvent("info", "buffer_test", "fill event buffer", nil))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	waitForScans(t, runtime, defaultSnapshotBuffer+10)
	if err := runtime.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}
