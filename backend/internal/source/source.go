package source

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

const (
	maxURLLength        = 2048
	maxSourceNameLength = 80
)

var (
	ErrNotFound = errors.New("product source not found")
)

type ProductSource struct {
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	URL        string    `json:"url"`
	SourceName string    `json:"source_name"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateInput struct {
	URL        string
	SourceName string
	IsActive   *bool
}

type UpdateInput struct {
	URL        *string
	SourceName *string
	IsActive   *bool
}

func NormalizeCreate(input CreateInput) (CreateInput, error) {
	normalizedURL, err := normalizeURL(input.URL)
	if err != nil {
		return CreateInput{}, err
	}

	sourceName, err := normalizeSourceName(input.SourceName, normalizedURL)
	if err != nil {
		return CreateInput{}, err
	}

	return CreateInput{
		URL:        normalizedURL,
		SourceName: sourceName,
		IsActive:   input.IsActive,
	}, nil
}

func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	if input.URL != nil {
		normalizedURL, err := normalizeURL(*input.URL)
		if err != nil {
			return UpdateInput{}, err
		}
		if err := validateSupportedSourceURL("digikala", normalizedURL); err != nil {
			return UpdateInput{}, err
		}
		input.URL = &normalizedURL
	}

	if input.SourceName != nil {
		sourceName, err := normalizeSupportedSourceName(*input.SourceName)
		if err != nil {
			return UpdateInput{}, err
		}
		input.SourceName = &sourceName
	}

	if input.URL == nil && input.SourceName == nil && input.IsActive == nil {
		return UpdateInput{}, fieldError("body", "must include at least one supported field")
	}

	return input, nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + " " + e.Message
}

func normalizeURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fieldError("url", "is required")
	}
	if len(trimmed) > maxURLLength {
		return "", fieldError("url", "must be 2048 characters or less")
	}

	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fieldError("url", "must be a valid absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fieldError("url", "must use http or https")
	}

	return parsed.String(), nil
}

func normalizeSourceName(value string, normalizedURL string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		normalized = sourceNameFromURL(normalizedURL)
	}
	sourceName, err := normalizeSupportedSourceName(normalized)
	if err != nil {
		return "", err
	}
	if err := validateSupportedSourceURL(sourceName, normalizedURL); err != nil {
		return "", err
	}
	return sourceName, nil
}

func normalizeSupportedSourceName(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if len(normalized) > maxSourceNameLength {
		return "", fieldError("source_name", "must be 80 characters or less")
	}
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return "", fieldError("source_name", "must contain only letters, numbers, hyphens, or underscores")
	}
	if normalized != "digikala" {
		return "", fieldError("source_name", "must be one of the supported sources")
	}
	return normalized, nil
}

func sourceNameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	switch strings.ToLower(parsed.Hostname()) {
	case "digikala.com", "www.digikala.com":
		return "digikala"
	default:
		return ""
	}
}

func validateSupportedSourceURL(sourceName string, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fieldError("url", "must be a valid absolute URL")
	}

	switch sourceName {
	case "digikala":
		host := strings.ToLower(parsed.Hostname())
		if host != "digikala.com" && host != "www.digikala.com" {
			return fieldError("url", "must be a Digikala product URL")
		}
		if !hasDigikalaProductID(parsed.Path) {
			return fieldError("url", "must include a Digikala dkp product id")
		}
		return nil
	default:
		return fieldError("source_name", "must be one of the supported sources")
	}
}

func hasDigikalaProductID(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		lower := strings.ToLower(part)
		if !strings.HasPrefix(lower, "dkp-") {
			continue
		}
		id := strings.TrimPrefix(lower, "dkp-")
		if id == "" {
			return false
		}
		for _, r := range id {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	}
	return false
}

func fieldError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}
