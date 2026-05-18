package price

import (
	"context"
	"testing"

	"gheymatchi/backend/internal/source"
)

func TestMockPriceFetcherReturnsDeterministicPrice(t *testing.T) {
	fetcher := NewMockPriceFetcher()
	productSource := source.ProductSource{
		ID:        "source-1",
		ProductID: "product-1",
		URL:       "https://example.com/products/1",
		IsActive:  true,
	}

	first, err := fetcher.Fetch(context.Background(), productSource)
	if err != nil {
		t.Fatalf("Fetch() first error = %v", err)
	}
	second, err := fetcher.Fetch(context.Background(), productSource)
	if err != nil {
		t.Fatalf("Fetch() second error = %v", err)
	}

	if first.PriceIRR <= 0 {
		t.Fatalf("PriceIRR = %d, want positive", first.PriceIRR)
	}
	if second.PriceIRR != first.PriceIRR {
		t.Fatalf("second PriceIRR = %d, want %d", second.PriceIRR, first.PriceIRR)
	}
	if first.RawPayload == nil || *first.RawPayload == "" {
		t.Fatal("RawPayload is empty")
	}
}
