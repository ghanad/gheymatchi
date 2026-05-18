package marketrate

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(ctx context.Context, input CreateInput) (MarketRate, error) {
	normalized, err := NormalizeCreate(input)
	if err != nil {
		return MarketRate{}, err
	}

	now := time.Now().UTC()
	rate := MarketRate{
		ID:         newID(),
		RateType:   normalized.RateType,
		Unit:       UnitForRateType(normalized.RateType),
		ValueText:  normalized.ValueText,
		ObservedAt: normalized.ObservedAt,
		CreatedAt:  now,
	}

	_, err = s.db.ExecContext(ctx, `INSERT INTO market_rates
(id, rate_type, unit, value_text, observed_at, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		rate.ID,
		rate.RateType,
		rate.Unit,
		rate.ValueText,
		formatTime(rate.ObservedAt),
		formatTime(rate.CreatedAt),
	)
	if err != nil {
		return MarketRate{}, fmt.Errorf("create market rate: %w", err)
	}

	return rate, nil
}

func (s *SQLiteStore) Latest(ctx context.Context, rateType *string) ([]MarketRate, error) {
	if rateType != nil {
		normalized, err := NormalizeRateType(*rateType)
		if err != nil {
			return nil, err
		}

		rate, err := s.latestByType(ctx, normalized)
		if err != nil {
			if err == ErrNotFound {
				return []MarketRate{}, nil
			}
			return nil, err
		}
		return []MarketRate{rate}, nil
	}

	rates := make([]MarketRate, 0, 2)
	for _, knownType := range []string{RateTypeUSDIRR, RateTypeGoldGramIRR} {
		rate, err := s.latestByType(ctx, knownType)
		if err == nil {
			rates = append(rates, rate)
			continue
		}
		if err != ErrNotFound {
			return nil, err
		}
	}
	return rates, nil
}

func (s *SQLiteStore) History(ctx context.Context, rateType *string) ([]MarketRate, error) {
	var rows *sql.Rows
	var err error

	if rateType != nil {
		normalized, normalizeErr := NormalizeRateType(*rateType)
		if normalizeErr != nil {
			return nil, normalizeErr
		}
		rows, err = s.db.QueryContext(ctx, `
SELECT id, rate_type, unit, value_text, observed_at, created_at
FROM market_rates
WHERE rate_type = ?
ORDER BY observed_at DESC, id DESC`, normalized)
	} else {
		rows, err = s.db.QueryContext(ctx, `
SELECT id, rate_type, unit, value_text, observed_at, created_at
FROM market_rates
ORDER BY observed_at DESC, id DESC`)
	}
	if err != nil {
		return nil, fmt.Errorf("list market rates: %w", err)
	}
	defer rows.Close()

	var rates []MarketRate
	for rows.Next() {
		rate, err := scanMarketRate(rows)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list market rate rows: %w", err)
	}

	return rates, nil
}

func (s *SQLiteStore) latestByType(ctx context.Context, rateType string) (MarketRate, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, rate_type, unit, value_text, observed_at, created_at
FROM market_rates
WHERE rate_type = ?
ORDER BY observed_at DESC, id DESC
LIMIT 1`, rateType)

	return scanMarketRate(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanMarketRate(row scanner) (MarketRate, error) {
	var rate MarketRate
	var observedAt string
	var createdAt string

	err := row.Scan(
		&rate.ID,
		&rate.RateType,
		&rate.Unit,
		&rate.ValueText,
		&observedAt,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return MarketRate{}, ErrNotFound
	}
	if err != nil {
		return MarketRate{}, fmt.Errorf("scan market rate: %w", err)
	}

	rate.ObservedAt, err = parseTime(observedAt)
	if err != nil {
		return MarketRate{}, err
	}
	rate.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return MarketRate{}, err
	}

	return rate, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse market rate time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate market rate id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}
