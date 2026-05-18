package alert

import "testing"

func TestNormalizeCreateAcceptsSupportedAlert(t *testing.T) {
	input, err := NormalizeCreate(CreateInput{
		Name:               "  Price drop  ",
		ConditionType:      "below",
		TargetUnit:         "usd",
		ThresholdValueText: "10.50",
	})
	if err != nil {
		t.Fatalf("NormalizeCreate() error = %v", err)
	}

	if input.Name != "Price drop" {
		t.Fatalf("Name = %q, want Price drop", input.Name)
	}
	if input.ConditionType != ConditionBelow {
		t.Fatalf("ConditionType = %q, want %q", input.ConditionType, ConditionBelow)
	}
	if input.TargetUnit != UnitUSD {
		t.Fatalf("TargetUnit = %q, want %q", input.TargetUnit, UnitUSD)
	}
	if input.ThresholdValueText != "10.50" {
		t.Fatalf("ThresholdValueText = %q, want 10.50", input.ThresholdValueText)
	}
}

func TestNormalizeCreateRejectsInvalidThreshold(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{
		Name:               "Price drop",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitIRR,
		ThresholdValueText: "0",
	})
	if err == nil {
		t.Fatal("NormalizeCreate() error = nil, want validation error")
	}
}

func TestNormalizeUpdateRequiresField(t *testing.T) {
	_, err := NormalizeUpdate(UpdateInput{})
	if err == nil {
		t.Fatal("NormalizeUpdate() error = nil, want validation error")
	}
}
