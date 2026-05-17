package product

import "context"

type Store interface {
	Create(ctx context.Context, input CreateInput) (Product, error)
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id string) (Product, error)
	Update(ctx context.Context, id string, input UpdateInput) (Product, error)
	Delete(ctx context.Context, id string) error
}
