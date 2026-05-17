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
