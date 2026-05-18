package marketrate

import (
	"context"
	"testing"
)

func TestNormalizeCreate(t *testing.T) {
	input, err := NormalizeCreate(CreateInput{
		RateType:  " usd_irr ",
		ValueText: " 920000 ",
	})
	if err != nil {
		t.Fatalf("NormalizeCreate() error = %v", err)
	}
	if input.RateType != RateTypeUSDIRR {
		t.Fatalf("RateType = %q, want %q", input.RateType, RateTypeUSDIRR)
	}
	if input.ValueText != "920000" {
		t.Fatalf("ValueText = %q, want 920000", input.ValueText)
	}
	if input.ObservedAt.IsZero() {
		t.Fatal("ObservedAt is zero")
	}
}

func TestNormalizeCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input CreateInput
		field string
	}{
		{
			name:  "unknown rate type",
			input: CreateInput{RateType: "EUR_IRR", ValueText: "1"},
			field: "rate_type",
		},
		{
			name:  "missing value",
			input: CreateInput{RateType: RateTypeUSDIRR},
			field: "value_text",
		},
		{
			name:  "zero value",
			input: CreateInput{RateType: RateTypeUSDIRR, ValueText: "0.00"},
			field: "value_text",
		},
		{
			name:  "negative value",
			input: CreateInput{RateType: RateTypeUSDIRR, ValueText: "-1"},
			field: "value_text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeCreate(tt.input)
			validationErr, ok := err.(ValidationError)
			if !ok {
				t.Fatalf("error = %T %v, want ValidationError", err, err)
			}
			if validationErr.Field != tt.field {
				t.Fatalf("field = %q, want %q", validationErr.Field, tt.field)
			}
		})
	}
}

func TestMockProviderFetchReturnsRates(t *testing.T) {
	provider := MockProvider{Rates: []CreateInput{
		{RateType: RateTypeUSDIRR, ValueText: "920000"},
	}}

	rates, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(rates) != 1 {
		t.Fatalf("len(rates) = %d, want 1", len(rates))
	}
	if rates[0].RateType != RateTypeUSDIRR {
		t.Fatalf("RateType = %q, want %q", rates[0].RateType, RateTypeUSDIRR)
	}
}
