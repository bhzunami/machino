package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"machino/internal/model"
)

func (s *Store) CreateUser(ctx context.Context, email, name, passwordHash string) (model.User, error) {
	email = NormalizeEmail(email)
	if email == "" || passwordHash == "" {
		return model.User{}, ErrInvalidInput
	}
	id, err := NewID()
	if err != nil {
		return model.User{}, err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash, created_at, searchable) VALUES (?, ?, ?, ?, ?, 1)`,
		id, email, strings.TrimSpace(name), passwordHash, now)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return model.User{}, ErrEmailConflict
		}
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}
	return model.User{ID: id, Email: email, Name: strings.TrimSpace(name), Searchable: true, CreatedAt: now}, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (model.User, string, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, created_at, searchable FROM users WHERE email = ?`,
		NormalizeEmail(email))
	var u model.User
	var passwordHash string
	var searchable int
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &passwordHash, &u.CreatedAt, &searchable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, "", ErrUnauthorized
		}
		return model.User{}, "", fmt.Errorf("select user by email: %w", err)
	}
	u.Searchable = searchable == 1
	return u, passwordHash, nil
}

func (s *Store) UserBySession(ctx context.Context, token string) (model.User, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT u.id, u.email, u.name, u.created_at, u.searchable
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token = ? AND s.expires_at > ?`, token, time.Now().UTC())
	var u model.User
	var searchable int
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &searchable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUnauthorized
		}
		return model.User{}, fmt.Errorf("select session user: %w", err)
	}
	u.Searchable = searchable == 1
	return u, nil
}

func (s *Store) CreateSession(ctx context.Context, userID string, ttl time.Duration) (string, time.Time, error) {
	token, err := NewID()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(ttl)
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("insert session: %w", err)
	}
	return token, expiresAt, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Store) UpdateProfile(ctx context.Context, userID, email, name string, searchable bool) (model.User, error) {
	email = NormalizeEmail(email)
	if email == "" {
		return model.User{}, ErrInvalidInput
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET email = ?, name = ?, searchable = ? WHERE id = ?`,
		email, strings.TrimSpace(name), boolToInt(searchable), userID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return model.User{}, ErrEmailConflict
		}
		return model.User{}, fmt.Errorf("update profile: %w", err)
	}
	return s.userByID(ctx, userID)
}

func (s *Store) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return requireAffected(result)
}

func (s *Store) CreatePasswordReset(ctx context.Context, email string, ttl time.Duration) (string, error) {
	u, _, err := s.UserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	token, err := NewID()
	if err != nil {
		return "", err
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO password_resets (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, u.ID, time.Now().UTC().Add(ttl)); err != nil {
		return "", fmt.Errorf("insert reset token: %w", err)
	}
	return token, nil
}

func (s *Store) UsePasswordReset(ctx context.Context, token, passwordHash string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var userID string
	row := tx.QueryRowContext(ctx,
		`SELECT user_id FROM password_resets WHERE token = ? AND used_at IS NULL AND expires_at > ?`,
		token, time.Now().UTC())
	if err = row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUnauthorized
		}
		return fmt.Errorf("select reset token: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID); err != nil {
		return fmt.Errorf("update reset password: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE password_resets SET used_at = ? WHERE token = ?`,
		time.Now().UTC(), token); err != nil {
		return fmt.Errorf("consume reset token: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit reset tx: %w", err)
	}
	return nil
}

func (s *Store) userByID(ctx context.Context, userID string) (model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, created_at, searchable FROM users WHERE id = ?`, userID)
	var u model.User
	var searchable int
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &searchable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("select user: %w", err)
	}
	u.Searchable = searchable == 1
	return u, nil
}

// SearchUsers returns users whose name or email starts with q (case-insensitive),
// limited to users who have opted in (searchable = 1), excluding the caller.
// Only name and id are returned — never email — to protect privacy.
func (s *Store) SearchUsers(ctx context.Context, q, excludeUserID string) ([]model.UserSearchResult, error) {
	q = strings.TrimSpace(q)
	if len([]rune(q)) < 3 {
		return nil, ErrInvalidInput
	}
	pattern := q + "%"
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name FROM users
WHERE searchable = 1
  AND id != ?
  AND (name LIKE ? OR email LIKE ?)
ORDER BY name
LIMIT 10`,
		excludeUserID, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	var results []model.UserSearchResult
	for rows.Next() {
		var r model.UserSearchResult
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, fmt.Errorf("scan user search: %w", err)
		}
		results = append(results, r)
	}
	if results == nil {
		results = []model.UserSearchResult{}
	}
	return results, rows.Err()
}
