package price

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"gheymatchi/backend/internal/source"
)

const (
	digikalaAPIBaseURL      = "https://api.digikala.com/v2/product/"
	digikalaFreshAPIBaseURL = "https://api.digikala.com/fresh/v1/product/"
)

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
	sourceInfo, err := parseDigikalaSourceURL(productSource.URL)
	if err != nil {
		return FetchResult{}, err
	}

	if err := f.waitForRateLimit(ctx); err != nil {
		return FetchResult{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	apiURL := digikalaAPIURL(sourceInfo)
	body, err := f.fetchJSON(requestCtx, apiURL)
	if err != nil {
		return FetchResult{}, err
	}

	if !sourceInfo.IsFresh && hasDigikalaFreshRedirect(body) {
		sourceInfo.IsFresh = true
		apiURL = digikalaAPIURL(sourceInfo)
		body, err = f.fetchJSON(requestCtx, apiURL)
		if err != nil {
			return FetchResult{}, err
		}
	}

	parsed, err := parseDigikalaPriceResponseForVariant(body, sourceInfo.VariantID)
	if err != nil {
		return FetchResult{}, err
	}

	capturedAt := time.Now().UTC()
	rawPayload, err := json.Marshal(map[string]any{
		"provider":     "digikala",
		"source_id":    productSource.ID,
		"url":          productSource.URL,
		"api_url":      apiURL,
		"product_id":   sourceInfo.ProductID,
		"variant_id":   parsed.VariantID,
		"price_irr":    parsed.PriceIRR,
		"status":       parsed.Status,
		"availability": parsed.Availability,
		"seller":       parsed.Seller,
		"title":        parsed.Title,
		"captured_at":  capturedAt.Format(time.RFC3339),
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

func (f *DigikalaFetcher) fetchJSON(ctx context.Context, apiURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build digikala request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch digikala product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, ErrProductUnavailable
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%w: digikala returned status %d", ErrSourceAccessDenied, resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch digikala product: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read digikala response: %w", err)
	}
	return body, nil
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
	PriceIRR     int64
	Status       string
	Availability string
	Seller       string
	Title        string
	VariantID    string
}

func parseDigikalaPriceResponse(body []byte) (digikalaParsedPrice, error) {
	return parseDigikalaPriceResponseForVariant(body, "")
}

func parseDigikalaPriceResponseForVariant(body []byte, requestedVariantID string) (digikalaParsedPrice, error) {
	var payload struct {
		Data struct {
			Product *digikalaProduct `json:"product"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return digikalaParsedPrice{}, fmt.Errorf("parse digikala response: %w", err)
	}
	if payload.Data.Product == nil {
		return digikalaParsedPrice{}, ErrProductUnavailable
	}

	variant, err := selectDigikalaVariant(payload.Data.Product, requestedVariantID)
	if err != nil {
		return digikalaParsedPrice{}, err
	}

	availability := digikalaAvailability(payload.Data.Product, variant)
	if availability == "out_of_stock" {
		return digikalaParsedPrice{}, ErrProductUnavailable
	}

	priceIRR, ok := digikalaVariantPrice(variant.Price)
	if !ok || priceIRR <= 0 {
		return digikalaParsedPrice{}, ErrPriceNotFound
	}

	return digikalaParsedPrice{
		PriceIRR:     priceIRR,
		Status:       firstNonEmpty(variant.Status, payload.Data.Product.Status),
		Availability: availability,
		Seller:       digikalaSeller(variant.Seller),
		Title:        firstNonEmpty(payload.Data.Product.TitleFA, payload.Data.Product.TitleEN, payload.Data.Product.TestTitleFA),
		VariantID:    rawNumericString(variant.ID),
	}, nil
}

type digikalaProduct struct {
	Status         string            `json:"status"`
	TitleFA        string            `json:"title_fa"`
	TitleEN        string            `json:"title_en"`
	TestTitleFA    string            `json:"test_title_fa"`
	DefaultVariant *digikalaVariant  `json:"default_variant"`
	Variants       []digikalaVariant `json:"variants"`
	DataLayer      digikalaDataLayer `json:"data_layer"`
}

type digikalaDataLayer struct {
	Dimension20 string `json:"dimension20"`
}

type digikalaVariant struct {
	ID     json.RawMessage `json:"id"`
	Status string          `json:"status"`
	Price  json.RawMessage `json:"price"`
	Seller json.RawMessage `json:"seller"`
}

func selectDigikalaVariant(product *digikalaProduct, requestedVariantID string) (digikalaVariant, error) {
	if requestedVariantID != "" {
		if product.DefaultVariant != nil && rawNumericString(product.DefaultVariant.ID) == requestedVariantID {
			return *product.DefaultVariant, nil
		}
		for _, variant := range product.Variants {
			if rawNumericString(variant.ID) == requestedVariantID {
				return variant, nil
			}
		}
		return digikalaVariant{}, ErrProductUnavailable
	}

	if product.DefaultVariant != nil {
		return *product.DefaultVariant, nil
	}
	for _, variant := range product.Variants {
		if normalizedDigikalaStatus(variant.Status) == "in_stock" {
			return variant, nil
		}
	}
	if len(product.Variants) > 0 {
		return product.Variants[0], nil
	}
	return digikalaVariant{}, ErrProductUnavailable
}

func digikalaVariantPrice(raw json.RawMessage) (int64, bool) {
	if value, ok := rawInt64(raw); ok {
		return value, true
	}

	var price map[string]json.RawMessage
	if err := json.Unmarshal(raw, &price); err != nil {
		return 0, false
	}
	for _, key := range []string{"selling_price", "final_price", "rrp_price"} {
		if value, ok := rawInt64(price[key]); ok {
			return value, true
		}
	}
	return 0, false
}

func digikalaSeller(raw json.RawMessage) string {
	var sellerText string
	if err := json.Unmarshal(raw, &sellerText); err == nil {
		return strings.TrimSpace(sellerText)
	}

	var seller map[string]json.RawMessage
	if err := json.Unmarshal(raw, &seller); err != nil {
		return ""
	}
	for _, key := range []string{"title", "name", "title_fa"} {
		var value string
		if err := json.Unmarshal(seller[key], &value); err == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func digikalaAvailability(product *digikalaProduct, variant digikalaVariant) string {
	for _, status := range []string{variant.Status, product.Status, product.DataLayer.Dimension20} {
		if normalized := normalizedDigikalaStatus(status); normalized != "" {
			return normalized
		}
	}
	return "unknown"
}

func normalizedDigikalaStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "marketable", "available", "active":
		return "in_stock"
	case "not_marketable", "out_of_stock", "unavailable", "inactive":
		return "out_of_stock"
	default:
		return ""
	}
}

func rawInt64(raw json.RawMessage) (int64, bool) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return 0, false
	}
	if strings.HasPrefix(trimmed, `"`) {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return 0, false
		}
		trimmed = strings.TrimSpace(value)
	}

	value, err := strconv.ParseInt(trimmed, 10, 64)
	return value, err == nil
}

func rawNumericString(raw json.RawMessage) string {
	value, ok := rawInt64(raw)
	if !ok {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func hasDigikalaFreshRedirect(body []byte) bool {
	var payload struct {
		RedirectURL struct {
			URI string `json:"uri"`
		} `json:"redirect_url"`
		Data struct {
			Product *json.RawMessage `json:"product"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&payload); err != nil {
		return false
	}
	return payload.Data.Product == nil && strings.Contains(payload.RedirectURL.URI, "/fresh/product/")
}

type digikalaSourceInfo struct {
	ProductID string
	VariantID string
	IsFresh   bool
}

func digikalaAPIURL(info digikalaSourceInfo) string {
	baseURL := digikalaAPIBaseURL
	if info.IsFresh {
		baseURL = digikalaFreshAPIBaseURL
	}
	return baseURL + url.PathEscape(info.ProductID) + "/"
}

func parseDigikalaSourceURL(rawURL string) (digikalaSourceInfo, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return digikalaSourceInfo{}, fmt.Errorf("%w: %s", ErrUnsupportedSourceURL, err.Error())
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "digikala.com" && host != "www.digikala.com" {
		return digikalaSourceInfo{}, ErrUnsupportedSourceURL
	}

	productID := digikalaProductIDFromPath(parsed.Path)
	if productID == "" {
		return digikalaSourceInfo{}, ErrUnsupportedSourceURL
	}

	variantID := parsed.Query().Get("variant_id")
	if !isNumeric(variantID) {
		variantID = ""
	}

	return digikalaSourceInfo{
		ProductID: productID,
		VariantID: variantID,
		IsFresh:   strings.Contains(strings.ToLower(parsed.Path), "/fresh/"),
	}, nil
}

func digikalaProductID(rawURL string) (string, error) {
	info, err := parseDigikalaSourceURL(rawURL)
	if err != nil {
		return "", err
	}
	return info.ProductID, nil
}

func digikalaProductIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for index, part := range parts {
		lower := strings.ToLower(part)
		if strings.HasPrefix(lower, "dkp-") {
			id := strings.TrimPrefix(lower, "dkp-")
			if isNumeric(id) {
				return id
			}
			return ""
		}
		if lower == "product" && index+1 < len(parts) && isNumeric(parts[index+1]) {
			return parts[index+1]
		}
	}
	return ""
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
