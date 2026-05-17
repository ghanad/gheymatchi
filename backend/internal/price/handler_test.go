package price

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreatePricePoint(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database, "test-product")
	sourceID := createTestSource(t, database, productID, "test-source")

	mux := http.NewServeMux()
	NewHandler(NewSQLiteStore(database)).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products/"+productID+"/sources/"+sourceID+"/price-points", strings.NewReader(`{"price_irr":123000,"captured_at":"2026-01-01T10:00:00Z"}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"price_irr":123000`) {
		t.Fatalf("body = %s, want price point JSON", rec.Body.String())
	}
}

func TestHandlerRejectsInvalidPrice(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database, "test-product")
	sourceID := createTestSource(t, database, productID, "test-source")

	mux := http.NewServeMux()
	NewHandler(NewSQLiteStore(database)).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products/"+productID+"/sources/"+sourceID+"/price-points", strings.NewReader(`{"price_irr":0}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"field":"price_irr"`) {
		t.Fatalf("body = %s, want price_irr validation error", rec.Body.String())
	}
}

func TestHandlerListPricePoints(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database, "test-product")
	sourceID := createTestSource(t, database, productID, "test-source")
	store := NewSQLiteStore(database)
	if _, err := store.Create(httptest.NewRequest(http.MethodPost, "/", nil).Context(), productID, sourceID, CreateInput{PriceIRR: 123000}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/products/"+productID+"/price-points", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"price_points":[`) {
		t.Fatalf("body = %s, want price_points wrapper", rec.Body.String())
	}
}
