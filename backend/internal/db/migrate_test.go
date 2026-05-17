package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyMigrationsCanRunRepeatedly(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "test.db")

	database, err := Open(ctx, databasePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.Close()

	migrations, err := LoadMigrations(os.DirFS("../../migrations"))
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("LoadMigrations() returned no migrations")
	}

	for i := 0; i < 2; i++ {
		if err := ApplyMigrations(ctx, database, migrations); err != nil {
			t.Fatalf("ApplyMigrations() run %d error = %v", i+1, err)
		}
	}

	var tableName string
	if err := database.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'products'").Scan(&tableName); err != nil {
		t.Fatalf("products table missing: %v", err)
	}
}
