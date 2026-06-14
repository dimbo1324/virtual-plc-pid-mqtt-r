package storage

import (
	"context"
	"testing"
	"time"
)

func TestInsertEvent(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	event := EventRecord{
		Timestamp: time.Now(),
		Level:     "info",
		Type:      "plc_started",
		Message:   "PLC runtime started",
	}
	if err := store.InsertEvent(ctx, event); err != nil {
		t.Fatalf("InsertEvent: %v", err)
	}
}

func TestInsertEvent_WithDetails(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	event := EventRecord{
		Timestamp: time.Now(),
		Level:     "info",
		Type:      "setpoint_changed",
		Message:   "Setpoint changed",
		Details:   map[string]any{"loop": "pressure", "old_value": 6.0, "new_value": 7.5},
	}
	if err := store.InsertEvent(ctx, event); err != nil {
		t.Fatalf("InsertEvent with details: %v", err)
	}
}

func TestInsertEvent_InvalidZeroTimestamp(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := EventRecord{Level: "info", Type: "x", Message: "y"}
	if err := store.InsertEvent(ctx, bad); err == nil {
		t.Error("zero timestamp should be rejected")
	}
}

func TestInsertEvent_InvalidEmptyType(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := EventRecord{Timestamp: time.Now(), Level: "info", Type: "", Message: "y"}
	if err := store.InsertEvent(ctx, bad); err == nil {
		t.Error("empty event type should be rejected")
	}
}

func TestRecentEvents(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		e := EventRecord{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Level:     "info",
			Type:      "test_event",
			Message:   "test",
		}
		if err := store.InsertEvent(ctx, e); err != nil {
			t.Fatalf("insert event %d: %v", i, err)
		}
	}

	events, err := store.RecentEvents(ctx, 3)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}
