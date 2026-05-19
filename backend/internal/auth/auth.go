package auth

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	minPasswordLength = 8
	maxEmailLength    = 320
)

var (
	ErrInvalidInput       = errors.New("invalid auth input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("authentication required")
	ErrNotFound           = errors.New("auth resource not found")
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthenticatedSession struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

func NormalizeRegister(input RegisterInput) (RegisterInput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return RegisterInput{}, fieldError("email", "is required")
	}
	if len(email) > maxEmailLength || !strings.Contains(email, "@") {
		return RegisterInput{}, fieldError("email", "must be a valid email address")
	}
	if len(input.Password) < minPasswordLength {
		return RegisterInput{}, fieldError("password", "must be at least 8 characters")
	}
	return RegisterInput{Email: email, Password: input.Password}, nil
}

func NormalizeLogin(input LoginInput) (LoginInput, error) {
	normalized, err := NormalizeRegister(RegisterInput(input))
	if err != nil {
		return LoginInput{}, err
	}
	return LoginInput(normalized), nil
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

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok && userID != ""
}

func RequireUserID(ctx context.Context) (string, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return "", ErrUnauthenticated
	}
	return userID, nil
}
