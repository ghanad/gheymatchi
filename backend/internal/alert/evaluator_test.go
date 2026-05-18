package alert

import (
	"context"
	"testing"
	"time"

	"gheymatchi/backend/internal/price"
)

func TestShouldTriggerSupportsUnitsAndConditions(t *testing.T) {
	priceUSD := "1000"
	priceGold := "2.5"
	point := price.PricePoint{
		ProductID:     "product-1",
		PriceIRR:      90_000_000,
		PriceUSD:      &priceUSD,
		PriceGoldGram: &priceGold,
		CapturedAt:    time.Now().UTC(),
	}

	tests := []struct {
		name  string
		alert Alert
		want  bool
	}{
		{
			name:  "IRR below",
			alert: Alert{ID: "a1", ProductID: "product-1", ConditionType: ConditionBelow, TargetUnit: UnitIRR, ThresholdValueText: "95000000", IsActive: true},
			want:  true,
		},
		{
			name:  "USD above",
			alert: Alert{ID: "a2", ProductID: "product-1", ConditionType: ConditionAbove, TargetUnit: UnitUSD, ThresholdValueText: "900", IsActive: true},
			want:  true,
		},
		{
			name:  "gold below",
			alert: Alert{ID: "a3", ProductID: "product-1", ConditionType: ConditionBelow, TargetUnit: UnitGoldGram, ThresholdValueText: "3", IsActive: true},
			want:  true,
		},
		{
			name:  "inactive",
			alert: Alert{ID: "a4", ProductID: "product-1", ConditionType: ConditionBelow, TargetUnit: UnitIRR, ThresholdValueText: "95000000", IsActive: false},
			want:  false,
		},
		{
			name:  "already triggered",
			alert: Alert{ID: "a5", ProductID: "product-1", ConditionType: ConditionBelow, TargetUnit: UnitIRR, ThresholdValueText: "95000000", IsActive: true, LastTriggeredAt: timePtr(time.Now().UTC())},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ShouldTrigger(tt.alert, point)
			if err != nil {
				t.Fatalf("ShouldTrigger() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ShouldTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldTriggerSkipsMissingDerivedValue(t *testing.T) {
	got, err := ShouldTrigger(Alert{
		ID:                 "alert-1",
		ProductID:          "product-1",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitUSD,
		ThresholdValueText: "1000",
		IsActive:           true,
	}, price.PricePoint{ProductID: "product-1", PriceIRR: 90_000_000})
	if err != nil {
		t.Fatalf("ShouldTrigger() error = %v", err)
	}
	if got {
		t.Fatal("ShouldTrigger() = true, want false")
	}
}

func TestEvaluatorCreatesNotificationAndMarksTriggered(t *testing.T) {
	capturedAt := time.Now().UTC()
	alertStore := &fakeEvaluationStore{alerts: []Alert{{
		ID:                 "alert-1",
		ProductID:          "product-1",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitIRR,
		ThresholdValueText: "100",
		IsActive:           true,
	}}}
	notifications := &fakeNotificationCreator{}
	evaluator := NewEvaluator(alertStore, notifications)

	err := evaluator.EvaluatePricePoint(context.Background(), price.PricePoint{
		ID:         "price-1",
		ProductID:  "product-1",
		PriceIRR:   90,
		CapturedAt: capturedAt,
	})
	if err != nil {
		t.Fatalf("EvaluatePricePoint() error = %v", err)
	}
	if len(notifications.created) != 1 {
		t.Fatalf("created notifications = %d, want 1", len(notifications.created))
	}
	if alertStore.triggeredAlertID != "alert-1" {
		t.Fatalf("triggered alert ID = %q, want alert-1", alertStore.triggeredAlertID)
	}
	if !alertStore.triggeredAt.Equal(capturedAt) {
		t.Fatalf("triggeredAt = %s, want %s", alertStore.triggeredAt, capturedAt)
	}
}

type fakeEvaluationStore struct {
	alerts           []Alert
	triggeredAlertID string
	triggeredAt      time.Time
}

func (f *fakeEvaluationStore) ListActiveByProduct(ctx context.Context, productID string) ([]Alert, error) {
	return f.alerts, nil
}

func (f *fakeEvaluationStore) MarkTriggered(ctx context.Context, alertID string, triggeredAt time.Time) error {
	f.triggeredAlertID = alertID
	f.triggeredAt = triggeredAt
	return nil
}

type fakeNotificationCreator struct {
	created []Alert
}

func (f *fakeNotificationCreator) CreateAlertTriggered(ctx context.Context, alert Alert, pricePoint price.PricePoint) error {
	f.created = append(f.created, alert)
	return nil
}

func timePtr(value time.Time) *time.Time {
	return &value
}
