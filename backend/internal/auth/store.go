package auth

import "context"

type Store interface {
	Register(ctx context.Context, input RegisterInput) (AuthenticatedSession, error)
	Login(ctx context.Context, input LoginInput) (AuthenticatedSession, error)
	AuthenticateToken(ctx context.Context, token string) (User, error)
	Logout(ctx context.Context, token string) error
}
