package source

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreCRUD(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database)
	store := NewSQLiteStore(database)
	ctx := context.Background()

	created, err := store.Create(ctx, productID, CreateInput{
		URL:        "https://www.digikala.com/product/dkp-123456/",
		SourceName: "Digikala",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("created ID is empty")
	}
	if created.ProductID != productID {
		t.Fatalf("ProductID = %q, want %q", created.ProductID, productID)
	}
	if created.SourceName != "digikala" {
		t.Fatalf("SourceName = %q, want digikala", created.SourceName)
	}
	if !created.IsActive {
		t.Fatal("IsActive = false, want true")
	}

	sources, err := store.List(ctx, productID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("len(sources) = %d, want 1", len(sources))
	}

	active := false
	name := "digikala"
	updated, err := store.Update(ctx, productID, created.ID, UpdateInput{
		SourceName: &name,
		IsActive:   &active,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.SourceName != "digikala" {
		t.Fatalf("updated SourceName = %q, want digikala", updated.SourceName)
	}
	if updated.IsActive {
		t.Fatal("updated IsActive = true, want false")
	}

	if err := store.Delete(ctx, productID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	sources, err = store.List(ctx, productID)
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("len(sources) after delete = %d, want 0", len(sources))
	}
}

func TestSQLiteStoreCreateRequiresProduct(t *testing.T) {
	store := NewSQLiteStore(newTestDB(t))

	_, err := store.Create(context.Background(), "missing", CreateInput{URL: "https://example.com"})
	if err != ErrNotFound {
		t.Fatalf("Create() error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreListActive(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database)
	store := NewSQLiteStore(database)
	ctx := context.Background()

	active, err := store.Create(ctx, productID, CreateInput{URL: "https://www.digikala.com/product/dkp-123456/"})
	if err != nil {
		t.Fatalf("Create() active error = %v", err)
	}
	inactiveFlag := false
	if _, err := store.Create(ctx, productID, CreateInput{URL: "https://www.digikala.com/product/dkp-654321/", IsActive: &inactiveFlag}); err != nil {
		t.Fatalf("Create() inactive error = %v", err)
	}

	sources, err := store.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive() error = %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("len(sources) = %d, want 1", len(sources))
	}
	if sources[0].ID != active.ID {
		t.Fatalf("active source ID = %q, want %q", sources[0].ID, active.ID)
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

func createTestProduct(t *testing.T, database *sql.DB) string {
	t.Helper()

	const productID = "test-product"
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
