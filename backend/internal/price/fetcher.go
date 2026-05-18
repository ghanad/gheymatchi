package price

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"time"

	"gheymatchi/backend/internal/source"
)

type FetchResult struct {
	PriceIRR   int64
	CapturedAt time.Time
	RawPayload *string
}

type Fetcher interface {
	Fetch(ctx context.Context, productSource source.ProductSource) (FetchResult, error)
}

type MockPriceFetcher struct{}

func NewMockPriceFetcher() MockPriceFetcher {
	return MockPriceFetcher{}
}

func (f MockPriceFetcher) Fetch(ctx context.Context, productSource source.ProductSource) (FetchResult, error) {
	if err := ctx.Err(); err != nil {
		return FetchResult{}, err
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(productSource.ID + "|" + productSource.URL))
	priceIRR := int64(10_000_000 + hash.Sum32()%90_000_000)
	capturedAt := time.Now().UTC()

	payload, err := json.Marshal(map[string]any{
		"provider":  "mock",
		"source_id": productSource.ID,
		"url":       productSource.URL,
		"price_irr": priceIRR,
	})
	if err != nil {
		return FetchResult{}, err
	}
	rawPayload := string(payload)

	return FetchResult{
		PriceIRR:   priceIRR,
		CapturedAt: capturedAt,
		RawPayload: &rawPayload,
	}, nil
}
