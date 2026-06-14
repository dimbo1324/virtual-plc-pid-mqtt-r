package storage

import (
	"context"
	"testing"
	"time"
)

func TestInsertCommand(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	cmd := CommandRecord{
		Timestamp:   time.Now(),
		CommandID:   "cmd-001",
		Source:      "mqtt",
		CommandType: "set_setpoint",
		LoopName:    "pressure",
		PayloadJSON: `{"command":"set_setpoint","loop":"pressure","value":7.5}`,
		Status:      "applied",
	}
	if err := store.InsertCommand(ctx, cmd); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}
}

func TestInsertCommand_Rejected(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	cmd := CommandRecord{
		Timestamp:    time.Now(),
		CommandID:    "cmd-002",
		Source:       "mqtt",
		CommandType:  "set_setpoint",
		LoopName:     "unknown",
		PayloadJSON:  `{"command":"set_setpoint","loop":"unknown","value":7.5}`,
		Status:       "rejected",
		ErrorMessage: "unknown loop",
	}
	if err := store.InsertCommand(ctx, cmd); err != nil {
		t.Fatalf("InsertCommand rejected: %v", err)
	}
}

func TestInsertCommand_InvalidZeroTimestamp(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := CommandRecord{Source: "mqtt", Status: "applied"}
	if err := store.InsertCommand(ctx, bad); err == nil {
		t.Error("zero timestamp should be rejected")
	}
}

func TestInsertCommand_InvalidEmptyStatus(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	bad := CommandRecord{Timestamp: time.Now(), Source: "mqtt", Status: ""}
	if err := store.InsertCommand(ctx, bad); err == nil {
		t.Error("empty status should be rejected")
	}
}
