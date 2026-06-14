package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestRetention_DeletesOldRows(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          filepath.Join(dir, "test.db"),
		EventsJSONLPath:     filepath.Join(dir, "events.jsonl"),
		RetentionMaxSamples: 3,
		WriteQueueSize:      16,
	}
	store, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now()
	for i := 0; i < 7; i++ {
		s := sampleAt(now.Add(time.Duration(i)*time.Second), "pressure")
		s.ScanCounter = uint64(i + 1)
		if err := store.InsertTelemetrySamples(ctx, []TelemetrySample{s}); err != nil {
			t.Fatalf("insert %d: %v", i, err)
		}
	}

	if err := store.ApplyRetention(ctx); err != nil {
		t.Fatalf("ApplyRetention: %v", err)
	}

	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telemetry_samples`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows after retention, got %d", count)
	}
}
