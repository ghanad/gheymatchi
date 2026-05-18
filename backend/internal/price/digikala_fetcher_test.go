package price

import (
	"errors"
	"os"
	"testing"
)

func TestDigikalaProductID(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "product url",
			url:  "https://www.digikala.com/product/dkp-1234567/example-product/",
			want: "1234567",
		},
		{
			name: "mobile product url",
			url:  "https://digikala.com/product/dkp-42/",
			want: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := digikalaProductID(tt.url)
			if err != nil {
				t.Fatalf("digikalaProductID() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("digikalaProductID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDigikalaProductIDUnsupportedURL(t *testing.T) {
	_, err := digikalaProductID("https://example.com/product/dkp-123/")
	if !errors.Is(err, ErrUnsupportedSourceURL) {
		t.Fatalf("digikalaProductID() error = %v, want ErrUnsupportedSourceURL", err)
	}
}

func TestParseDigikalaPriceResponse(t *testing.T) {
	body := readTestFixture(t, "digikala_marketable.json")

	parsed, err := parseDigikalaPriceResponse(body)
	if err != nil {
		t.Fatalf("parseDigikalaPriceResponse() error = %v", err)
	}
	if parsed.PriceIRR != 129_990_000 {
		t.Fatalf("PriceIRR = %d, want 129990000", parsed.PriceIRR)
	}
	if parsed.Status != "marketable" {
		t.Fatalf("Status = %q, want marketable", parsed.Status)
	}
}

func TestParseDigikalaUnavailableResponse(t *testing.T) {
	body := readTestFixture(t, "digikala_unavailable.json")

	_, err := parseDigikalaPriceResponse(body)
	if !errors.Is(err, ErrProductUnavailable) {
		t.Fatalf("parseDigikalaPriceResponse() error = %v, want ErrProductUnavailable", err)
	}
}

func TestParseDigikalaMissingPriceResponse(t *testing.T) {
	body := readTestFixture(t, "digikala_missing_price.json")

	_, err := parseDigikalaPriceResponse(body)
	if !errors.Is(err, ErrPriceNotFound) {
		t.Fatalf("parseDigikalaPriceResponse() error = %v, want ErrPriceNotFound", err)
	}
}

func readTestFixture(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}
