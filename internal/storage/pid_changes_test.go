package storage

import (
	"context"
	"testing"
	"time"
)

func ptr(f float64) *float64 { return &f }

func TestInsertPIDChange(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	change := PIDChangeRecord{
		Timestamp: time.Now(),
		LoopName:  "pressure",
		OldKp:     ptr(3.0),
		OldKi:     ptr(0.25),
		OldKd:     ptr(0.05),
		NewKp:     2.5,
		NewKi:     0.3,
		NewKd:     0.05,
		Source:    "mqtt",
		CommandID: "cmd-002",
	}
	if err := store.InsertPIDChange(ctx, change); err != nil {
		t.Fatalf("InsertPIDChange: %v", err)
	}
}

func TestInsertPIDChange_NilOldValues(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	change := PIDChangeRecord{
		Timestamp: time.Now(),
		LoopName:  "temperature",
		NewKp:     1.8, NewKi: 0.1, NewKd: 0.02,
		Source: "app",
	}
	if err := store.InsertPIDChange(ctx, change); err != nil {
		t.Fatalf("InsertPIDChange nil old values: %v", err)
	}
}

func TestInsertPIDChange_InvalidEmptyLoopName(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := PIDChangeRecord{Timestamp: time.Now(), Source: "mqtt", NewKp: 1, NewKi: 1, NewKd: 1}
	if err := store.InsertPIDChange(ctx, bad); err == nil {
		t.Error("empty loop_name should be rejected")
	}
}
