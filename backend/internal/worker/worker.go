package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gheymatchi/backend/internal/crawl"
	"gheymatchi/backend/internal/price"
	"gheymatchi/backend/internal/source"
)

type SourceLister interface {
	ListActive(ctx context.Context) ([]source.ProductSource, error)
}

type PriceStore interface {
	Create(ctx context.Context, productID string, productSourceID string, input price.CreateInput) (price.PricePoint, error)
}

type AlertEvaluator interface {
	EvaluatePricePoint(ctx context.Context, pricePoint price.PricePoint) error
}

type NotificationProcessor interface {
	ProcessPending(ctx context.Context) error
}

type CrawlStore interface {
	Start(ctx context.Context, sourceID string) (crawl.Run, error)
	Finish(ctx context.Context, id string, status string, errorMessage *string) error
}

type Runner struct {
	sources  SourceLister
	prices   PriceStore
	crawls   CrawlStore
	alerts   AlertEvaluator
	notifier NotificationProcessor
	fetcher  price.Fetcher
	logger   *slog.Logger
}

func NewRunner(sources SourceLister, prices PriceStore, crawls CrawlStore, alerts AlertEvaluator, notifier NotificationProcessor, fetcher price.Fetcher, logger *slog.Logger) Runner {
	return Runner{
		sources:  sources,
		prices:   prices,
		crawls:   crawls,
		alerts:   alerts,
		notifier: notifier,
		fetcher:  fetcher,
		logger:   logger,
	}
}

func (r Runner) RunOnce(ctx context.Context) error {
	if r.notifier != nil {
		if err := r.notifier.ProcessPending(ctx); err != nil {
			r.logger.Error("notification processing failed", slog.String("error", err.Error()))
		}
	}

	sources, err := r.sources.ListActive(ctx)
	if err != nil {
		return err
	}

	r.logger.Info("loaded active sources", slog.Int("count", len(sources)))
	for _, productSource := range sources {
		if err := r.checkSource(ctx, productSource); err != nil {
			r.logSourceCheckFailure(ctx, productSource, err)
		}
	}

	if r.notifier != nil {
		if err := r.notifier.ProcessPending(ctx); err != nil {
			r.logger.Error("notification processing failed", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (r Runner) logSourceCheckFailure(ctx context.Context, productSource source.ProductSource, err error) {
	attrs := []slog.Attr{
		slog.String("source_id", productSource.ID),
		slog.String("product_id", productSource.ProductID),
		slog.String("error", err.Error()),
	}

	if errors.Is(err, price.ErrSourceAccessDenied) {
		r.logger.LogAttrs(ctx, slog.LevelWarn, "source access denied", attrs...)
		return
	}

	r.logger.LogAttrs(ctx, slog.LevelError, "source check failed", attrs...)
}

func (r Runner) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	if err := r.RunOnce(ctx); err != nil {
		r.logger.Error("worker tick failed", slog.String("error", err.Error()))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				r.logger.Error("worker tick failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (r Runner) checkSource(ctx context.Context, productSource source.ProductSource) error {
	r.logger.Info(
		"checking source",
		slog.String("source_id", productSource.ID),
		slog.String("product_id", productSource.ProductID),
		slog.String("url", productSource.URL),
	)

	run, err := r.crawls.Start(ctx, productSource.ID)
	if err != nil {
		return err
	}

	result, err := r.fetcher.Fetch(ctx, productSource)
	if err != nil {
		message := err.Error()
		_ = r.crawls.Finish(ctx, run.ID, crawl.StatusFailed, &message)
		return err
	}

	pricePoint, err := r.prices.Create(ctx, productSource.ProductID, productSource.ID, price.CreateInput{
		PriceIRR:   result.PriceIRR,
		CapturedAt: result.CapturedAt,
		RawPayload: result.RawPayload,
	})
	if err != nil {
		message := err.Error()
		_ = r.crawls.Finish(ctx, run.ID, crawl.StatusFailed, &message)
		return err
	}

	if r.alerts != nil {
		if err := r.alerts.EvaluatePricePoint(ctx, pricePoint); err != nil {
			message := err.Error()
			_ = r.crawls.Finish(ctx, run.ID, crawl.StatusFailed, &message)
			return err
		}
	}

	if err := r.crawls.Finish(ctx, run.ID, crawl.StatusSucceeded, nil); err != nil {
		return err
	}

	r.logger.Info(
		"source check succeeded",
		slog.String("source_id", productSource.ID),
		slog.String("product_id", productSource.ProductID),
		slog.Int64("price_irr", result.PriceIRR),
	)
	return nil
}
