package price

import (
	"errors"
	"testing"
)

func TestNormalizeCreateRequiresPositivePrice(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{PriceIRR: 0})
	var validationErr ValidationError
	if err == nil {
		t.Fatal("NormalizeCreate() error = nil, want validation error")
	}
	if !errors.As(err, &validationErr) {
		t.Fatalf("NormalizeCreate() error = %T, want ValidationError", err)
	}
	if validationErr.Field != "price_irr" {
		t.Fatalf("ValidationError.Field = %q, want price_irr", validationErr.Field)
	}
}

func TestDerivePriceText(t *testing.T) {
	derived, err := DerivePriceText(123000000, "615000")
	if err != nil {
		t.Fatalf("DerivePriceText() error = %v", err)
	}
	if derived == nil || *derived != "200" {
		t.Fatalf("DerivePriceText() = %v, want 200", derived)
	}
}

func TestDerivePriceTextMissingRate(t *testing.T) {
	derived, err := DerivePriceText(123000000, "")
	if err != nil {
		t.Fatalf("DerivePriceText() error = %v", err)
	}
	if derived != nil {
		t.Fatalf("DerivePriceText() = %v, want nil", *derived)
	}
}

func TestDerivePriceTextRejectsInvalidRate(t *testing.T) {
	_, err := DerivePriceText(123000000, "0")
	var validationErr ValidationError
	if err == nil {
		t.Fatal("DerivePriceText() error = nil, want validation error")
	}
	if !errors.As(err, &validationErr) {
		t.Fatalf("DerivePriceText() error = %T, want ValidationError", err)
	}
	if validationErr.Field != "rate_value_text" {
		t.Fatalf("ValidationError.Field = %q, want rate_value_text", validationErr.Field)
	}
}
