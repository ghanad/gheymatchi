package alert

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"gheymatchi/backend/internal/auth"
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

func (s *SQLiteStore) Create(ctx context.Context, productID string, input CreateInput) (Alert, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return Alert{}, err
	}

	normalized, err := NormalizeCreate(input)
	if err != nil {
		return Alert{}, err
	}

	now := time.Now().UTC()
	isActive := true
	if normalized.IsActive != nil {
		isActive = *normalized.IsActive
	}

	alert := Alert{
		ID:                 newID(),
		ProductID:          productID,
		Name:               normalized.Name,
		ConditionType:      normalized.ConditionType,
		TargetUnit:         normalized.TargetUnit,
		ThresholdValueText: normalized.ThresholdValueText,
		IsActive:           isActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if userID, ok := auth.UserIDFromContext(ctx); ok {
		alert.UserID = &userID
	}

	_, err = s.db.ExecContext(ctx, s.rebind(`INSERT INTO alerts
(id, user_id, product_id, name, condition_type, target_unit, threshold_value_text, is_active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		alert.ID,
		nullStringPtr(alert.UserID),
		alert.ProductID,
		alert.Name,
		alert.ConditionType,
		alert.TargetUnit,
		alert.ThresholdValueText,
		boolInt(alert.IsActive),
		formatTime(alert.CreatedAt),
		formatTime(alert.UpdatedAt),
	)
	if err != nil {
		return Alert{}, fmt.Errorf("create alert: %w", err)
	}

	return alert, nil
}

func (s *SQLiteStore) List(ctx context.Context, productID string) ([]Alert, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, s.rebind(`
SELECT id, user_id, product_id, name, condition_type, target_unit, threshold_value_text, is_active, last_triggered_at, created_at, updated_at
FROM alerts
WHERE product_id = ?
ORDER BY created_at DESC, id DESC`), productID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list alerts rows: %w", err)
	}

	return alerts, nil
}

func (s *SQLiteStore) Update(ctx context.Context, productID string, alertID string, input UpdateInput) (Alert, error) {
	normalized, err := NormalizeUpdate(input)
	if err != nil {
		return Alert{}, err
	}

	existing, err := s.get(ctx, productID, alertID)
	if err != nil {
		return Alert{}, err
	}

	if normalized.Name != nil {
		existing.Name = *normalized.Name
	}
	if normalized.ConditionType != nil {
		existing.ConditionType = *normalized.ConditionType
	}
	if normalized.TargetUnit != nil {
		existing.TargetUnit = *normalized.TargetUnit
	}
	if normalized.ThresholdValueText != nil {
		existing.ThresholdValueText = *normalized.ThresholdValueText
	}
	criteriaChanged := normalized.ConditionType != nil || normalized.TargetUnit != nil || normalized.ThresholdValueText != nil
	if criteriaChanged {
		existing.LastTriggeredAt = nil
	}
	if normalized.IsActive != nil {
		existing.IsActive = *normalized.IsActive
	}
	existing.UpdatedAt = time.Now().UTC()

	result, err := s.db.ExecContext(ctx, s.rebind(`UPDATE alerts
SET name = ?, condition_type = ?, target_unit = ?, threshold_value_text = ?, is_active = ?, last_triggered_at = ?, updated_at = ?
WHERE product_id = ? AND id = ?`),
		existing.Name,
		existing.ConditionType,
		existing.TargetUnit,
		existing.ThresholdValueText,
		boolInt(existing.IsActive),
		nullTimePtr(existing.LastTriggeredAt),
		formatTime(existing.UpdatedAt),
		productID,
		alertID,
	)
	if err != nil {
		return Alert{}, fmt.Errorf("update alert: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Alert{}, fmt.Errorf("check alert update: %w", err)
	}
	if affected == 0 {
		return Alert{}, ErrNotFound
	}

	return existing, nil
}

func (s *SQLiteStore) Delete(ctx context.Context, productID string, alertID string) error {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, s.rebind(`DELETE FROM alerts WHERE product_id = ? AND id = ?`), productID, alertID)
	if err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check alert delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) ListActiveByProduct(ctx context.Context, productID string) ([]Alert, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
SELECT id, user_id, product_id, name, condition_type, target_unit, threshold_value_text, is_active, last_triggered_at, created_at, updated_at
FROM alerts
WHERE product_id = ? AND is_active = 1
ORDER BY created_at ASC, id ASC`), productID)
	if err != nil {
		return nil, fmt.Errorf("list active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list active alerts rows: %w", err)
	}

	return alerts, nil
}

func (s *SQLiteStore) MarkTriggered(ctx context.Context, alertID string, triggeredAt time.Time) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, s.rebind(`
UPDATE alerts
SET last_triggered_at = ?, updated_at = ?
WHERE id = ?`),
		formatTime(triggeredAt),
		formatTime(now),
		alertID,
	)
	if err != nil {
		return fmt.Errorf("mark alert triggered: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check alert trigger update: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) get(ctx context.Context, productID string, alertID string) (Alert, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return Alert{}, err
	}

	row := s.db.QueryRowContext(ctx, s.rebind(`
SELECT id, user_id, product_id, name, condition_type, target_unit, threshold_value_text, is_active, last_triggered_at, created_at, updated_at
FROM alerts
WHERE product_id = ? AND id = ?`), productID, alertID)

	alert, err := scanAlert(row)
	if err != nil {
		return Alert{}, err
	}
	return alert, nil
}

func (s *SQLiteStore) ensureProductExists(ctx context.Context, productID string) error {
	var id string
	userID, ok := auth.UserIDFromContext(ctx)
	if ok {
		err := s.db.QueryRowContext(ctx, s.rebind(`SELECT id FROM products WHERE id = ? AND user_id = ?`), productID, userID).Scan(&id)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check product exists: %w", err)
		}
		return nil
	}

	err := s.db.QueryRowContext(ctx, s.rebind(`SELECT id FROM products WHERE id = ?`), productID).Scan(&id)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("check product exists: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAlert(row scanner) (Alert, error) {
	var alert Alert
	var userID sql.NullString
	var isActive int
	var lastTriggeredAt sql.NullString
	var createdAt string
	var updatedAt string

	err := row.Scan(
		&alert.ID,
		&userID,
		&alert.ProductID,
		&alert.Name,
		&alert.ConditionType,
		&alert.TargetUnit,
		&alert.ThresholdValueText,
		&isActive,
		&lastTriggeredAt,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return Alert{}, ErrNotFound
	}
	if err != nil {
		return Alert{}, fmt.Errorf("scan alert: %w", err)
	}

	if userID.Valid {
		alert.UserID = &userID.String
	}
	alert.IsActive = isActive != 0
	if lastTriggeredAt.Valid {
		parsed, err := parseTime(lastTriggeredAt.String)
		if err != nil {
			return Alert{}, err
		}
		alert.LastTriggeredAt = &parsed
	}

	alert.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Alert{}, err
	}
	alert.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Alert{}, err
	}

	return alert, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullTimePtr(value *time.Time) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: formatTime(*value), Valid: true}
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

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse alert time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate alert id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}

func (s *SQLiteStore) rebind(query string) string {
	return db.Rebind(s.driver, query)
}
