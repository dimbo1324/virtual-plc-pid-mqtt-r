package storage

import (
	"context"
	"math"
	"testing"
	"time"
)

func sampleAt(t time.Time, loop string) TelemetrySample {
	return TelemetrySample{
		Timestamp: t, DeviceID: "vplc_001", ScanCounter: 1,
		LoopName: loop, SP: 6.0, PV: 5.8, MV: 62.0, Error: 0.2,
		Mode: "auto", Quality: "good", Unit: "bar",
	}
}

func TestInsertTelemetrySample(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	sample := sampleAt(time.Now(), "pressure")
	if err := store.InsertTelemetrySamples(ctx, []TelemetrySample{sample}); err != nil {
		t.Fatalf("InsertTelemetrySamples: %v", err)
	}
}

func TestInsertTelemetryBatch(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	now := time.Now()

	samples := []TelemetrySample{
		sampleAt(now, "pressure"),
		sampleAt(now, "temperature"),
		sampleAt(now, "level"),
	}
	if err := store.InsertTelemetrySamples(ctx, samples); err != nil {
		t.Fatalf("InsertTelemetrySamples batch: %v", err)
	}
}

func TestInsertTelemetryEmptyBatch(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	// Empty batch must be a no-op, not an error.
	if err := store.InsertTelemetrySamples(ctx, nil); err != nil {
		t.Fatalf("empty batch should be no-op, got: %v", err)
	}
}

func TestInsertTelemetry_InvalidRejectsNaN(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := sampleAt(time.Now(), "pressure")
	bad.PV = math.NaN()
	if err := store.InsertTelemetrySamples(ctx, []TelemetrySample{bad}); err == nil {
		t.Error("NaN PV should be rejected")
	}
}

func TestInsertTelemetry_InvalidRejectsZeroTimestamp(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := sampleAt(time.Time{}, "pressure")
	if err := store.InsertTelemetrySamples(ctx, []TelemetrySample{bad}); err == nil {
		t.Error("zero timestamp should be rejected")
	}
}

func TestRecentTelemetry(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 5; i++ {
		s := sampleAt(now.Add(time.Duration(i)*time.Second), "pressure")
		s.ScanCounter = uint64(i + 1)
		if err := store.InsertTelemetrySamples(ctx, []TelemetrySample{s}); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	samples, err := store.RecentTelemetry(ctx, "pressure", 3)
	if err != nil {
		t.Fatalf("RecentTelemetry: %v", err)
	}
	if len(samples) != 3 {
		t.Errorf("expected 3 samples, got %d", len(samples))
	}
	// Results are newest-first (ORDER BY id DESC).
	if samples[0].ScanCounter < samples[1].ScanCounter {
		t.Error("expected newest-first ordering")
	}
}
