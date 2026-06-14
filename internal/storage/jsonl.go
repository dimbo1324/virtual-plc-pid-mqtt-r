package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONLWriter appends events as JSON Lines to a file. It is safe for
// concurrent use. Log rotation is not implemented in this stage; the file
// grows until cleaned manually or by an OS log-rotation tool.
type JSONLWriter struct {
	mu   sync.Mutex
	file *os.File
}

// NewJSONLWriter opens or creates path in append mode, creating any parent
// directories that do not exist.
func NewJSONLWriter(path string) (*JSONLWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create jsonl directory %q: %w", filepath.Dir(path), err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("open jsonl file %q: %w", path, err)
	}
	return &JSONLWriter{file: f}, nil
}

type jsonlLine struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	EventType string         `json:"event_type"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
}

// WriteEvent appends one JSON line. Errors are returned but must not crash the
// caller; the PLC runtime continues regardless.
func (w *JSONLWriter) WriteEvent(_ context.Context, event EventRecord) error {
	line := jsonlLine{
		Timestamp: event.Timestamp.UTC().Format(time.RFC3339Nano),
		Level:     event.Level,
		EventType: event.Type,
		Message:   event.Message,
		Details:   event.Details,
	}
	b, err := json.Marshal(line)
	if err != nil {
		return fmt.Errorf("marshal jsonl event: %w", err)
	}
	b = append(b, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.file.Write(b); err != nil {
		return fmt.Errorf("write jsonl event: %w", err)
	}
	return nil
}

// Close flushes and closes the underlying file.
func (w *JSONLWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close jsonl file: %w", err)
	}
	return nil
}
