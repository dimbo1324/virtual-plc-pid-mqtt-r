package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONLWriter_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "events.jsonl")

	w, err := NewJSONLWriter(path)
	if err != nil {
		t.Fatalf("NewJSONLWriter: %v", err)
	}
	defer w.Close()

	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Error("parent directory not created")
	}
}

func TestJSONLWriter_WritesValidJSONLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	w, err := NewJSONLWriter(path)
	if err != nil {
		t.Fatalf("NewJSONLWriter: %v", err)
	}

	event := EventRecord{
		Timestamp: time.Now(),
		Level:     "info",
		Type:      "plc_started",
		Message:   "test",
	}
	if err := w.WriteEvent(context.Background(), event); err != nil {
		t.Fatalf("WriteEvent: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	f, _ := os.Open(path)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("no lines written")
	}
	var obj map[string]any
	if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
		t.Errorf("line is not valid JSON: %v", err)
	}
	if obj["event_type"] != "plc_started" {
		t.Errorf("unexpected event_type: %v", obj["event_type"])
	}
}

func TestJSONLWriter_RepeatedWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	w, err := NewJSONLWriter(path)
	if err != nil {
		t.Fatalf("NewJSONLWriter: %v", err)
	}
	defer w.Close()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		e := EventRecord{
			Timestamp: time.Now(),
			Level:     "info",
			Type:      "test_event",
			Message:   "test",
		}
		if err := w.WriteEvent(ctx, e); err != nil {
			t.Fatalf("WriteEvent %d: %v", i, err)
		}
	}

	f, _ := os.Open(path)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if lines != 5 {
		t.Errorf("expected 5 lines, got %d", lines)
	}
}
