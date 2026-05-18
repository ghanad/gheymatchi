package marketrate

import "context"

type Provider interface {
	Fetch(ctx context.Context) ([]CreateInput, error)
}

type MockProvider struct {
	Rates []CreateInput
}

func (p MockProvider) Fetch(ctx context.Context) ([]CreateInput, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	rates := make([]CreateInput, len(p.Rates))
	copy(rates, p.Rates)
	return rates, nil
}
