package price

import "context"

type Store interface {
	Create(ctx context.Context, productID string, productSourceID string, input CreateInput) (PricePoint, error)
	ListByProduct(ctx context.Context, productID string) ([]PricePoint, error)
}
