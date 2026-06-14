package storage

import (
	"context"
	"fmt"
)

// ApplyRetention deletes telemetry rows older than the configured
// retention_max_samples, keeping the newest N rows.
func (s *Store) ApplyRetention(ctx context.Context) error {
	maxSamples := s.cfg.RetentionMaxSamples
	if maxSamples <= 0 {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM telemetry_samples
		WHERE id NOT IN (
			SELECT id FROM telemetry_samples
			ORDER BY id DESC
			LIMIT ?
		)`, maxSamples)
	if err != nil {
		return fmt.Errorf("apply telemetry retention: %w", err)
	}
	return nil
}
