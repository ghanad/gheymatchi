package source

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateSource(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database)

	mux := http.NewServeMux()
	NewHandler(NewSQLiteStore(database)).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products/"+productID+"/sources", strings.NewReader(`{"url":"https://www.digikala.com/product/dkp-123456/","source_name":"digikala"}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"source_name":"digikala"`) {
		t.Fatalf("body = %s, want source JSON", rec.Body.String())
	}
}

func TestHandlerRejectsInvalidURL(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database)

	mux := http.NewServeMux()
	NewHandler(NewSQLiteStore(database)).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products/"+productID+"/sources", strings.NewReader(`{"url":"ftp://example.com/p/1"}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"field":"url"`) {
		t.Fatalf("body = %s, want url validation error", rec.Body.String())
	}
}

func TestHandlerUpdatesSourceActiveState(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database)
	store := NewSQLiteStore(database)
	created, err := store.Create(httptest.NewRequest(http.MethodPost, "/", nil).Context(), productID, CreateInput{URL: "https://www.digikala.com/product/dkp-123456/"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPatch, "/api/products/"+productID+"/sources/"+created.ID, strings.NewReader(`{"is_active":false}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"is_active":false`) {
		t.Fatalf("body = %s, want inactive source", rec.Body.String())
	}
}
