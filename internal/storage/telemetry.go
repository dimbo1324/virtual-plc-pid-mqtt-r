package storage

import (
	"context"
	"fmt"
	"time"
)

// InsertTelemetrySamples persists a batch of samples in a single transaction.
// An empty batch is a no-op (returns nil).
func (s *Store) InsertTelemetrySamples(ctx context.Context, samples []TelemetrySample) error {
	if len(samples) == 0 {
		return nil
	}
	for i, sample := range samples {
		if err := validateTelemetrySample(sample); err != nil {
			return fmt.Errorf("telemetry sample[%d]: %w", i, err)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("insert telemetry: begin tx: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO telemetry_samples
			(timestamp, device_id, scan_counter, loop_name, sp, pv, mv, error, mode, quality, unit)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("insert telemetry: prepare: %w", err)
	}
	defer stmt.Close()

	for _, s := range samples {
		_, err := stmt.ExecContext(ctx,
			s.Timestamp.UTC().Format(time.RFC3339Nano),
			s.DeviceID, s.ScanCounter, s.LoopName,
			s.SP, s.PV, s.MV, s.Error, s.Mode, s.Quality, s.Unit,
		)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert telemetry row: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("insert telemetry: commit: %w", err)
	}
	return nil
}

// RecentTelemetry returns the newest limit rows for the given loop, ordered
// newest-first.
func (s *Store) RecentTelemetry(ctx context.Context, loopName string, limit int) ([]TelemetrySample, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT timestamp, device_id, scan_counter, loop_name, sp, pv, mv, error, mode, quality, unit
		FROM telemetry_samples
		WHERE loop_name = ?
		ORDER BY id DESC
		LIMIT ?`, loopName, limit)
	if err != nil {
		return nil, fmt.Errorf("recent telemetry: %w", err)
	}
	defer rows.Close()

	var samples []TelemetrySample
	for rows.Next() {
		var s TelemetrySample
		var ts string
		if err := rows.Scan(&ts, &s.DeviceID, &s.ScanCounter, &s.LoopName,
			&s.SP, &s.PV, &s.MV, &s.Error, &s.Mode, &s.Quality, &s.Unit); err != nil {
			return nil, fmt.Errorf("scan telemetry row: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, fmt.Errorf("parse telemetry timestamp: %w", err)
		}
		s.Timestamp = t
		samples = append(samples, s)
	}
	return samples, rows.Err()
}
