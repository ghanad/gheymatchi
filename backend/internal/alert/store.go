package alert

import (
	"context"
	"time"
)

type Store interface {
	Create(ctx context.Context, productID string, input CreateInput) (Alert, error)
	List(ctx context.Context, productID string) ([]Alert, error)
	Update(ctx context.Context, productID string, alertID string, input UpdateInput) (Alert, error)
	Delete(ctx context.Context, productID string, alertID string) error
}

type EvaluationStore interface {
	ListActiveByProduct(ctx context.Context, productID string) ([]Alert, error)
	MarkTriggered(ctx context.Context, alertID string, triggeredAt time.Time) error
}
