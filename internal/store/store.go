package store

import (
"context"
"crypto/rand"
"database/sql"
"encoding/hex"
"errors"
"fmt"
"strings"

_ "modernc.org/sqlite"
)

var (
ErrInvalidInput  = errors.New("invalid input")
ErrNotFound      = errors.New("not found")
ErrUnauthorized  = errors.New("unauthorized")
ErrEmailConflict = errors.New("email already exists")
)

type Store struct {
db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
db, err := sql.Open("sqlite", path)
if err != nil {
return nil, fmt.Errorf("open database: %w", err)
}
db.SetMaxOpenConns(1)
if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;"); err != nil {
_ = db.Close()
return nil, fmt.Errorf("configure database: %w", err)
}
s := &Store{db: db}
if err := s.migrate(ctx); err != nil {
_ = db.Close()
return nil, fmt.Errorf("migrate database: %w", err)
}
return s, nil
}

func (s *Store) Close() error {
return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS users (
id TEXT PRIMARY KEY,
email TEXT NOT NULL UNIQUE,
name TEXT NOT NULL DEFAULT '',
password_hash TEXT NOT NULL,
created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
token TEXT PRIMARY KEY,
user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
expires_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS password_resets (
token TEXT PRIMARY KEY,
user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
expires_at DATETIME NOT NULL,
used_at DATETIME
);
CREATE TABLE IF NOT EXISTS projects (
id TEXT PRIMARY KEY,
title TEXT NOT NULL,
description TEXT NOT NULL DEFAULT '',
color TEXT NOT NULL DEFAULT '#4f46e5',
created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS project_favorites (
user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
PRIMARY KEY (user_id, project_id)
);
CREATE TABLE IF NOT EXISTS todos (
id TEXT PRIMARY KEY,
project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
title TEXT NOT NULL,
description TEXT NOT NULL DEFAULT '',
due_date DATETIME,
priority TEXT NOT NULL DEFAULT 'normal',
completed INTEGER NOT NULL DEFAULT 0,
position INTEGER NOT NULL,
created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_todos_project_position ON todos(project_id, position);
`)
if err != nil {
return fmt.Errorf("create schema: %w", err)
}
return nil
}

func NewID() (string, error) {
var b [16]byte
if _, err := rand.Read(b[:]); err != nil {
return "", fmt.Errorf("random id: %w", err)
}
return hex.EncodeToString(b[:]), nil
}

func NormalizeEmail(email string) string {
return strings.ToLower(strings.TrimSpace(email))
}

func boolToInt(value bool) int {
if value {
return 1
}
return 0
}

func requireAffected(result sql.Result) error {
affected, err := result.RowsAffected()
if err != nil {
return fmt.Errorf("rows affected: %w", err)
}
if affected == 0 {
return ErrNotFound
}
return nil
}
