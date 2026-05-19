package auth

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"gheymatchi/backend/internal/db"
)

func TestSQLiteStoreRegisterLoginAuthenticateLogout(t *testing.T) {
	store, database := newTestStore(t)
	ctx := context.Background()

	registered, err := store.Register(ctx, RegisterInput{Email: "USER@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registered.User.Email != "user@example.com" {
		t.Fatalf("registered email = %q, want normalized lowercase", registered.User.Email)
	}
	if registered.Token == "" {
		t.Fatal("registered token is empty")
	}

	var passwordHash string
	if err := database.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = ?`, registered.User.ID).Scan(&passwordHash); err != nil {
		t.Fatalf("load password hash: %v", err)
	}
	if passwordHash == "password123" {
		t.Fatal("stored password in plain text")
	}

	authenticated, err := store.AuthenticateToken(ctx, registered.Token)
	if err != nil {
		t.Fatalf("AuthenticateToken() error = %v", err)
	}
	if authenticated.ID != registered.User.ID {
		t.Fatalf("authenticated user ID = %q, want %q", authenticated.ID, registered.User.ID)
	}

	loggedIn, err := store.Login(ctx, LoginInput{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loggedIn.Token == "" || loggedIn.Token == registered.Token {
		t.Fatal("Login() did not issue a distinct session token")
	}

	if _, err := store.Login(ctx, LoginInput{Email: "user@example.com", Password: "wrongpass"}); err != ErrInvalidCredentials {
		t.Fatalf("Login(wrong password) error = %v, want ErrInvalidCredentials", err)
	}

	if err := store.Logout(ctx, registered.Token); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := store.AuthenticateToken(ctx, registered.Token); err != ErrUnauthenticated {
		t.Fatalf("AuthenticateToken(after logout) error = %v, want ErrUnauthenticated", err)
	}
}

func newTestStore(t *testing.T) (*SQLiteStore, *sql.DB) {
	t.Helper()

	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	migrations, err := db.LoadMigrations(os.DirFS("../../migrations"))
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}
	if err := db.ApplyMigrations(ctx, database, migrations); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	return NewSQLiteStore(database), database
}
