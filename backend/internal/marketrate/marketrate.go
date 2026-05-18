package marketrate

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	RateTypeUSDIRR      = "USD_IRR"
	RateTypeGoldGramIRR = "GOLD_GRAM_IRR"

	UnitIRR = "IRR"
)

var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

var (
	ErrNotFound = errors.New("market rate not found")
)

type MarketRate struct {
	ID         string    `json:"id"`
	RateType   string    `json:"rate_type"`
	Unit       string    `json:"unit"`
	ValueText  string    `json:"value_text"`
	ObservedAt time.Time `json:"observed_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateInput struct {
	RateType   string
	ValueText  string
	ObservedAt time.Time
}

func NormalizeCreate(input CreateInput) (CreateInput, error) {
	rateType, err := NormalizeRateType(input.RateType)
	if err != nil {
		return CreateInput{}, err
	}
	valueText, err := normalizeValueText(input.ValueText)
	if err != nil {
		return CreateInput{}, err
	}
	if input.ObservedAt.IsZero() {
		input.ObservedAt = time.Now().UTC()
	} else {
		input.ObservedAt = input.ObservedAt.UTC()
	}

	return CreateInput{
		RateType:   rateType,
		ValueText:  valueText,
		ObservedAt: input.ObservedAt,
	}, nil
}

func NormalizeRateType(value string) (string, error) {
	rateType := strings.ToUpper(strings.TrimSpace(value))
	if rateType != RateTypeUSDIRR && rateType != RateTypeGoldGramIRR {
		return "", fieldError("rate_type", "must be USD_IRR or GOLD_GRAM_IRR")
	}
	return rateType, nil
}

func UnitForRateType(rateType string) string {
	switch rateType {
	case RateTypeUSDIRR, RateTypeGoldGramIRR:
		return UnitIRR
	default:
		return ""
	}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + " " + e.Message
}

func normalizeValueText(value string) (string, error) {
	valueText := strings.TrimSpace(value)
	if valueText == "" {
		return "", fieldError("value_text", "is required")
	}
	if !decimalPattern.MatchString(valueText) {
		return "", fieldError("value_text", "must be a positive decimal value")
	}
	if isZeroDecimal(valueText) {
		return "", fieldError("value_text", "must be greater than zero")
	}
	return valueText, nil
}

func isZeroDecimal(value string) bool {
	for _, r := range value {
		if r >= '1' && r <= '9' {
			return false
		}
	}
	return true
}

func fieldError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}
