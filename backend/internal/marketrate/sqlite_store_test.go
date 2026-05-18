package marketrate

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreCreateLatestAndHistory(t *testing.T) {
	store := NewSQLiteStore(newTestDB(t))
	ctx := context.Background()

	firstObservedAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	secondObservedAt := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)

	first, err := store.Create(ctx, CreateInput{
		RateType:   RateTypeUSDIRR,
		ValueText:  "910000",
		ObservedAt: firstObservedAt,
	})
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	if first.Unit != UnitIRR {
		t.Fatalf("Unit = %q, want %q", first.Unit, UnitIRR)
	}

	second, err := store.Create(ctx, CreateInput{
		RateType:   RateTypeUSDIRR,
		ValueText:  "920000",
		ObservedAt: secondObservedAt,
	})
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	_, err = store.Create(ctx, CreateInput{
		RateType:   RateTypeGoldGramIRR,
		ValueText:  "65000000",
		ObservedAt: firstObservedAt,
	})
	if err != nil {
		t.Fatalf("Create() gold error = %v", err)
	}

	rateType := RateTypeUSDIRR
	latest, err := store.Latest(ctx, &rateType)
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if len(latest) != 1 {
		t.Fatalf("len(latest) = %d, want 1", len(latest))
	}
	if latest[0].ID != second.ID {
		t.Fatalf("latest ID = %q, want %q", latest[0].ID, second.ID)
	}

	allLatest, err := store.Latest(ctx, nil)
	if err != nil {
		t.Fatalf("Latest(nil) error = %v", err)
	}
	if len(allLatest) != 2 {
		t.Fatalf("len(allLatest) = %d, want 2", len(allLatest))
	}

	history, err := store.History(ctx, &rateType)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(history))
	}
	if history[0].ID != second.ID {
		t.Fatalf("history[0].ID = %q, want %q", history[0].ID, second.ID)
	}
}

func TestSQLiteStoreRejectsInvalidRate(t *testing.T) {
	store := NewSQLiteStore(newTestDB(t))

	_, err := store.Create(context.Background(), CreateInput{
		RateType:  RateTypeUSDIRR,
		ValueText: "0",
	})
	if err == nil {
		t.Fatal("Create() error = nil, want validation error")
	}
}

func TestSQLiteStoreLatestReturnsEmptyWhenMissing(t *testing.T) {
	store := NewSQLiteStore(newTestDB(t))
	rateType := RateTypeUSDIRR

	latest, err := store.Latest(context.Background(), &rateType)
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if len(latest) != 0 {
		t.Fatalf("len(latest) = %d, want 0", len(latest))
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
