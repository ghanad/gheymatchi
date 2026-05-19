package alert

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"gheymatchi/backend/internal/auth"
	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreCRUD(t *testing.T) {
	database := newTestDB(t)
	productID := createTestProduct(t, database)
	store := NewSQLiteStore(database)
	ctx := context.Background()

	created, err := store.Create(ctx, productID, CreateInput{
		Name:               "Target price",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitIRR,
		ThresholdValueText: "85000000",
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
	if !created.IsActive {
		t.Fatal("IsActive = false, want true")
	}

	alerts, err := store.List(ctx, productID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("len(alerts) = %d, want 1", len(alerts))
	}

	active := false
	threshold := "900"
	unit := UnitUSD
	updated, err := store.Update(ctx, productID, created.ID, UpdateInput{
		TargetUnit:         &unit,
		ThresholdValueText: &threshold,
		IsActive:           &active,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.TargetUnit != UnitUSD {
		t.Fatalf("TargetUnit = %q, want %q", updated.TargetUnit, UnitUSD)
	}
	if updated.ThresholdValueText != "900" {
		t.Fatalf("ThresholdValueText = %q, want 900", updated.ThresholdValueText)
	}
	if updated.IsActive {
		t.Fatal("IsActive = true, want false")
	}

	if err := store.Delete(ctx, productID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	alerts, err = store.List(ctx, productID)
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("len(alerts) after delete = %d, want 0", len(alerts))
	}
}

func TestSQLiteStoreCreateRequiresProduct(t *testing.T) {
	store := NewSQLiteStore(newTestDB(t))

	_, err := store.Create(context.Background(), "missing", CreateInput{
		Name:               "Target price",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitIRR,
		ThresholdValueText: "10",
	})
	if err != ErrNotFound {
		t.Fatalf("Create() error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreScopesAlertsByAuthenticatedProductOwner(t *testing.T) {
	database := newTestDB(t)
	insertTestUser(t, database, "user-a")
	insertTestUser(t, database, "user-b")
	productID := createTestProductForUser(t, database, "private-product", "user-a")
	store := NewSQLiteStore(database)
	ctxA := auth.ContextWithUserID(context.Background(), "user-a")
	ctxB := auth.ContextWithUserID(context.Background(), "user-b")

	created, err := store.Create(ctxA, productID, CreateInput{
		Name:               "Target price",
		ConditionType:      ConditionBelow,
		TargetUnit:         UnitIRR,
		ThresholdValueText: "85000000",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.UserID == nil || *created.UserID != "user-a" {
		t.Fatalf("created user ID = %v, want user-a", created.UserID)
	}

	if _, err := store.List(ctxB, productID); err != ErrNotFound {
		t.Fatalf("List(other user) error = %v, want ErrNotFound", err)
	}
	if err := store.Delete(ctxB, productID, created.ID); err != ErrNotFound {
		t.Fatalf("Delete(other user) error = %v, want ErrNotFound", err)
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

func insertTestUser(t *testing.T, database *sql.DB, userID string) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO users (id, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		userID,
		userID+"@example.com",
		"test-hash",
		"2026-01-01T00:00:00Z",
		"2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func createTestProductForUser(t *testing.T, database *sql.DB, productID string, userID string) string {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO products (id, user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		productID,
		userID,
		"Phone",
		"2026-01-01T00:00:00Z",
		"2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	return productID
}
