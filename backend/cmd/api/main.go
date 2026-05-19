package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gheymatchi/backend/internal/alert"
	"gheymatchi/backend/internal/auth"
	"gheymatchi/backend/internal/config"
	"gheymatchi/backend/internal/db"
	"gheymatchi/backend/internal/marketrate"
	"gheymatchi/backend/internal/notification"
	"gheymatchi/backend/internal/price"
	"gheymatchi/backend/internal/product"
	"gheymatchi/backend/internal/source"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ReadTimeout)
	defer cancel()

	database, err := db.Open(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("open database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	if err := run(cfg, logger, database); err != nil {
		logger.Error("api stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger, database *sql.DB) error {
	mux := http.NewServeMux()
	authStore := auth.NewSQLiteStore(database)
	mux.HandleFunc("/healthz", statusHandler("ok"))
	mux.HandleFunc("/readyz", readinessHandler(database))
	auth.NewHandler(authStore).Register(mux)
	product.NewHandler(product.NewSQLiteStore(database)).Register(mux)
	source.NewHandler(source.NewSQLiteStore(database)).Register(mux)
	price.NewHandler(price.NewSQLiteStore(database)).Register(mux)
	alert.NewHandler(alert.NewSQLiteStore(database)).Register(mux)
	marketrate.NewHandler(marketrate.NewSQLiteStore(database)).Register(mux)
	notification.NewHandler(notification.NewSQLiteStore(database)).Register(mux)

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      requestLogger(logger, corsMiddleware(auth.Middleware(authStore, mux))),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api listening", slog.String("addr", cfg.HTTPAddr), slog.String("env", cfg.Env))
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown requested")
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("shutdown complete")
	return nil
}

func statusHandler(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": status})
	}
}

func readinessHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("write response", slog.String("error", err.Error()))
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(lrw, r)

		logger.Info(
			"http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", lrw.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigins[r.Header.Get("Origin")] {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
