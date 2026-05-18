package alert

import (
	"context"
	"fmt"
	"math/big"

	"gheymatchi/backend/internal/price"
)

type NotificationCreator interface {
	CreateAlertTriggered(ctx context.Context, alert Alert, pricePoint price.PricePoint) error
}

type Evaluator struct {
	alerts        EvaluationStore
	notifications NotificationCreator
}

func NewEvaluator(alerts EvaluationStore, notifications NotificationCreator) Evaluator {
	return Evaluator{alerts: alerts, notifications: notifications}
}

func (e Evaluator) EvaluatePricePoint(ctx context.Context, pricePoint price.PricePoint) error {
	alerts, err := e.alerts.ListActiveByProduct(ctx, pricePoint.ProductID)
	if err != nil {
		return err
	}

	for _, alert := range alerts {
		triggered, err := ShouldTrigger(alert, pricePoint)
		if err != nil {
			return err
		}
		if !triggered {
			continue
		}
		if err := e.notifications.CreateAlertTriggered(ctx, alert, pricePoint); err != nil {
			return err
		}
		if err := e.alerts.MarkTriggered(ctx, alert.ID, pricePoint.CapturedAt); err != nil {
			return err
		}
	}

	return nil
}

func ShouldTrigger(alert Alert, pricePoint price.PricePoint) (bool, error) {
	if !alert.IsActive || alert.LastTriggeredAt != nil {
		return false, nil
	}

	current, ok := priceValueForUnit(alert.TargetUnit, pricePoint)
	if !ok {
		return false, nil
	}

	threshold, ok := new(big.Rat).SetString(alert.ThresholdValueText)
	if !ok {
		return false, fmt.Errorf("parse alert threshold %s: %w", alert.ID, ValidationError{Field: "threshold_value_text", Message: "must be a decimal value"})
	}

	switch alert.ConditionType {
	case ConditionBelow:
		return current.Cmp(threshold) < 0, nil
	case ConditionAbove:
		return current.Cmp(threshold) > 0, nil
	default:
		return false, fmt.Errorf("alert %s has unsupported condition type %q", alert.ID, alert.ConditionType)
	}
}

func priceValueForUnit(unit string, pricePoint price.PricePoint) (*big.Rat, bool) {
	switch unit {
	case UnitIRR:
		return new(big.Rat).SetInt64(pricePoint.PriceIRR), true
	case UnitUSD:
		return parseOptionalDecimal(pricePoint.PriceUSD)
	case UnitGoldGram:
		return parseOptionalDecimal(pricePoint.PriceGoldGram)
	default:
		return nil, false
	}
}

func parseOptionalDecimal(value *string) (*big.Rat, bool) {
	if value == nil {
		return nil, false
	}
	parsed, ok := new(big.Rat).SetString(*value)
	return parsed, ok
}
