package notification

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"gheymatchi/backend/internal/alert"
	"gheymatchi/backend/internal/price"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) CreateAlertTriggered(ctx context.Context, alert alert.Alert, pricePoint price.PricePoint) error {
	now := time.Now().UTC()
	recipient := alert.ProductID
	if alert.UserID != nil {
		recipient = *alert.UserID
	}

	_, err := s.db.ExecContext(ctx, `INSERT INTO notifications
(id, alert_id, channel, recipient, status, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		newID(),
		alert.ID,
		ChannelInternal,
		recipient,
		StatusPending,
		formatTime(now),
	)
	if err != nil {
		return fmt.Errorf("create alert notification for price point %s: %w", pricePoint.ID, err)
	}
	return nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate notification id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}
