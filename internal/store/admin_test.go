package store

import (
	"context"
	"database/sql"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/bhzunami/machino/internal/model"

	_ "modernc.org/sqlite"

	"golang.org/x/crypto/bcrypt"
)

func TestAdminUserManagement(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:", slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	admin, err := s.CreateUser(ctx, "admin@example.com", "Admin", string(hash), model.RoleUser)
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	regular, err := s.CreateUser(ctx, "user@example.com", "User", string(hash), model.RoleUser)
	if err != nil {
		t.Fatalf("create regular user: %v", err)
	}
	if admin.Role != model.RoleUser || regular.Role != model.RoleUser {
		t.Fatalf("new users should default to user role: %#v %#v", admin, regular)
	}

	admin, err = s.SetAdminByEmail(ctx, "ADMIN@example.com")
	if err != nil {
		t.Fatalf("set admin: %v", err)
	}
	if admin.Role != model.RoleAdmin {
		t.Fatalf("expected admin role, got %q", admin.Role)
	}

	updated, err := s.UpdateUser(ctx, regular.ID, "renamed@example.com", "Renamed", false, model.RoleAdmin)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.Email != "renamed@example.com" || updated.Name != "Renamed" || updated.Searchable || updated.Role != model.RoleAdmin {
		t.Fatalf("unexpected updated user: %#v", updated)
	}

	users, err := s.ListUsers(ctx)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if err := s.DeleteUser(ctx, regular.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	users, err = s.ListUsers(ctx)
	if err != nil {
		t.Fatalf("list users after delete: %v", err)
	}
	if len(users) != 1 || users[0].ID != admin.ID {
		t.Fatalf("unexpected users after delete: %#v", users)
	}
}

// TestMigrationAddsRoleWithoutLosingUsers verifies that opening a database that already
// has the users table (but no role column and no migration tracking table) correctly
// runs the pending migrations and preserves existing user data.
func TestMigrationAddsRoleWithoutLosingUsers(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open old db: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
CREATE TABLE users (
id TEXT PRIMARY KEY,
email TEXT NOT NULL UNIQUE,
name TEXT NOT NULL DEFAULT '',
password_hash TEXT NOT NULL,
created_at DATETIME NOT NULL,
searchable INTEGER NOT NULL DEFAULT 1
);
`); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at, searchable) VALUES (?, ?, ?, ?, ?, ?)`,
		"u1", "old@example.com", "Old User", "hash", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 1,
	); err != nil {
		t.Fatalf("seed old db: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close old db: %v", err)
	}

	s, err := Open(ctx, dbPath, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open migrated store: %v", err)
	}
	defer s.Close()

	u, _, err := s.UserByEmail(ctx, "old@example.com")
	if err != nil {
		t.Fatalf("load migrated user: %v", err)
	}
	if u.ID != "u1" || u.Role != model.RoleUser || !u.Searchable {
		t.Fatalf("unexpected migrated user: %#v", u)
	}
}
