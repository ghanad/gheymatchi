package alert

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateAlert(t *testing.T) {
	store := newMemoryStore(t)
	productID := createTestProduct(t, store.db)
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products/"+productID+"/alerts", strings.NewReader(`{
		"name":"Target price",
		"condition_type":"BELOW",
		"target_unit":"IRR",
		"threshold_value_text":"85000000"
	}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"target_unit":"IRR"`) {
		t.Fatalf("body = %s, want alert JSON", rec.Body.String())
	}
}

func TestHandlerRejectsInvalidAlert(t *testing.T) {
	store := newMemoryStore(t)
	productID := createTestProduct(t, store.db)
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/products/"+productID+"/alerts", strings.NewReader(`{
		"name":"Target price",
		"condition_type":"EQUALS",
		"target_unit":"IRR",
		"threshold_value_text":"85000000"
	}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"field":"condition_type"`) {
		t.Fatalf("body = %s, want condition_type validation error", rec.Body.String())
	}
}

func TestHandlerDeletesAlert(t *testing.T) {
	store := newMemoryStore(t)
	productID := createTestProduct(t, store.db)
	created, err := store.Create(httptest.NewRequest(http.MethodPost, "/", nil).Context(), productID, CreateInput{
		Name:               "Target price",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitIRR,
		ThresholdValueText: "85000000",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(store).Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/"+productID+"/alerts/"+created.ID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func newMemoryStore(t *testing.T) *SQLiteStore {
	t.Helper()
	return NewSQLiteStore(newTestDB(t))
}
