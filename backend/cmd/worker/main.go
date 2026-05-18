package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gheymatchi/backend/internal/alert"
	"gheymatchi/backend/internal/config"
	"gheymatchi/backend/internal/crawl"
	"gheymatchi/backend/internal/db"
	"gheymatchi/backend/internal/notification"
	"gheymatchi/backend/internal/price"
	"gheymatchi/backend/internal/source"
	"gheymatchi/backend/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	database, err := db.Open(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("open database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := worker.NewRunner(
		source.NewSQLiteStore(database),
		price.NewSQLiteStore(database),
		crawl.NewSQLiteStore(database),
		alert.NewEvaluator(alert.NewSQLiteStore(database), notification.NewSQLiteStore(database)),
		price.NewMockPriceFetcher(),
		logger,
	)

	logger.Info("worker started", slog.Duration("interval", cfg.WorkerInterval), slog.String("database_path", cfg.DatabasePath))
	if err := runner.Run(runCtx, cfg.WorkerInterval); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("worker stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("worker stopped")
}
