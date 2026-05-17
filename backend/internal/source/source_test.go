package source

import "testing"

func TestNormalizeCreateRejectsInvalidURL(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{URL: "not-a-url", SourceName: "digikala"})
	if err == nil {
		t.Fatal("NormalizeCreate() error = nil, want validation error")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	if validationErr.Field != "url" {
		t.Fatalf("field = %q, want url", validationErr.Field)
	}
}

func TestNormalizeCreateDefaultsSourceName(t *testing.T) {
	input, err := NormalizeCreate(CreateInput{URL: " https://example.com/product/1 "})
	if err != nil {
		t.Fatalf("NormalizeCreate() error = %v", err)
	}
	if input.URL != "https://example.com/product/1" {
		t.Fatalf("URL = %q, want normalized URL", input.URL)
	}
	if input.SourceName != "unknown" {
		t.Fatalf("SourceName = %q, want unknown", input.SourceName)
	}
}
