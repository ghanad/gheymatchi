package source

import "context"

type Store interface {
	Create(ctx context.Context, productID string, input CreateInput) (ProductSource, error)
	List(ctx context.Context, productID string) ([]ProductSource, error)
	ListActive(ctx context.Context) ([]ProductSource, error)
	Update(ctx context.Context, productID string, sourceID string, input UpdateInput) (ProductSource, error)
	Delete(ctx context.Context, productID string, sourceID string) error
}
