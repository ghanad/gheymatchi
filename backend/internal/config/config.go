package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env                     string
	HTTPAddr                string
	DatabasePath            string
	WorkerInterval          time.Duration
	PriceFetcher            string
	PriceFetchTimeout       time.Duration
	PriceFetchDelay         time.Duration
	NotificationProvider    string
	NotificationEmailTo     string
	NotificationMaxAttempts int
	SMTPHost                string
	SMTPPort                string
	SMTPUsername            string
	SMTPPassword            string
	SMTPFrom                string
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	IdleTimeout             time.Duration
	ShutdownTimeout         time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:                     getenv("APP_ENV", "local"),
		HTTPAddr:                getenv("HTTP_ADDR", ":8080"),
		DatabasePath:            getenv("DB_PATH", "data/gheymatchi.db"),
		WorkerInterval:          5 * time.Minute,
		PriceFetcher:            getenv("PRICE_FETCHER", "mock"),
		PriceFetchTimeout:       10 * time.Second,
		PriceFetchDelay:         2 * time.Second,
		NotificationProvider:    getenv("NOTIFICATION_PROVIDER", "dry_run"),
		NotificationEmailTo:     getenv("NOTIFICATION_EMAIL_TO", ""),
		NotificationMaxAttempts: 3,
		SMTPHost:                getenv("SMTP_HOST", ""),
		SMTPPort:                getenv("SMTP_PORT", "587"),
		SMTPUsername:            getenv("SMTP_USERNAME", ""),
		SMTPPassword:            getenv("SMTP_PASSWORD", ""),
		SMTPFrom:                getenv("SMTP_FROM", ""),
		ReadTimeout:             5 * time.Second,
		WriteTimeout:            10 * time.Second,
		IdleTimeout:             60 * time.Second,
		ShutdownTimeout:         10 * time.Second,
	}

	var err error
	if cfg.WorkerInterval, err = durationFromEnv("WORKER_INTERVAL", cfg.WorkerInterval); err != nil {
		return Config{}, err
	}
	if cfg.PriceFetchTimeout, err = durationFromEnv("PRICE_FETCH_TIMEOUT", cfg.PriceFetchTimeout); err != nil {
		return Config{}, err
	}
	if cfg.PriceFetchDelay, err = durationFromEnv("PRICE_FETCH_DELAY", cfg.PriceFetchDelay); err != nil {
		return Config{}, err
	}
	if cfg.NotificationMaxAttempts, err = intFromEnv("NOTIFICATION_MAX_ATTEMPTS", cfg.NotificationMaxAttempts); err != nil {
		return Config{}, err
	}
	if cfg.NotificationMaxAttempts <= 0 {
		return Config{}, fmt.Errorf("NOTIFICATION_MAX_ATTEMPTS must be greater than zero")
	}
	if cfg.ReadTimeout, err = durationFromEnv("HTTP_READ_TIMEOUT", cfg.ReadTimeout); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = durationFromEnv("HTTP_WRITE_TIMEOUT", cfg.WriteTimeout); err != nil {
		return Config{}, err
	}
	if cfg.IdleTimeout, err = durationFromEnv("HTTP_IDLE_TIMEOUT", cfg.IdleTimeout); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = durationFromEnv("HTTP_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return duration, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
