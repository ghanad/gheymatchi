package marketrate

import "context"

type Store interface {
	Create(ctx context.Context, input CreateInput) (MarketRate, error)
	Latest(ctx context.Context, rateType *string) ([]MarketRate, error)
	History(ctx context.Context, rateType *string) ([]MarketRate, error)
}
