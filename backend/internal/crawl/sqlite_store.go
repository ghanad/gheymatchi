package crawl

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"gheymatchi/backend/internal/db"
)

type SQLiteStore struct {
	db     *sql.DB
	driver db.Driver
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return NewStore(db, "sqlite")
}

func NewStore(database *sql.DB, driver db.Driver) *SQLiteStore {
	return &SQLiteStore{db: database, driver: driver}
}

func (s *SQLiteStore) Start(ctx context.Context, sourceID string) (Run, error) {
	now := time.Now().UTC()
	run := Run{
		ID:        newID(),
		SourceID:  sourceID,
		Status:    StatusRunning,
		StartedAt: now,
		CreatedAt: now,
	}

	_, err := s.db.ExecContext(ctx, s.rebind(`
INSERT INTO crawl_runs (id, source_id, status, started_at, created_at)
VALUES (?, ?, ?, ?, ?)`),
		run.ID,
		run.SourceID,
		run.Status,
		formatTime(run.StartedAt),
		formatTime(run.CreatedAt),
	)
	if err != nil {
		return Run{}, fmt.Errorf("start crawl run: %w", err)
	}

	return run, nil
}

func (s *SQLiteStore) Finish(ctx context.Context, id string, status string, errorMessage *string) error {
	finishedAt := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, s.rebind(`
UPDATE crawl_runs
SET status = ?, finished_at = ?, error_message = ?
WHERE id = ?`),
		status,
		formatTime(finishedAt),
		nullStringPtr(errorMessage),
		id,
	)
	if err != nil {
		return fmt.Errorf("finish crawl run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check crawl run finish: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("crawl run %s not found", id)
	}
	return nil
}

func nullStringPtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate crawl run id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}

func (s *SQLiteStore) rebind(query string) string {
	return db.Rebind(s.driver, query)
}
