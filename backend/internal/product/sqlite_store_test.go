package product

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreCRUD(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, CreateInput{Name: "Phone", Description: "Digikala candidate"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("created ID is empty")
	}
	if created.Name != "Phone" {
		t.Fatalf("created name = %q, want Phone", created.Name)
	}

	products, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("len(products) = %d, want 1", len(products))
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Description != "Digikala candidate" {
		t.Fatalf("description = %q, want Digikala candidate", got.Description)
	}

	name := "Updated Phone"
	description := ""
	updated, err := store.Update(ctx, created.ID, UpdateInput{Name: &name, Description: &description})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != name {
		t.Fatalf("updated name = %q, want %q", updated.Name, name)
	}
	if updated.Description != "" {
		t.Fatalf("updated description = %q, want empty", updated.Description)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(ctx, created.ID); err != ErrNotFound {
		t.Fatalf("Get() after delete error = %v, want ErrNotFound", err)
	}
}

func newTestStore(t *testing.T) *SQLiteStore {
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

	return NewSQLiteStore(database)
}
