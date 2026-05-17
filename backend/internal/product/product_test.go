package product

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeCreateRequiresName(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{Name: "   "})
	if err == nil {
		t.Fatal("NormalizeCreate() error = nil, want validation error")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %T, want ValidationError", err)
	}
	if validationErr.Field != "name" {
		t.Fatalf("field = %q, want name", validationErr.Field)
	}
}

func TestNormalizeCreateTrimsInput(t *testing.T) {
	input, err := NormalizeCreate(CreateInput{Name: "  Phone  ", Description: "  local watch  "})
	if err != nil {
		t.Fatalf("NormalizeCreate() error = %v", err)
	}

	if input.Name != "Phone" {
		t.Fatalf("name = %q, want Phone", input.Name)
	}
	if input.Description != "local watch" {
		t.Fatalf("description = %q, want local watch", input.Description)
	}
}

func TestNormalizeUpdateRejectsEmptyBody(t *testing.T) {
	_, err := NormalizeUpdate(UpdateInput{})
	if err == nil {
		t.Fatal("NormalizeUpdate() error = nil, want validation error")
	}
}

func TestNormalizeUpdateRejectsLongName(t *testing.T) {
	name := strings.Repeat("a", maxNameLength+1)

	_, err := NormalizeUpdate(UpdateInput{Name: &name})
	if err == nil {
		t.Fatal("NormalizeUpdate() error = nil, want validation error")
	}
}
