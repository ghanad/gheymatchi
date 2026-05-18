package marketrate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateMarketRate(t *testing.T) {
	store := newMemoryStore(t)
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/market-rates", strings.NewReader(`{
		"rate_type":"USD_IRR",
		"value_text":"920000",
		"observed_at":"2026-01-01T10:00:00Z"
	}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"rate_type":"USD_IRR"`) {
		t.Fatalf("body = %s, want market rate JSON", rec.Body.String())
	}
}

func TestHandlerRejectsInvalidMarketRate(t *testing.T) {
	store := newMemoryStore(t)
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/market-rates", strings.NewReader(`{
		"rate_type":"EUR_IRR",
		"value_text":"920000"
	}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"field":"rate_type"`) {
		t.Fatalf("body = %s, want rate_type validation error", rec.Body.String())
	}
}

func TestHandlerReturnsLatestAndHistory(t *testing.T) {
	store := newMemoryStore(t)
	_, err := store.Create(context.Background(), CreateInput{
		RateType:  RateTypeUSDIRR,
		ValueText: "920000",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	for _, path := range []string{"/api/market-rates/latest", "/api/market-rates/history?rate_type=USD_IRR"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d; body=%s", path, rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"market_rates"`) {
			t.Fatalf("%s body = %s, want market_rates response", path, rec.Body.String())
		}
	}
}

func newMemoryStore(t *testing.T) *SQLiteStore {
	t.Helper()
	return NewSQLiteStore(newTestDB(t))
}
