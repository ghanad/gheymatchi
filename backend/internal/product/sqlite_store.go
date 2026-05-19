package product

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

func (s *SQLiteStore) Create(ctx context.Context, userID string, input CreateInput) (Product, error) {
	normalized, err := NormalizeCreate(input)
	if err != nil {
		return Product{}, err
	}

	now := time.Now().UTC()
	product := Product{
		ID:          newID(),
		UserID:      &userID,
		Name:        normalized.Name,
		Description: normalized.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO products (id, user_id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		product.ID,
		userID,
		product.Name,
		nullString(product.Description),
		formatTime(product.CreatedAt),
		formatTime(product.UpdatedAt),
	)
	if err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}

	return product, nil
}

func (s *SQLiteStore) List(ctx context.Context) ([]Product, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, user_id, name, description, created_at, updated_at
FROM products
WHERE user_id = ?
ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list products rows: %w", err)
	}

	return products, nil
}

func (s *SQLiteStore) Get(ctx context.Context, id string) (Product, error) {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return Product{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, user_id, name, description, created_at, updated_at
FROM products
WHERE id = ? AND user_id = ?`, id, userID)

	product, err := scanProduct(row)
	if err != nil {
		return Product{}, err
	}
	return product, nil
}

func (s *SQLiteStore) Update(ctx context.Context, id string, input UpdateInput) (Product, error) {
	normalized, err := NormalizeUpdate(input)
	if err != nil {
		return Product{}, err
	}

	existing, err := s.Get(ctx, id)
	if err != nil {
		return Product{}, err
	}

	if normalized.Name != nil {
		existing.Name = *normalized.Name
	}
	if normalized.Description != nil {
		existing.Description = *normalized.Description
	}
	existing.UpdatedAt = time.Now().UTC()

	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return Product{}, err
	}
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE products SET name = ?, description = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		existing.Name,
		nullString(existing.Description),
		formatTime(existing.UpdatedAt),
		id,
		userID,
	)
	if err != nil {
		return Product{}, fmt.Errorf("update product: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Product{}, fmt.Errorf("check product update: %w", err)
	}
	if affected == 0 {
		return Product{}, ErrNotFound
	}

	return existing, nil
}

func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM products WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check product delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(row scanner) (Product, error) {
	var product Product
	var userID sql.NullString
	var description sql.NullString
	var createdAt string
	var updatedAt string

	err := row.Scan(&product.ID, &userID, &product.Name, &description, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("scan product: %w", err)
	}

	if userID.Valid {
		product.UserID = &userID.String
	}
	if description.Valid {
		product.Description = description.String
	}

	product.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Product{}, err
	}
	product.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Product{}, err
	}

	return product, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse product time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate product id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}
