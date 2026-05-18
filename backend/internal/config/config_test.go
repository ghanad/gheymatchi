package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("WORKER_INTERVAL", "")
	t.Setenv("PRICE_FETCHER", "")
	t.Setenv("PRICE_FETCH_TIMEOUT", "")
	t.Setenv("PRICE_FETCH_DELAY", "")
	t.Setenv("HTTP_READ_TIMEOUT", "")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("HTTP_IDLE_TIMEOUT", "")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "local" {
		t.Fatalf("Env = %q, want local", cfg.Env)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.DatabasePath != "data/gheymatchi.db" {
		t.Fatalf("DatabasePath = %q, want data/gheymatchi.db", cfg.DatabasePath)
	}
	if cfg.WorkerInterval != 5*time.Minute {
		t.Fatalf("WorkerInterval = %s, want 5m", cfg.WorkerInterval)
	}
	if cfg.PriceFetcher != "mock" {
		t.Fatalf("PriceFetcher = %q, want mock", cfg.PriceFetcher)
	}
	if cfg.PriceFetchTimeout != 10*time.Second {
		t.Fatalf("PriceFetchTimeout = %s, want 10s", cfg.PriceFetchTimeout)
	}
	if cfg.PriceFetchDelay != 2*time.Second {
		t.Fatalf("PriceFetchDelay = %s, want 2s", cfg.PriceFetchDelay)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
