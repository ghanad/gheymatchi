package price

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreCreateAndListByProduct(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database, "test-product")
	sourceID := createTestSource(t, database, productID, "test-source")
	store := NewSQLiteStore(database)
	ctx := context.Background()

	rawPayload := `{"source":"manual"}`
	newer := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	older := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	if _, err := store.Create(ctx, productID, sourceID, CreateInput{PriceIRR: 120000, CapturedAt: newer}); err != nil {
		t.Fatalf("Create() newer error = %v", err)
	}
	created, err := store.Create(ctx, productID, sourceID, CreateInput{
		PriceIRR:   110000,
		CapturedAt: older,
		RawPayload: &rawPayload,
	})
	if err != nil {
		t.Fatalf("Create() older error = %v", err)
	}
	if created.ProductSourceID != sourceID {
		t.Fatalf("ProductSourceID = %q, want %q", created.ProductSourceID, sourceID)
	}
	if created.RawPayload == nil || *created.RawPayload != rawPayload {
		t.Fatalf("RawPayload = %v, want %q", created.RawPayload, rawPayload)
	}

	pricePoints, err := store.ListByProduct(ctx, productID)
	if err != nil {
		t.Fatalf("ListByProduct() error = %v", err)
	}
	if len(pricePoints) != 2 {
		t.Fatalf("len(pricePoints) = %d, want 2", len(pricePoints))
	}
	if pricePoints[0].PriceIRR != 110000 || pricePoints[1].PriceIRR != 120000 {
		t.Fatalf("prices ordered = [%d, %d], want [110000, 120000]", pricePoints[0].PriceIRR, pricePoints[1].PriceIRR)
	}
}

func TestSQLiteStoreCreateRequiresMatchingSource(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database, "test-product")
	otherProductID := createTestProduct(t, database, "other-product")
	sourceID := createTestSource(t, database, otherProductID, "other-source")
	store := NewSQLiteStore(database)

	_, err := store.Create(context.Background(), productID, sourceID, CreateInput{PriceIRR: 1000})
	if err != ErrNotFound {
		t.Fatalf("Create() error = %v, want ErrNotFound", err)
	}
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	migrations, err := db.LoadMigrations(os.DirFS("../../migrations"))
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}
	if err := db.ApplyMigrations(ctx, database, migrations); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	return database
}

func createTestProduct(t *testing.T, database *sql.DB, productID string) string {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO products (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		productID,
		"Phone",
		"2026-01-01T00:00:00Z",
		"2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	return productID
}

func createTestSource(t *testing.T, database *sql.DB, productID string, sourceID string) string {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO product_sources (id, product_id, url, source_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		sourceID,
		productID,
		"https://example.com/p/"+sourceID,
		"manual",
		"2026-01-01T00:00:00Z",
		"2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert product source: %v", err)
	}
	return sourceID
}
