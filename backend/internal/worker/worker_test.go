package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"gheymatchi/backend/internal/crawl"
	"gheymatchi/backend/internal/price"
	"gheymatchi/backend/internal/source"
)

func TestRunnerRunOnceCreatesPricePointAndRecordsSuccess(t *testing.T) {
	crawls := &fakeCrawlStore{}
	prices := &fakePriceStore{}
	runner := NewRunner(
		fakeSourceLister{sources: []source.ProductSource{testSource()}},
		prices,
		crawls,
		&fakeAlertEvaluator{},
		&fakeNotificationProcessor{},
		fakeFetcher{result: price.FetchResult{PriceIRR: 123_000, CapturedAt: time.Now().UTC()}},
		testLogger(),
	)

	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if len(prices.created) != 1 {
		t.Fatalf("created price points = %d, want 1", len(prices.created))
	}
	if prices.created[0].PriceIRR != 123_000 {
		t.Fatalf("created PriceIRR = %d, want 123000", prices.created[0].PriceIRR)
	}
	if crawls.finishedStatus != crawl.StatusSucceeded {
		t.Fatalf("finished status = %q, want %q", crawls.finishedStatus, crawl.StatusSucceeded)
	}
	if crawls.errorMessage != nil {
		t.Fatalf("errorMessage = %q, want nil", *crawls.errorMessage)
	}
}

func TestRunnerRunOnceRecordsFailedFetch(t *testing.T) {
	crawls := &fakeCrawlStore{}
	prices := &fakePriceStore{}
	runner := NewRunner(
		fakeSourceLister{sources: []source.ProductSource{testSource()}},
		prices,
		crawls,
		&fakeAlertEvaluator{},
		&fakeNotificationProcessor{},
		fakeFetcher{err: errors.New("fetch failed")},
		testLogger(),
	)

	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if len(prices.created) != 0 {
		t.Fatalf("created price points = %d, want 0", len(prices.created))
	}
	if crawls.finishedStatus != crawl.StatusFailed {
		t.Fatalf("finished status = %q, want %q", crawls.finishedStatus, crawl.StatusFailed)
	}
	if crawls.errorMessage == nil || *crawls.errorMessage != "fetch failed" {
		t.Fatalf("errorMessage = %v, want fetch failed", crawls.errorMessage)
	}
}

type fakeSourceLister struct {
	sources []source.ProductSource
	err     error
}

func (f fakeSourceLister) ListActive(ctx context.Context) ([]source.ProductSource, error) {
	return f.sources, f.err
}

type fakePriceStore struct {
	created []price.CreateInput
	err     error
}

func (f *fakePriceStore) Create(ctx context.Context, productID string, productSourceID string, input price.CreateInput) (price.PricePoint, error) {
	if f.err != nil {
		return price.PricePoint{}, f.err
	}
	f.created = append(f.created, input)
	return price.PricePoint{ID: "price-1", ProductID: productID, ProductSourceID: productSourceID, PriceIRR: input.PriceIRR}, nil
}

type fakeCrawlStore struct {
	finishedStatus string
	errorMessage   *string
}

func (f *fakeCrawlStore) Start(ctx context.Context, sourceID string) (crawl.Run, error) {
	return crawl.Run{ID: "crawl-1", SourceID: sourceID, Status: crawl.StatusRunning}, nil
}

func (f *fakeCrawlStore) Finish(ctx context.Context, id string, status string, errorMessage *string) error {
	f.finishedStatus = status
	f.errorMessage = errorMessage
	return nil
}

type fakeFetcher struct {
	result price.FetchResult
	err    error
}

func (f fakeFetcher) Fetch(ctx context.Context, productSource source.ProductSource) (price.FetchResult, error) {
	return f.result, f.err
}

type fakeAlertEvaluator struct {
	evaluated []price.PricePoint
	err       error
}

func (f *fakeAlertEvaluator) EvaluatePricePoint(ctx context.Context, pricePoint price.PricePoint) error {
	if f.err != nil {
		return f.err
	}
	f.evaluated = append(f.evaluated, pricePoint)
	return nil
}

type fakeNotificationProcessor struct {
	processed int
	err       error
}

func (f *fakeNotificationProcessor) ProcessPending(ctx context.Context) error {
	f.processed++
	return f.err
}

func testSource() source.ProductSource {
	return source.ProductSource{
		ID:        "source-1",
		ProductID: "product-1",
		URL:       "https://example.com/products/1",
		IsActive:  true,
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
