package price

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"gheymatchi/backend/internal/marketrate"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(ctx context.Context, productID string, productSourceID string, input CreateInput) (PricePoint, error) {
	if err := s.ensureProductSource(ctx, productID, productSourceID); err != nil {
		return PricePoint{}, err
	}

	normalized, err := NormalizeCreate(input)
	if err != nil {
		return PricePoint{}, err
	}

	now := time.Now().UTC()
	pricePoint := PricePoint{
		ID:              newID(),
		ProductID:       productID,
		ProductSourceID: productSourceID,
		PriceIRR:        normalized.PriceIRR,
		CapturedAt:      normalized.CapturedAt,
		RawPayload:      normalized.RawPayload,
		CreatedAt:       now,
	}

	if err := s.attachDerivedPrices(ctx, &pricePoint); err != nil {
		return PricePoint{}, err
	}

	_, err = s.db.ExecContext(ctx, `INSERT INTO price_points
(id, product_id, product_source_id, price_irr, usd_irr_rate_value_text, gold_gram_irr_rate_value_text, price_usd, price_gold_gram, captured_at, raw_payload, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pricePoint.ID,
		pricePoint.ProductID,
		pricePoint.ProductSourceID,
		pricePoint.PriceIRR,
		nullStringPtr(pricePoint.USDIRRRateValueText),
		nullStringPtr(pricePoint.GoldGramIRRRateValueText),
		nullStringPtr(pricePoint.PriceUSD),
		nullStringPtr(pricePoint.PriceGoldGram),
		formatTime(pricePoint.CapturedAt),
		nullStringPtr(pricePoint.RawPayload),
		formatTime(pricePoint.CreatedAt),
	)
	if err != nil {
		return PricePoint{}, fmt.Errorf("create price point: %w", err)
	}

	return pricePoint, nil
}

func (s *SQLiteStore) ListByProduct(ctx context.Context, productID string) ([]PricePoint, error) {
	if err := s.ensureProduct(ctx, productID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, product_id, product_source_id, price_irr, usd_irr_rate_value_text, gold_gram_irr_rate_value_text, price_usd, price_gold_gram, captured_at, raw_payload, created_at
FROM price_points
WHERE product_id = ?
ORDER BY captured_at ASC, id ASC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list price points: %w", err)
	}
	defer rows.Close()

	var pricePoints []PricePoint
	for rows.Next() {
		pricePoint, err := scanPricePoint(rows)
		if err != nil {
			return nil, err
		}
		pricePoints = append(pricePoints, pricePoint)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list price points rows: %w", err)
	}

	return pricePoints, nil
}

func (s *SQLiteStore) attachDerivedPrices(ctx context.Context, pricePoint *PricePoint) error {
	usdRate, err := s.latestRateAt(ctx, marketrate.RateTypeUSDIRR, pricePoint.CapturedAt)
	if err != nil {
		return err
	}
	if usdRate != nil {
		pricePoint.USDIRRRateValueText = usdRate
		priceUSD, err := DerivePriceText(pricePoint.PriceIRR, *usdRate)
		if err != nil {
			return err
		}
		pricePoint.PriceUSD = priceUSD
	}

	goldRate, err := s.latestRateAt(ctx, marketrate.RateTypeGoldGramIRR, pricePoint.CapturedAt)
	if err != nil {
		return err
	}
	if goldRate != nil {
		pricePoint.GoldGramIRRRateValueText = goldRate
		priceGoldGram, err := DerivePriceText(pricePoint.PriceIRR, *goldRate)
		if err != nil {
			return err
		}
		pricePoint.PriceGoldGram = priceGoldGram
	}

	return nil
}

func (s *SQLiteStore) latestRateAt(ctx context.Context, rateType string, capturedAt time.Time) (*string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, value_text, observed_at
FROM market_rates
WHERE rate_type = ?`, rateType)
	if err != nil {
		return nil, fmt.Errorf("load %s rate for price point: %w", rateType, err)
	}
	defer rows.Close()

	var bestValueText *string
	var bestObservedAt time.Time
	var bestID string
	for rows.Next() {
		var id string
		var valueText string
		var observedAtText string
		if err := rows.Scan(&id, &valueText, &observedAtText); err != nil {
			return nil, fmt.Errorf("scan %s rate for price point: %w", rateType, err)
		}
		observedAt, err := parseTime(observedAtText)
		if err != nil {
			return nil, err
		}
		if observedAt.After(capturedAt) {
			continue
		}
		if bestValueText == nil || observedAt.After(bestObservedAt) || (observedAt.Equal(bestObservedAt) && id > bestID) {
			value := valueText
			bestValueText = &value
			bestObservedAt = observedAt
			bestID = id
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load %s rate rows for price point: %w", rateType, err)
	}

	return bestValueText, nil
}

func (s *SQLiteStore) ensureProduct(ctx context.Context, productID string) error {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM products WHERE id = ?`, productID).Scan(&id)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("check product exists: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ensureProductSource(ctx context.Context, productID string, productSourceID string) error {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM product_sources WHERE product_id = ? AND id = ?`, productID, productSourceID).Scan(&id)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("check product source exists: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPricePoint(row scanner) (PricePoint, error) {
	var pricePoint PricePoint
	var capturedAt string
	var createdAt string
	var rawPayload sql.NullString
	var usdIRRRateValueText sql.NullString
	var goldGramIRRRateValueText sql.NullString
	var priceUSD sql.NullString
	var priceGoldGram sql.NullString

	err := row.Scan(
		&pricePoint.ID,
		&pricePoint.ProductID,
		&pricePoint.ProductSourceID,
		&pricePoint.PriceIRR,
		&usdIRRRateValueText,
		&goldGramIRRRateValueText,
		&priceUSD,
		&priceGoldGram,
		&capturedAt,
		&rawPayload,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return PricePoint{}, ErrNotFound
	}
	if err != nil {
		return PricePoint{}, fmt.Errorf("scan price point: %w", err)
	}
	if rawPayload.Valid {
		pricePoint.RawPayload = &rawPayload.String
	}
	if usdIRRRateValueText.Valid {
		pricePoint.USDIRRRateValueText = &usdIRRRateValueText.String
	}
	if goldGramIRRRateValueText.Valid {
		pricePoint.GoldGramIRRRateValueText = &goldGramIRRRateValueText.String
	}
	if priceUSD.Valid {
		pricePoint.PriceUSD = &priceUSD.String
	}
	if priceGoldGram.Valid {
		pricePoint.PriceGoldGram = &priceGoldGram.String
	}

	pricePoint.CapturedAt, err = parseTime(capturedAt)
	if err != nil {
		return PricePoint{}, err
	}
	pricePoint.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return PricePoint{}, err
	}

	return pricePoint, nil
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
		return time.Time{}, fmt.Errorf("parse price point time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate price point id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}
