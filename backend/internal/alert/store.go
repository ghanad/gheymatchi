package alert

import "context"

type Store interface {
	Create(ctx context.Context, productID string, input CreateInput) (Alert, error)
	List(ctx context.Context, productID string) ([]Alert, error)
	Update(ctx context.Context, productID string, alertID string, input UpdateInput) (Alert, error)
	Delete(ctx context.Context, productID string, alertID string) error
}
