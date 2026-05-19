package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Register(ctx context.Context, input RegisterInput) (AuthenticatedSession, error) {
	normalized, err := NormalizeRegister(input)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	passwordHash, err := HashPassword(normalized.Password)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	now := time.Now().UTC()
	user := User{ID: newID(), Email: normalized.Email, CreatedAt: now, UpdatedAt: now}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("begin register: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
INSERT INTO users (id, email, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)`,
		user.ID,
		user.Email,
		passwordHash,
		formatTime(user.CreatedAt),
		formatTime(user.UpdatedAt),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return AuthenticatedSession{}, fieldError("email", "is already registered")
		}
		return AuthenticatedSession{}, fmt.Errorf("create user: %w", err)
	}

	session, err := createSession(ctx, tx, user)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	if err := tx.Commit(); err != nil {
		return AuthenticatedSession{}, fmt.Errorf("commit register: %w", err)
	}
	committed = true
	return session, nil
}

func (s *SQLiteStore) Login(ctx context.Context, input LoginInput) (AuthenticatedSession, error) {
	normalized, err := NormalizeLogin(input)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	row := s.db.QueryRowContext(ctx, `
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = ?`, normalized.Email)

	var user User
	var passwordHash string
	var createdAt string
	var updatedAt string
	if err := row.Scan(&user.ID, &user.Email, &passwordHash, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return AuthenticatedSession{}, ErrInvalidCredentials
		}
		return AuthenticatedSession{}, fmt.Errorf("load user for login: %w", err)
	}
	ok, err := VerifyPassword(normalized.Password, passwordHash)
	if err != nil || !ok {
		return AuthenticatedSession{}, ErrInvalidCredentials
	}
	user.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	user.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	session, err := createSession(ctx, s.db, user)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	return session, nil
}

func (s *SQLiteStore) AuthenticateToken(ctx context.Context, token string) (User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, ErrUnauthenticated
	}
	hash := hashToken(token)
	row := s.db.QueryRowContext(ctx, `
SELECT users.id, users.email, users.created_at, users.updated_at
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.token_hash = ?`, hash)

	var user User
	var createdAt string
	var updatedAt string
	if err := row.Scan(&user.ID, &user.Email, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrUnauthenticated
		}
		return User{}, fmt.Errorf("authenticate token: %w", err)
	}
	var err error
	user.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return User{}, err
	}
	user.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *SQLiteStore) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, hashToken(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

type sessionExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func createSession(ctx context.Context, execer sessionExecer, user User) (AuthenticatedSession, error) {
	token, err := newToken()
	if err != nil {
		return AuthenticatedSession{}, err
	}
	now := time.Now().UTC()
	_, err = execer.ExecContext(ctx, `
INSERT INTO sessions (id, user_id, token_hash, created_at)
VALUES (?, ?, ?, ?)`,
		newID(),
		user.ID,
		hashToken(token),
		formatTime(now),
	)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("create session: %w", err)
	}
	return AuthenticatedSession{User: user, Token: token}, nil
}

func newToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(bytes[:]), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse auth time: %w", err)
	}
	return parsed, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate auth id: %v", err))
	}
	return hex.EncodeToString(bytes[:])
}
