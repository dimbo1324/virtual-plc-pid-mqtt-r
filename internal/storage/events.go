package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// InsertEvent persists a single event record.
func (s *Store) InsertEvent(ctx context.Context, event EventRecord) error {
	if err := validateEventRecord(event); err != nil {
		return err
	}
	detailsJSON, err := marshalDetails(event.Details)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO events (timestamp, level, event_type, message, details_json)
		VALUES (?, ?, ?, ?, ?)`,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.Level, event.Type, event.Message, detailsJSON,
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

// RecentEvents returns the newest limit events, ordered newest-first.
func (s *Store) RecentEvents(ctx context.Context, limit int) ([]EventRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT timestamp, level, event_type, message, details_json
		FROM events
		ORDER BY id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("recent events: %w", err)
	}
	defer rows.Close()

	var events []EventRecord
	for rows.Next() {
		var e EventRecord
		var ts, detailsStr string
		if err := rows.Scan(&ts, &e.Level, &e.Type, &e.Message, &detailsStr); err != nil {
			return nil, fmt.Errorf("scan event row: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, fmt.Errorf("parse event timestamp: %w", err)
		}
		e.Timestamp = t
		if detailsStr != "" && detailsStr != "{}" {
			if err := json.Unmarshal([]byte(detailsStr), &e.Details); err != nil {
				return nil, fmt.Errorf("parse event details: %w", err)
			}
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
