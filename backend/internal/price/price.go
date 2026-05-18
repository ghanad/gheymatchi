package price

import (
	"errors"
	"math/big"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("price resource not found")
)

type PricePoint struct {
	ID                       string    `json:"id"`
	ProductID                string    `json:"product_id"`
	ProductSourceID          string    `json:"product_source_id"`
	PriceIRR                 int64     `json:"price_irr"`
	USDIRRRateValueText      *string   `json:"usd_irr_rate_value_text,omitempty"`
	GoldGramIRRRateValueText *string   `json:"gold_gram_irr_rate_value_text,omitempty"`
	PriceUSD                 *string   `json:"price_usd,omitempty"`
	PriceGoldGram            *string   `json:"price_gold_gram,omitempty"`
	CapturedAt               time.Time `json:"captured_at"`
	RawPayload               *string   `json:"raw_payload,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
}

type CreateInput struct {
	PriceIRR   int64
	CapturedAt time.Time
	RawPayload *string
}

func NormalizeCreate(input CreateInput) (CreateInput, error) {
	if input.PriceIRR <= 0 {
		return CreateInput{}, fieldError("price_irr", "must be a positive integer")
	}
	if input.CapturedAt.IsZero() {
		input.CapturedAt = time.Now().UTC()
	} else {
		input.CapturedAt = input.CapturedAt.UTC()
	}
	return input, nil
}

func DerivePriceText(priceIRR int64, rateValueText string) (*string, error) {
	if priceIRR <= 0 {
		return nil, fieldError("price_irr", "must be a positive integer")
	}
	rate := strings.TrimSpace(rateValueText)
	if rate == "" {
		return nil, nil
	}

	rateValue, ok := new(big.Rat).SetString(rate)
	if !ok || rateValue.Sign() <= 0 {
		return nil, fieldError("rate_value_text", "must be a positive decimal value")
	}

	derived := new(big.Rat).SetInt64(priceIRR)
	derived.Quo(derived, rateValue)
	result := trimDecimalZeros(derived.FloatString(8))
	return &result, nil
}

func trimDecimalZeros(value string) string {
	value = strings.TrimRight(value, "0")
	value = strings.TrimRight(value, ".")
	if value == "" {
		return "0"
	}
	return value
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + " " + e.Message
}

func fieldError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}
