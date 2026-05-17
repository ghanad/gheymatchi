package price

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("price resource not found")
)

type PricePoint struct {
	ID              string    `json:"id"`
	ProductID       string    `json:"product_id"`
	ProductSourceID string    `json:"product_source_id"`
	PriceIRR        int64     `json:"price_irr"`
	CapturedAt      time.Time `json:"captured_at"`
	RawPayload      *string   `json:"raw_payload,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
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
