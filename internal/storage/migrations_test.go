package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestMigrate_Idempotent(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := migrate(ctx, db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := migrate(ctx, db); err != nil {
		t.Fatalf("second migration (idempotent): %v", err)
	}
}

func TestMigrate_TablesExist(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	if err := migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tables := []string{
		"schema_migrations",
		"telemetry_samples",
		"events",
		"commands",
		"pid_changes",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRowContext(ctx,
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil || name != table {
			t.Errorf("table %q not found after migration", table)
		}
	}
}

func TestMigrate_VersionsRecorded(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	if err := migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("expected %d migration records, got %d", len(migrations), count)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          filepath.Join(dir, "test.db"),
		EventsJSONLPath:     filepath.Join(dir, "events.jsonl"),
		AppLogPath:          filepath.Join(dir, "app.log"),
		RetentionMaxSamples: 1000,
		WriteQueueSize:      64,
	}
	ctx := context.Background()
	store, err := Open(ctx, cfg)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestOpen_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		Enabled:             true,
		Type:                "sqlite",
		SQLitePath:          filepath.Join(dir, "sub", "history.db"),
		EventsJSONLPath:     filepath.Join(dir, "events.jsonl"),
		RetentionMaxSamples: 100,
		WriteQueueSize:      16,
	}
	store, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_ = store.Close()
	if _, err := os.Stat(filepath.Join(dir, "sub", "history.db")); err != nil {
		t.Error("db file not created")
	}
}
