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

func TestNormalizeCreateDefaultsDigikalaSourceName(t *testing.T) {
	input, err := NormalizeCreate(CreateInput{URL: " https://www.digikala.com/product/dkp-123456/ "})
	if err != nil {
		t.Fatalf("NormalizeCreate() error = %v", err)
	}
	if input.URL != "https://www.digikala.com/product/dkp-123456/" {
		t.Fatalf("URL = %q, want normalized URL", input.URL)
	}
	if input.SourceName != "digikala" {
		t.Fatalf("SourceName = %q, want digikala", input.SourceName)
	}
}

func TestNormalizeCreateAcceptsNumericDigikalaProductURL(t *testing.T) {
	input, err := NormalizeCreate(CreateInput{URL: "https://www.digikala.com/product/20769143/"})
	if err != nil {
		t.Fatalf("NormalizeCreate() error = %v", err)
	}
	if input.SourceName != "digikala" {
		t.Fatalf("SourceName = %q, want digikala", input.SourceName)
	}
}

func TestNormalizeCreateRejectsUnsupportedSourceName(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{URL: "https://www.digikala.com/product/dkp-123456/", SourceName: "example"})
	if err == nil {
		t.Fatal("NormalizeCreate() error = nil, want validation error")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	if validationErr.Field != "source_name" {
		t.Fatalf("field = %q, want source_name", validationErr.Field)
	}
}

func TestNormalizeCreateRejectsNonDigikalaURL(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{URL: "https://example.com/product/dkp-123456/", SourceName: "digikala"})
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

func TestNormalizeCreateRejectsDigikalaURLWithoutProductID(t *testing.T) {
	_, err := NormalizeCreate(CreateInput{URL: "https://www.digikala.com/search/category-mobile-phone/", SourceName: "digikala"})
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
