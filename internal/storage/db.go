package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

// Store wraps a SQLite database and exposes all repository methods.
type Store struct {
	db  *sql.DB
	cfg Config
}

// Open validates cfg, creates parent directories, opens the SQLite database,
// and runs all migrations. On failure the caller should not call Close.
func Open(ctx context.Context, cfg Config) (*Store, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("storage config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o750); err != nil {
		return nil, fmt.Errorf("create storage directory %q: %w", filepath.Dir(cfg.SQLitePath), err)
	}

	db, err := sql.Open("sqlite", cfg.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", cfg.SQLitePath, err)
	}

	// Serialise writes; SQLite is single-writer.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite %q: %w", cfg.SQLitePath, err)
	}

	// Reasonable SQLite pragmas for a local, non-concurrent workload.
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set pragma %q: %w", p, err)
		}
	}

	s := &Store{db: db, cfg: cfg}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate storage: %w", err)
	}
	return s, nil
}

// Close releases the database connection.
// It checkpoints the WAL and switches journal mode to DELETE so the -wal and
// -shm files are merged and removed before the handle is closed. This prevents
// "directory not empty" errors when the caller removes the database directory
// (common in tests on Windows).
func (s *Store) Close() error {
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	_, _ = s.db.Exec("PRAGMA journal_mode=DELETE;")
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close storage db: %w", err)
	}
	return nil
}
