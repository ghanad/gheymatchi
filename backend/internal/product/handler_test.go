package product

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateProduct(t *testing.T) {
	store := newTestStore(t)
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{"name":"Phone"}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if !strings.Contains(rec.Body.String(), `"name":"Phone"`) {
		t.Fatalf("body = %s, want product JSON", rec.Body.String())
	}
}

func TestHandlerRejectsInvalidProduct(t *testing.T) {
	store := newTestStore(t)
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{"name":" "}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_input"`) {
		t.Fatalf("body = %s, want invalid_input error", rec.Body.String())
	}
}

func TestHandlerDeletesProduct(t *testing.T) {
	store := newTestStore(t)
	product, err := store.Create(httptest.NewRequest(http.MethodPost, "/", nil).Context(), CreateInput{Name: "Phone"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/"+product.ID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
