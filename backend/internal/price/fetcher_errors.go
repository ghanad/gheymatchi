package price

import "errors"

var (
	ErrUnsupportedSourceURL = errors.New("unsupported source url")
	ErrProductUnavailable   = errors.New("product unavailable")
	ErrPriceNotFound        = errors.New("price not found")
)
