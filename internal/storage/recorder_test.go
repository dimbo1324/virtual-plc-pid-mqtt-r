package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func openTestRecorder(t *testing.T) (*Recorder, *Store) {
	t.Helper()
	store := openTestStore(t)
	dir := t.TempDir()
	jsonlPath := filepath.Join(dir, "events.jsonl")
	jw, err := NewJSONLWriter(jsonlPath)
	if err != nil {
		t.Fatalf("NewJSONLWriter: %v", err)
	}
	t.Cleanup(func() { _ = jw.Close() })
	rec := NewRecorder(store, jw, 64, nil)
	return rec, store
}

func makeSnapshot() plc.Snapshot {
	return plc.Snapshot{
		Timestamp: time.Now(),
		DeviceID:  "vplc_001",
		PLC:       plc.PLCStatus{State: plc.StateRunning, ScanCounter: 1},
		Loops: map[string]plc.LoopSnapshot{
			"pressure": {
				Name: "pressure", DisplayName: "Pressure", Unit: "bar",
				Setpoint: 6.0, ProcessValue: 5.8, Output: 62.0, Error: 0.2,
				Mode: "auto", Quality: "good",
			},
		},
	}
}

func TestRecorder_RecordSnapshotAsync(t *testing.T) {
	rec, store := openTestRecorder(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec.Start(ctx)

	snap := makeSnapshot()
	if !rec.RecordSnapshot(snap) {
		t.Error("expected RecordSnapshot to return true (queue not full)")
	}

	// Give the worker time to process.
	time.Sleep(100 * time.Millisecond)
	cancel()
	_ = rec.Stop(context.Background())

	samples, err := store.RecentTelemetry(context.Background(), "pressure", 10)
	if err != nil {
		t.Fatalf("RecentTelemetry: %v", err)
	}
	if len(samples) == 0 {
		t.Error("expected at least one telemetry sample after recording snapshot")
	}
}

func TestRecorder_RecordEventAsync(t *testing.T) {
	rec, store := openTestRecorder(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec.Start(ctx)

	event := EventRecord{
		Timestamp: time.Now(),
		Level:     "info",
		Type:      "plc_started",
		Message:   "test",
	}
	if !rec.RecordEvent(event) {
		t.Error("expected RecordEvent to return true")
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	_ = rec.Stop(context.Background())

	events, err := store.RecentEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}
	if len(events) == 0 {
		t.Error("expected at least one event after recording")
	}
}

func TestRecorder_FullQueueDoesNotBlock(t *testing.T) {
	// Tiny queue, no worker running — submissions must all return false without
	// blocking the caller.
	store := openTestStore(t)
	rec := NewRecorder(store, nil, 2, nil)

	snap := makeSnapshot()
	// Fill the queue.
	rec.RecordSnapshot(snap)
	rec.RecordSnapshot(snap)

	// This must not block.
	done := make(chan bool, 1)
	go func() {
		result := rec.RecordSnapshot(snap)
		done <- result
	}()

	select {
	case got := <-done:
		if got {
			t.Error("expected false when queue is full")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("RecordSnapshot blocked when queue was full")
	}
}

func TestRecorder_StopsCleanly(t *testing.T) {
	rec, _ := openTestRecorder(t)
	ctx, cancel := context.WithCancel(context.Background())
	rec.Start(ctx)

	event := EventRecord{
		Timestamp: time.Now(), Level: "info",
		Type: "test_event", Message: "test",
	}
	rec.RecordEvent(event)

	cancel()
	if err := rec.Stop(context.Background()); err != nil {
		t.Errorf("Stop returned error: %v", err)
	}
}
