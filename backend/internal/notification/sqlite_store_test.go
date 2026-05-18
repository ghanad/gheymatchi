package notification

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gheymatchi/backend/internal/alert"
	"gheymatchi/backend/internal/db"
	"gheymatchi/backend/internal/price"
)

func TestSQLiteStoreCreateAlertTriggered(t *testing.T) {
	database := newTestDB(t)
	store := NewSQLiteStore(database)
	insertAlert(t, database)

	err := store.CreateAlertTriggered(context.Background(), alert.Alert{
		ID:        "alert-1",
		ProductID: "product-1",
	}, price.PricePoint{
		ID:         "price-1",
		ProductID:  "product-1",
		PriceIRR:   90,
		CapturedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateAlertTriggered() error = %v", err)
	}

	var channel string
	var recipient string
	var status string
	if err := database.QueryRow(`SELECT channel, recipient, status FROM notifications WHERE alert_id = ?`, "alert-1").Scan(&channel, &recipient, &status); err != nil {
		t.Fatalf("query notification: %v", err)
	}
	if channel != ChannelInternal {
		t.Fatalf("channel = %q, want %q", channel, ChannelInternal)
	}
	if recipient != "product-1" {
		t.Fatalf("recipient = %q, want product-1", recipient)
	}
	if status != StatusPending {
		t.Fatalf("status = %q, want %q", status, StatusPending)
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

func insertAlert(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(
		`INSERT INTO products (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"product-1",
		"Phone",
		"2026-01-01T00:00:00Z",
		"2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	_, err = database.Exec(
		`INSERT INTO alerts (id, product_id, name, condition_type, target_unit, threshold_value_text, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"alert-1",
		"product-1",
		"Target price",
		alert.ConditionBelow,
		alert.UnitIRR,
		"100",
		"2026-01-01T00:00:00Z",
		"2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert alert: %v", err)
	}
}
