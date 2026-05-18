package price

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gheymatchi/backend/internal/source"
)

const digikalaAPIBaseURL = "https://api.digikala.com/v2/product/"

type DigikalaFetcher struct {
	client       *http.Client
	timeout      time.Duration
	requestDelay time.Duration
	userAgent    string

	mu          sync.Mutex
	lastRequest time.Time
}

func NewDigikalaFetcher(client *http.Client, timeout time.Duration, requestDelay time.Duration) *DigikalaFetcher {
	if client == nil {
		client = http.DefaultClient
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if requestDelay < 0 {
		requestDelay = 0
	}

	return &DigikalaFetcher{
		client:       client,
		timeout:      timeout,
		requestDelay: requestDelay,
		userAgent:    "GheymatChi/phase18 (+local price tracking; contact: local-dev)",
	}
}

func (f *DigikalaFetcher) Fetch(ctx context.Context, productSource source.ProductSource) (FetchResult, error) {
	dkp, err := digikalaProductID(productSource.URL)
	if err != nil {
		return FetchResult{}, err
	}

	if err := f.waitForRateLimit(ctx); err != nil {
		return FetchResult{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	apiURL := digikalaAPIBaseURL + url.PathEscape(dkp) + "/"
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return FetchResult{}, fmt.Errorf("build digikala request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetch digikala product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return FetchResult{}, ErrProductUnavailable
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return FetchResult{}, fmt.Errorf("fetch digikala product: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return FetchResult{}, fmt.Errorf("read digikala response: %w", err)
	}

	parsed, err := parseDigikalaPriceResponse(body)
	if err != nil {
		return FetchResult{}, err
	}

	capturedAt := time.Now().UTC()
	rawPayload, err := json.Marshal(map[string]any{
		"provider":    "digikala",
		"source_id":   productSource.ID,
		"url":         productSource.URL,
		"api_url":     apiURL,
		"dkp":         dkp,
		"price_irr":   parsed.PriceIRR,
		"status":      parsed.Status,
		"captured_at": capturedAt.Format(time.RFC3339),
	})
	if err != nil {
		return FetchResult{}, fmt.Errorf("build digikala metadata: %w", err)
	}
	rawPayloadText := string(rawPayload)

	return FetchResult{
		PriceIRR:   parsed.PriceIRR,
		CapturedAt: capturedAt,
		RawPayload: &rawPayloadText,
	}, nil
}

func (f *DigikalaFetcher) waitForRateLimit(ctx context.Context) error {
	if f.requestDelay == 0 {
		return ctx.Err()
	}

	f.mu.Lock()
	wait := time.Until(f.lastRequest.Add(f.requestDelay))
	if wait <= 0 {
		f.lastRequest = time.Now()
		f.mu.Unlock()
		return ctx.Err()
	}
	f.mu.Unlock()

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		f.mu.Lock()
		f.lastRequest = time.Now()
		f.mu.Unlock()
		return ctx.Err()
	}
}

type digikalaParsedPrice struct {
	PriceIRR int64
	Status   string
}

func parseDigikalaPriceResponse(body []byte) (digikalaParsedPrice, error) {
	var payload struct {
		Data struct {
			Product struct {
				Status         string `json:"status"`
				DefaultVariant *struct {
					Price struct {
						SellingPrice int64 `json:"selling_price"`
					} `json:"price"`
				} `json:"default_variant"`
			} `json:"product"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&payload); err != nil {
		return digikalaParsedPrice{}, fmt.Errorf("parse digikala response: %w", err)
	}

	status := strings.TrimSpace(payload.Data.Product.Status)
	if status != "" && status != "marketable" {
		return digikalaParsedPrice{}, ErrProductUnavailable
	}
	if payload.Data.Product.DefaultVariant == nil {
		return digikalaParsedPrice{}, ErrProductUnavailable
	}

	priceIRR := payload.Data.Product.DefaultVariant.Price.SellingPrice
	if priceIRR <= 0 {
		return digikalaParsedPrice{}, ErrPriceNotFound
	}

	return digikalaParsedPrice{PriceIRR: priceIRR, Status: status}, nil
}

func digikalaProductID(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedSourceURL, err.Error())
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "digikala.com" && host != "www.digikala.com" {
		return "", ErrUnsupportedSourceURL
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for _, part := range parts {
		if strings.HasPrefix(strings.ToLower(part), "dkp-") {
			id := strings.TrimPrefix(strings.ToLower(part), "dkp-")
			if id == "" {
				return "", ErrUnsupportedSourceURL
			}
			for _, r := range id {
				if r < '0' || r > '9' {
					return "", ErrUnsupportedSourceURL
				}
			}
			return id, nil
		}
	}

	return "", ErrUnsupportedSourceURL
}
