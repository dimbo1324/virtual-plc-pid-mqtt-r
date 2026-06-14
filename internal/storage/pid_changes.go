package storage

import (
	"context"
	"fmt"
	"time"
)

// InsertPIDChange persists a PID tuning change.
func (s *Store) InsertPIDChange(ctx context.Context, change PIDChangeRecord) error {
	if change.Timestamp.IsZero() {
		return fmt.Errorf("%w: pid change timestamp is zero", ErrInvalidRecord)
	}
	if change.LoopName == "" {
		return fmt.Errorf("%w: pid change loop_name is empty", ErrInvalidRecord)
	}
	if change.Source == "" {
		return fmt.Errorf("%w: pid change source is empty", ErrInvalidRecord)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO pid_changes
			(timestamp, loop_name, old_kp, old_ki, old_kd, new_kp, new_ki, new_kd, source, command_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		change.Timestamp.UTC().Format(time.RFC3339Nano),
		change.LoopName,
		change.OldKp, change.OldKi, change.OldKd,
		change.NewKp, change.NewKi, change.NewKd,
		change.Source, nullStr(change.CommandID),
	)
	if err != nil {
		return fmt.Errorf("insert pid change: %w", err)
	}
	return nil
}
