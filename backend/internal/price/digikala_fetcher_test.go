package price

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
		{
			name: "numeric product url",
			url:  "https://www.digikala.com/product/20769143/",
			want: "20769143",
		},
		{
			name: "fresh product url",
			url:  "https://www.digikala.com/fresh/product/dkp-856977/",
			want: "856977",
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

func TestParseDigikalaSourceURL(t *testing.T) {
	info, err := parseDigikalaSourceURL("https://www.digikala.com/fresh/product/dkp-14590108/?variant_id=50179968")
	if err != nil {
		t.Fatalf("parseDigikalaSourceURL() error = %v", err)
	}
	if info.ProductID != "14590108" {
		t.Fatalf("ProductID = %q, want 14590108", info.ProductID)
	}
	if info.VariantID != "50179968" {
		t.Fatalf("VariantID = %q, want 50179968", info.VariantID)
	}
	if !info.IsFresh {
		t.Fatal("IsFresh = false, want true")
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

func TestParseDigikalaPriceResponseSelectsRequestedVariant(t *testing.T) {
	body := []byte(`{
		"data": {
			"product": {
				"title_fa": "Test product",
				"status": "marketable",
				"default_variant": {
					"id": 1,
					"status": "marketable",
					"price": {"selling_price": 100},
					"seller": {"title": "Default seller"}
				},
				"variants": [
					{
						"id": 2,
						"status": "marketable",
						"price": {"final_price": 200},
						"seller": {"title_fa": "Requested seller"}
					}
				]
			}
		}
	}`)

	parsed, err := parseDigikalaPriceResponseForVariant(body, "2")
	if err != nil {
		t.Fatalf("parseDigikalaPriceResponseForVariant() error = %v", err)
	}
	if parsed.PriceIRR != 200 {
		t.Fatalf("PriceIRR = %d, want 200", parsed.PriceIRR)
	}
	if parsed.VariantID != "2" {
		t.Fatalf("VariantID = %q, want 2", parsed.VariantID)
	}
	if parsed.Seller != "Requested seller" {
		t.Fatalf("Seller = %q, want Requested seller", parsed.Seller)
	}
	if parsed.Title != "Test product" {
		t.Fatalf("Title = %q, want Test product", parsed.Title)
	}
}

func TestParseDigikalaPriceResponseUsesFirstMarketableVariant(t *testing.T) {
	body := []byte(`{
		"data": {
			"product": {
				"variants": [
					{"id": 1, "status": "not_marketable", "price": {"selling_price": 100}},
					{"id": 2, "status": "marketable", "price": 200}
				]
			}
		}
	}`)

	parsed, err := parseDigikalaPriceResponse(body)
	if err != nil {
		t.Fatalf("parseDigikalaPriceResponse() error = %v", err)
	}
	if parsed.PriceIRR != 200 {
		t.Fatalf("PriceIRR = %d, want 200", parsed.PriceIRR)
	}
}

func TestHasDigikalaFreshRedirect(t *testing.T) {
	body := []byte(`{"status":302,"redirect_url":{"uri":"/fresh/product/dkp-856977/"}}`)
	if !hasDigikalaFreshRedirect(body) {
		t.Fatal("hasDigikalaFreshRedirect() = false, want true")
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

func TestDigikalaFetcherFetchJSONAccessDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	fetcher := NewDigikalaFetcher(server.Client(), 0, 0)
	_, err := fetcher.fetchJSON(context.Background(), server.URL)
	if !errors.Is(err, ErrSourceAccessDenied) {
		t.Fatalf("fetchJSON() error = %v, want ErrSourceAccessDenied", err)
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
