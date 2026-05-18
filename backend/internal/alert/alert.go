package alert

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	ConditionBelow = "BELOW"
	ConditionAbove = "ABOVE"

	UnitIRR      = "IRR"
	UnitUSD      = "USD"
	UnitGoldGram = "GOLD_GRAM"

	maxNameLength = 160
)

var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

var (
	ErrNotFound = errors.New("alert not found")
)

type Alert struct {
	ID                 string    `json:"id"`
	UserID             *string   `json:"user_id,omitempty"`
	ProductID          string    `json:"product_id"`
	Name               string    `json:"name"`
	ConditionType      string    `json:"condition_type"`
	TargetUnit         string    `json:"target_unit"`
	ThresholdValueText string    `json:"threshold_value_text"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name               string
	ConditionType      string
	TargetUnit         string
	ThresholdValueText string
	IsActive           *bool
}

type UpdateInput struct {
	Name               *string
	ConditionType      *string
	TargetUnit         *string
	ThresholdValueText *string
	IsActive           *bool
}

func NormalizeCreate(input CreateInput) (CreateInput, error) {
	name, err := normalizeName(input.Name)
	if err != nil {
		return CreateInput{}, err
	}
	conditionType, err := normalizeConditionType(input.ConditionType)
	if err != nil {
		return CreateInput{}, err
	}
	targetUnit, err := normalizeTargetUnit(input.TargetUnit)
	if err != nil {
		return CreateInput{}, err
	}
	threshold, err := normalizeThreshold(input.ThresholdValueText)
	if err != nil {
		return CreateInput{}, err
	}

	return CreateInput{
		Name:               name,
		ConditionType:      conditionType,
		TargetUnit:         targetUnit,
		ThresholdValueText: threshold,
		IsActive:           input.IsActive,
	}, nil
}

func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	if input.Name != nil {
		name, err := normalizeName(*input.Name)
		if err != nil {
			return UpdateInput{}, err
		}
		input.Name = &name
	}
	if input.ConditionType != nil {
		conditionType, err := normalizeConditionType(*input.ConditionType)
		if err != nil {
			return UpdateInput{}, err
		}
		input.ConditionType = &conditionType
	}
	if input.TargetUnit != nil {
		targetUnit, err := normalizeTargetUnit(*input.TargetUnit)
		if err != nil {
			return UpdateInput{}, err
		}
		input.TargetUnit = &targetUnit
	}
	if input.ThresholdValueText != nil {
		threshold, err := normalizeThreshold(*input.ThresholdValueText)
		if err != nil {
			return UpdateInput{}, err
		}
		input.ThresholdValueText = &threshold
	}

	if input.Name == nil && input.ConditionType == nil && input.TargetUnit == nil && input.ThresholdValueText == nil && input.IsActive == nil {
		return UpdateInput{}, fieldError("body", "must include at least one supported field")
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

func normalizeName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", fieldError("name", "is required")
	}
	if len(name) > maxNameLength {
		return "", fieldError("name", "must be 160 characters or less")
	}
	return name, nil
}

func normalizeConditionType(value string) (string, error) {
	conditionType := strings.ToUpper(strings.TrimSpace(value))
	if conditionType != ConditionBelow && conditionType != ConditionAbove {
		return "", fieldError("condition_type", "must be BELOW or ABOVE")
	}
	return conditionType, nil
}

func normalizeTargetUnit(value string) (string, error) {
	targetUnit := strings.ToUpper(strings.TrimSpace(value))
	if targetUnit != UnitIRR && targetUnit != UnitUSD && targetUnit != UnitGoldGram {
		return "", fieldError("target_unit", "must be IRR, USD, or GOLD_GRAM")
	}
	return targetUnit, nil
}

func normalizeThreshold(value string) (string, error) {
	threshold := strings.TrimSpace(value)
	if threshold == "" {
		return "", fieldError("threshold_value_text", "is required")
	}
	if !decimalPattern.MatchString(threshold) {
		return "", fieldError("threshold_value_text", "must be a positive decimal value")
	}
	if isZeroDecimal(threshold) {
		return "", fieldError("threshold_value_text", "must be greater than zero")
	}
	return threshold, nil
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
