package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "create_schema_migrations",
		sql: `CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       TEXT    NOT NULL,
			applied_at TEXT    NOT NULL
		);`,
	},
	{
		version: 2,
		name:    "create_telemetry_samples",
		sql: `CREATE TABLE IF NOT EXISTS telemetry_samples (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp    TEXT    NOT NULL,
			device_id    TEXT    NOT NULL,
			scan_counter INTEGER NOT NULL,
			loop_name    TEXT    NOT NULL,
			sp           REAL    NOT NULL,
			pv           REAL    NOT NULL,
			mv           REAL    NOT NULL,
			error        REAL    NOT NULL,
			mode         TEXT    NOT NULL,
			quality      TEXT    NOT NULL,
			unit         TEXT    NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp       ON telemetry_samples(timestamp);
		CREATE INDEX IF NOT EXISTS idx_telemetry_loop_timestamp  ON telemetry_samples(loop_name, timestamp);`,
	},
	{
		version: 3,
		name:    "create_events",
		sql: `CREATE TABLE IF NOT EXISTS events (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp    TEXT    NOT NULL,
			level        TEXT    NOT NULL,
			event_type   TEXT    NOT NULL,
			message      TEXT    NOT NULL,
			details_json TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_events_timestamp       ON events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_events_type_timestamp  ON events(event_type, timestamp);`,
	},
	{
		version: 4,
		name:    "create_commands",
		sql: `CREATE TABLE IF NOT EXISTS commands (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     TEXT    NOT NULL,
			command_id    TEXT,
			source        TEXT    NOT NULL,
			command_type  TEXT    NOT NULL,
			loop_name     TEXT,
			payload_json  TEXT    NOT NULL,
			status        TEXT    NOT NULL,
			error_message TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_commands_timestamp        ON commands(timestamp);
		CREATE INDEX IF NOT EXISTS idx_commands_status_timestamp ON commands(status, timestamp);`,
	},
	{
		version: 5,
		name:    "create_pid_changes",
		sql: `CREATE TABLE IF NOT EXISTS pid_changes (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  TEXT    NOT NULL,
			loop_name  TEXT    NOT NULL,
			old_kp     REAL,
			old_ki     REAL,
			old_kd     REAL,
			new_kp     REAL    NOT NULL,
			new_ki     REAL    NOT NULL,
			new_kd     REAL    NOT NULL,
			source     TEXT    NOT NULL,
			command_id TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_pid_changes_timestamp       ON pid_changes(timestamp);
		CREATE INDEX IF NOT EXISTS idx_pid_changes_loop_timestamp  ON pid_changes(loop_name, timestamp);`,
	},
}

// migrate runs all pending migrations in version order. It is idempotent.
func migrate(ctx context.Context, db *sql.DB) error {
	// Ensure the migrations table exists first (migration 1).
	if _, err := db.ExecContext(ctx, migrations[0].sql); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	if err := recordMigration(ctx, db, migrations[0]); err != nil {
		return err
	}

	for _, m := range migrations[1:] {
		applied, err := isMigrationApplied(ctx, db, m.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if _, err := db.ExecContext(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %d %q: %w", m.version, m.name, err)
		}
		if err := recordMigration(ctx, db, m); err != nil {
			return err
		}
	}
	return nil
}

func isMigrationApplied(ctx context.Context, db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check migration %d: %w", version, err)
	}
	return count > 0, nil
}

func recordMigration(ctx context.Context, db *sql.DB, m migration) error {
	applied, err := isMigrationApplied(ctx, db, m.version)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
		m.version, m.name, time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("record migration %d: %w", m.version, err)
	}
	return nil
}
