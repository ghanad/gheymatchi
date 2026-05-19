package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"gheymatchi/backend/internal/alert"
	"gheymatchi/backend/internal/auth"
	"gheymatchi/backend/internal/config"
	"gheymatchi/backend/internal/db"
	"gheymatchi/backend/internal/marketrate"
	"gheymatchi/backend/internal/price"
	"gheymatchi/backend/internal/product"
	"gheymatchi/backend/internal/source"
)

const demoProductName = "Demo Phone"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.OpenConfigured(ctx, cfg.DatabaseDriver, cfg.DatabasePath, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	migrationDir := "migrations"
	if cfg.DatabaseDriver == string(db.DriverPostgres) {
		migrationDir = "postgres_migrations"
	}
	migrations, err := db.LoadMigrations(os.DirFS(migrationDir))
	if err != nil {
		logger.Error("load migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := db.ApplyMigrationsForDriver(ctx, database, db.Driver(cfg.DatabaseDriver), migrations); err != nil {
		logger.Error("apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	driver := db.Driver(cfg.DatabaseDriver)
	productStore := product.NewStore(database, driver)
	sourceStore := source.NewStore(database, driver)
	priceStore := price.NewStore(database, driver)
	rateStore := marketrate.NewStore(database, driver)
	alertStore := alert.NewStore(database, driver)
	authStore := auth.NewStore(database, driver)

	demoUser, err := ensureDemoUser(ctx, authStore)
	if err != nil {
		logger.Error("seed user", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ctx = auth.ContextWithUserID(ctx, demoUser.ID)

	demoProduct, createdProduct, err := ensureDemoProduct(ctx, productStore)
	if err != nil {
		logger.Error("seed product", slog.String("error", err.Error()))
		os.Exit(1)
	}

	demoSource, createdSource, err := ensureDemoSource(ctx, sourceStore, demoProduct.ID)
	if err != nil {
		logger.Error("seed product source", slog.String("error", err.Error()))
		os.Exit(1)
	}

	createdRates, err := ensureDemoRates(ctx, rateStore)
	if err != nil {
		logger.Error("seed market rates", slog.String("error", err.Error()))
		os.Exit(1)
	}

	createdPrices, err := ensureDemoPrices(ctx, priceStore, demoProduct.ID, demoSource.ID)
	if err != nil {
		logger.Error("seed price points", slog.String("error", err.Error()))
		os.Exit(1)
	}

	createdAlerts, err := ensureDemoAlert(ctx, alertStore, demoProduct.ID)
	if err != nil {
		logger.Error("seed alert", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info(
		"seed complete",
		slog.String("database_driver", cfg.DatabaseDriver),
		slog.Bool("created_product", createdProduct),
		slog.Bool("created_source", createdSource),
		slog.Int("created_market_rates", createdRates),
		slog.Int("created_price_points", createdPrices),
		slog.Int("created_alerts", createdAlerts),
	)
}

func ensureDemoUser(ctx context.Context, store auth.Store) (auth.User, error) {
	session, err := store.Login(ctx, auth.LoginInput{Email: "demo@gheymatchi.local", Password: "password123"})
	if err == nil {
		return session.User, nil
	}
	if err != auth.ErrInvalidCredentials {
		return auth.User{}, err
	}

	created, err := store.Register(ctx, auth.RegisterInput{Email: "demo@gheymatchi.local", Password: "password123"})
	if err != nil {
		return auth.User{}, err
	}
	return created.User, nil
}

func ensureDemoProduct(ctx context.Context, store product.Store) (product.Product, bool, error) {
	products, err := store.List(ctx)
	if err != nil {
		return product.Product{}, false, err
	}
	for _, item := range products {
		if item.Name == demoProductName {
			return item, false, nil
		}
	}

	userID, err := auth.RequireUserID(ctx)
	if err != nil {
		return product.Product{}, false, err
	}
	created, err := store.Create(ctx, userID, product.CreateInput{
		Name:        demoProductName,
		Description: "Local seed product for testing price history, market rates, alerts, and notifications.",
	})
	if err != nil {
		return product.Product{}, false, err
	}
	return created, true, nil
}

func ensureDemoSource(ctx context.Context, store source.Store, productID string) (source.ProductSource, bool, error) {
	const demoURL = "https://www.digikala.com/product/dkp-123456/demo-phone/"

	sources, err := store.List(ctx, productID)
	if err != nil {
		return source.ProductSource{}, false, err
	}
	for _, item := range sources {
		if item.URL == demoURL {
			return item, false, nil
		}
	}

	active := true
	created, err := store.Create(ctx, productID, source.CreateInput{
		URL:        demoURL,
		SourceName: "digikala",
		IsActive:   &active,
	})
	if err != nil {
		return source.ProductSource{}, false, err
	}
	return created, true, nil
}

func ensureDemoRates(ctx context.Context, store marketrate.Store) (int, error) {
	now := time.Now().UTC().Truncate(time.Second)
	created := 0

	if count, err := ensureRate(ctx, store, marketrate.RateTypeUSDIRR, "620000", now.Add(-96*time.Hour)); err != nil {
		return 0, err
	} else {
		created += count
	}
	if count, err := ensureRate(ctx, store, marketrate.RateTypeGoldGramIRR, "35000000", now.Add(-96*time.Hour)); err != nil {
		return 0, err
	} else {
		created += count
	}

	return created, nil
}

func ensureRate(ctx context.Context, store marketrate.Store, rateType string, valueText string, observedAt time.Time) (int, error) {
	rates, err := store.Latest(ctx, &rateType)
	if err != nil {
		return 0, err
	}
	if len(rates) > 0 {
		return 0, nil
	}

	if _, err := store.Create(ctx, marketrate.CreateInput{
		RateType:   rateType,
		ValueText:  valueText,
		ObservedAt: observedAt,
	}); err != nil {
		return 0, err
	}
	return 1, nil
}

func ensureDemoPrices(ctx context.Context, store price.Store, productID string, sourceID string) (int, error) {
	existing, err := store.ListByProduct(ctx, productID)
	if err != nil {
		return 0, err
	}
	if len(existing) > 0 {
		return 0, nil
	}

	now := time.Now().UTC().Truncate(time.Second)
	rawPayload := `{"seed":true,"source":"local"}`
	inputs := []price.CreateInput{
		{PriceIRR: 84000000, CapturedAt: now.Add(-72 * time.Hour), RawPayload: &rawPayload},
		{PriceIRR: 81500000, CapturedAt: now.Add(-48 * time.Hour), RawPayload: &rawPayload},
		{PriceIRR: 82900000, CapturedAt: now.Add(-24 * time.Hour), RawPayload: &rawPayload},
	}

	for _, input := range inputs {
		if _, err := store.Create(ctx, productID, sourceID, input); err != nil {
			return 0, err
		}
	}
	return len(inputs), nil
}

func ensureDemoAlert(ctx context.Context, store alert.Store, productID string) (int, error) {
	alerts, err := store.List(ctx, productID)
	if err != nil {
		return 0, err
	}
	if len(alerts) > 0 {
		return 0, nil
	}

	active := true
	if _, err := store.Create(ctx, productID, alert.CreateInput{
		Name:               "Demo target price",
		ConditionType:      alert.ConditionBelow,
		TargetUnit:         alert.UnitIRR,
		ThresholdValueText: "80000000",
		IsActive:           &active,
	}); err != nil {
		return 0, err
	}
	return 1, nil
}
