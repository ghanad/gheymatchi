package product

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gheymatchi/backend/internal/auth"
	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreCRUD(t *testing.T) {
	store := newTestStore(t)
	ctx := auth.ContextWithUserID(context.Background(), "user-1")

	created, err := store.Create(ctx, "user-1", CreateInput{Name: "Phone", Description: "Digikala candidate"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("created ID is empty")
	}
	if created.Name != "Phone" {
		t.Fatalf("created name = %q, want Phone", created.Name)
	}
	if created.UserID == nil || *created.UserID != "user-1" {
		t.Fatalf("created user ID = %v, want user-1", created.UserID)
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

func TestSQLiteStoreScopesProductsByUser(t *testing.T) {
	store := newTestStore(t)
	ctxA := auth.ContextWithUserID(context.Background(), "user-a")
	ctxB := auth.ContextWithUserID(context.Background(), "user-b")

	created, err := store.Create(ctxA, "user-a", CreateInput{Name: "Private phone"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	products, err := store.List(ctxB)
	if err != nil {
		t.Fatalf("List(other user) error = %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("other user products = %d, want 0", len(products))
	}

	if _, err := store.Get(ctxB, created.ID); err != ErrNotFound {
		t.Fatalf("Get(other user) error = %v, want ErrNotFound", err)
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
	for _, userID := range []string{"user-1", "user-a", "user-b"} {
		if _, err := database.ExecContext(ctx, `
INSERT INTO users (id, email, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)`,
			userID,
			userID+"@example.com",
			"test-hash",
			"2026-01-01T00:00:00Z",
			"2026-01-01T00:00:00Z",
		); err != nil {
			t.Fatalf("insert user %s: %v", userID, err)
		}
	}

	return NewSQLiteStore(database)
}
