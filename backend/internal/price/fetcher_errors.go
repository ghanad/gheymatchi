package price

import "errors"

var (
	ErrUnsupportedSourceURL = errors.New("unsupported source url")
	ErrSourceAccessDenied   = errors.New("source access denied")
	ErrProductUnavailable   = errors.New("product unavailable")
	ErrPriceNotFound        = errors.New("price not found")
)
