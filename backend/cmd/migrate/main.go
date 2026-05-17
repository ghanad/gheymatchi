package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"gheymatchi/backend/internal/config"
	"gheymatchi/backend/internal/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.Open(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("open database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	migrations, err := db.LoadMigrations(os.DirFS("migrations"))
	if err != nil {
		logger.Error("load migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := db.ApplyMigrations(ctx, database, migrations); err != nil {
		logger.Error("apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("migrations applied", slog.Int("count", len(migrations)), slog.String("database_path", cfg.DatabasePath))
}
