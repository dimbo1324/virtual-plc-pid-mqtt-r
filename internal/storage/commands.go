package storage

import (
	"context"
	"fmt"
	"time"
)

// InsertCommand persists a command record (applied or rejected).
func (s *Store) InsertCommand(ctx context.Context, command CommandRecord) error {
	if err := validateCommandRecord(command); err != nil {
		return err
	}
	payload := command.PayloadJSON
	if payload == "" {
		payload = "{}"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO commands
			(timestamp, command_id, source, command_type, loop_name, payload_json, status, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		command.Timestamp.UTC().Format(time.RFC3339Nano),
		command.CommandID, command.Source, command.CommandType,
		nullStr(command.LoopName), payload,
		command.Status, nullStr(command.ErrorMessage),
	)
	if err != nil {
		return fmt.Errorf("insert command: %w", err)
	}
	return nil
}

// nullStr converts an empty string to nil for nullable DB columns.
func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
