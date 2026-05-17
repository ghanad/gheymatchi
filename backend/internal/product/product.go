package product

import (
	"errors"
	"strings"
	"time"
)

const (
	maxNameLength        = 200
	maxDescriptionLength = 2000
)

var (
	ErrInvalidInput = errors.New("invalid product input")
	ErrNotFound     = errors.New("product not found")
)

type Product struct {
	ID          string    `json:"id"`
	UserID      *string   `json:"user_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name        string
	Description string
}

type UpdateInput struct {
	Name        *string
	Description *string
}

func NormalizeCreate(input CreateInput) (CreateInput, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if name == "" {
		return CreateInput{}, fieldError("name", "is required")
	}
	if len(name) > maxNameLength {
		return CreateInput{}, fieldError("name", "must be 200 characters or less")
	}
	if len(description) > maxDescriptionLength {
		return CreateInput{}, fieldError("description", "must be 2000 characters or less")
	}

	return CreateInput{Name: name, Description: description}, nil
}

func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return UpdateInput{}, fieldError("name", "is required")
		}
		if len(name) > maxNameLength {
			return UpdateInput{}, fieldError("name", "must be 200 characters or less")
		}
		input.Name = &name
	}

	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		if len(description) > maxDescriptionLength {
			return UpdateInput{}, fieldError("description", "must be 2000 characters or less")
		}
		input.Description = &description
	}

	if input.Name == nil && input.Description == nil {
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

func fieldError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}
