package source

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"gheymatchi/backend/internal/auth"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(ctx context.Context, productID string, input CreateInput) (ProductSource, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return ProductSource{}, err
	}

	normalized, err := NormalizeCreate(input)
	if err != nil {
		return ProductSource{}, err
	}

	now := time.Now().UTC()
	isActive := true
	if normalized.IsActive != nil {
		isActive = *normalized.IsActive
	}

	productSource := ProductSource{
		ID:         newID(),
		ProductID:  productID,
		URL:        normalized.URL,
		SourceName: normalized.SourceName,
		IsActive:   isActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO product_sources (id, product_id, url, source_name, is_active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		productSource.ID,
		productSource.ProductID,
		productSource.URL,
		nullString(productSource.SourceName),
		boolInt(productSource.IsActive),
		formatTime(productSource.CreatedAt),
		formatTime(productSource.UpdatedAt),
	)
	if err != nil {
		return ProductSource{}, fmt.Errorf("create product source: %w", err)
	}

	return productSource, nil
}

func (s *SQLiteStore) List(ctx context.Context, productID string) ([]ProductSource, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, product_id, url, source_name, is_active, created_at, updated_at
FROM product_sources
WHERE product_id = ?
ORDER BY created_at DESC, id DESC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list product sources: %w", err)
	}
	defer rows.Close()

	var sources []ProductSource
	for rows.Next() {
		productSource, err := scanProductSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, productSource)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list product sources rows: %w", err)
	}

	return sources, nil
}

func (s *SQLiteStore) ListActive(ctx context.Context) ([]ProductSource, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, product_id, url, source_name, is_active, created_at, updated_at
FROM product_sources
WHERE is_active = 1
ORDER BY updated_at ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list active product sources: %w", err)
	}
	defer rows.Close()

	var sources []ProductSource
	for rows.Next() {
		productSource, err := scanProductSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, productSource)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list active product sources rows: %w", err)
	}

	return sources, nil
}

func (s *SQLiteStore) Update(ctx context.Context, productID string, sourceID string, input UpdateInput) (ProductSource, error) {
	normalized, err := NormalizeUpdate(input)
	if err != nil {
		return ProductSource{}, err
	}

	existing, err := s.get(ctx, productID, sourceID)
	if err != nil {
		return ProductSource{}, err
	}

	if normalized.URL != nil {
		existing.URL = *normalized.URL
	}
	if normalized.SourceName != nil {
		existing.SourceName = *normalized.SourceName
	}
	if normalized.IsActive != nil {
		existing.IsActive = *normalized.IsActive
	}
	existing.UpdatedAt = time.Now().UTC()

	result, err := s.db.ExecContext(
		ctx,
		`UPDATE product_sources
SET url = ?, source_name = ?, is_active = ?, updated_at = ?
WHERE product_id = ? AND id = ?`,
		existing.URL,
		nullString(existing.SourceName),
		boolInt(existing.IsActive),
		formatTime(existing.UpdatedAt),
		productID,
		sourceID,
	)
	if err != nil {
		return ProductSource{}, fmt.Errorf("update product source: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return ProductSource{}, fmt.Errorf("check product source update: %w", err)
	}
	if affected == 0 {
		return ProductSource{}, ErrNotFound
	}

	return existing, nil
}

func (s *SQLiteStore) Delete(ctx context.Context, productID string, sourceID string) error {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM product_sources WHERE product_id = ? AND id = ?`, productID, sourceID)
	if err != nil {
		return fmt.Errorf("delete product source: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check product source delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) get(ctx context.Context, productID string, sourceID string) (ProductSource, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return ProductSource{}, err
	}

	row := s.db.QueryRowContext(ctx, `
SELECT id, product_id, url, source_name, is_active, created_at, updated_at
FROM product_sources
WHERE product_id = ? AND id = ?`, productID, sourceID)

	productSource, err := scanProductSource(row)
	if err != nil {
		return ProductSource{}, err
	}
	return productSource, nil
}

func (s *SQLiteStore) ensureProductExists(ctx context.Context, productID string) error {
	var id string
	userID, ok := auth.UserIDFromContext(ctx)
	if ok {
		err := s.db.QueryRowContext(ctx, `SELECT id FROM products WHERE id = ? AND user_id = ?`, productID, userID).Scan(&id)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check product exists: %w", err)
		}
		return nil
	}

	err := s.db.QueryRowContext(ctx, `SELECT id FROM products WHERE id = ?`, productID).Scan(&id)
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

func scanProductSource(row scanner) (ProductSource, error) {
	var productSource ProductSource
	var sourceName sql.NullString
	var isActive int
	var createdAt string
	var updatedAt string

	err := row.Scan(
		&productSource.ID,
		&productSource.ProductID,
		&productSource.URL,
		&sourceName,
		&isActive,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return ProductSource{}, ErrNotFound
	}
	if err != nil {
		return ProductSource{}, fmt.Errorf("scan product source: %w", err)
	}

	productSource.SourceName = "unknown"
	if sourceName.Valid && sourceName.String != "" {
		productSource.SourceName = sourceName.String
	}
	productSource.IsActive = isActive != 0

	productSource.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return ProductSource{}, err
	}
	productSource.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return ProductSource{}, err
	}

	return productSource, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse product source time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate product source id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}
