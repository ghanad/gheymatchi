package notification

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"gheymatchi/backend/internal/alert"
	"gheymatchi/backend/internal/db"
	"gheymatchi/backend/internal/price"
)

type SQLiteStore struct {
	db               *sql.DB
	driver           db.Driver
	defaultChannel   string
	defaultRecipient string
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return NewStore(db, "sqlite")
}

func NewStore(database *sql.DB, driver db.Driver) *SQLiteStore {
	return &SQLiteStore{db: database, driver: driver, defaultChannel: ChannelDryRun}
}

func NewSQLiteStoreWithDefaults(db *sql.DB, channel string, recipient string) *SQLiteStore {
	return NewStoreWithDefaults(db, "sqlite", channel, recipient)
}

func NewStoreWithDefaults(database *sql.DB, driver db.Driver, channel string, recipient string) *SQLiteStore {
	if channel == "" {
		channel = ChannelDryRun
	}
	return &SQLiteStore{db: database, driver: driver, defaultChannel: channel, defaultRecipient: recipient}
}

func (s *SQLiteStore) CreateAlertTriggered(ctx context.Context, alert alert.Alert, pricePoint price.PricePoint) error {
	now := time.Now().UTC()
	recipient := s.defaultRecipient
	if recipient == "" && alert.UserID != nil {
		recipient = *alert.UserID
	}
	if recipient == "" {
		recipient = alert.ProductID
	}

	_, err := s.db.ExecContext(ctx, s.rebind(`INSERT INTO notifications
(id, alert_id, channel, recipient, status, created_at)
VALUES (?, ?, ?, ?, ?, ?)`),
		newID(),
		alert.ID,
		s.defaultChannel,
		recipient,
		StatusPending,
		formatTime(now),
	)
	if err != nil {
		return fmt.Errorf("create alert notification for price point %s: %w", pricePoint.ID, err)
	}
	return nil
}

func (s *SQLiteStore) List(ctx context.Context) ([]Notification, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, alert_id, channel, recipient, status, attempt_count, last_error, sent_at, created_at
FROM notifications
ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (s *SQLiteStore) ListForUser(ctx context.Context, userID string) ([]Notification, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
SELECT notifications.id, notifications.alert_id, notifications.channel, notifications.recipient, notifications.status, notifications.attempt_count, notifications.last_error, notifications.sent_at, notifications.created_at
FROM notifications
JOIN alerts ON alerts.id = notifications.alert_id
WHERE alerts.user_id = ?
ORDER BY notifications.created_at DESC, notifications.id DESC`), userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (s *SQLiteStore) ListPending(ctx context.Context, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, s.rebind(`
SELECT id, alert_id, channel, recipient, status, attempt_count, last_error, sent_at, created_at
FROM notifications
WHERE status = ?
ORDER BY created_at ASC, id ASC
LIMIT ?`), StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (s *SQLiteStore) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	result, err := s.db.ExecContext(ctx, s.rebind(`
UPDATE notifications
SET status = ?, sent_at = ?, last_error = NULL
WHERE id = ?`), StatusSent, formatTime(sentAt), id)
	if err != nil {
		return fmt.Errorf("mark notification sent: %w", err)
	}
	return checkAffected(result, "sent")
}

func (s *SQLiteStore) RecordFailedAttempt(ctx context.Context, id string, message string, maxAttempts int) error {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	result, err := s.db.ExecContext(ctx, s.rebind(`
UPDATE notifications
SET attempt_count = attempt_count + 1,
	last_error = ?,
	status = CASE WHEN attempt_count + 1 >= ? THEN ? ELSE ? END
WHERE id = ?`), message, maxAttempts, StatusFailed, StatusPending, id)
	if err != nil {
		return fmt.Errorf("record notification failed attempt: %w", err)
	}
	return checkAffected(result, "failed")
}

type rowScanner interface {
	Scan(dest ...any) error
}

type rowsScanner interface {
	rowScanner
	Next() bool
	Err() error
}

func scanNotifications(rows rowsScanner) ([]Notification, error) {
	var notifications []Notification
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan notification rows: %w", err)
	}
	return notifications, nil
}

func scanNotification(row rowScanner) (Notification, error) {
	var notification Notification
	var alertID sql.NullString
	var lastError sql.NullString
	var sentAt sql.NullString
	var createdAt string

	err := row.Scan(
		&notification.ID,
		&alertID,
		&notification.Channel,
		&notification.Recipient,
		&notification.Status,
		&notification.AttemptCount,
		&lastError,
		&sentAt,
		&createdAt,
	)
	if err != nil {
		return Notification{}, fmt.Errorf("scan notification: %w", err)
	}

	if alertID.Valid {
		notification.AlertID = &alertID.String
	}
	if lastError.Valid {
		notification.LastError = &lastError.String
	}
	if sentAt.Valid {
		parsed, err := parseTime(sentAt.String)
		if err != nil {
			return Notification{}, err
		}
		notification.SentAt = &parsed
	}

	parsed, err := parseTime(createdAt)
	if err != nil {
		return Notification{}, err
	}
	notification.CreatedAt = parsed

	return notification, nil
}

func checkAffected(result sql.Result, action string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check notification %s update: %w", action, err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse notification time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate notification id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}

func (s *SQLiteStore) rebind(query string) string {
	return db.Rebind(s.driver, query)
}
