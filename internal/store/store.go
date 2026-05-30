package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrationsFS embed.FS

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
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// migrate runs all pending database migrations using golang-migrate.
// It also handles the transition from the legacy hand-rolled migration system:
// if the old schema_migrations table (single "version" column, no "dirty" column)
// is detected, its version is read, the table is dropped, and golang-migrate is
// told to fast-forward to that version so it does not re-run already-applied migrations.
func (s *Store) migrate() error {
	legacyVersion, hasLegacy, err := s.detectLegacyMigrations()
	if err != nil {
		return fmt.Errorf("detect legacy migrations: %w", err)
	}
	if hasLegacy {
		if _, err := s.db.Exec(`DROP TABLE IF EXISTS schema_migrations`); err != nil {
			return fmt.Errorf("drop legacy migration table: %w", err)
		}
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(s.db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	// Close only the source (embedded FS — no-op). Do NOT call m.Close() because
	// the sqlite driver's Close() would close our shared *sql.DB connection.
	defer src.Close()

	if hasLegacy {
		if err := m.Force(legacyVersion); err != nil {
			return fmt.Errorf("force migration version %d: %w", legacyVersion, err)
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

// detectLegacyMigrations checks whether the database was previously managed by the
// hand-rolled migration system (schema_migrations table with a single "version" column
// and no "dirty" column). Returns (version, true, nil) when legacy state is detected.
func (s *Store) detectLegacyMigrations() (version int, exists bool, err error) {
	var versionColCount int
	row := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('schema_migrations') WHERE name='version'`)
	if err := row.Scan(&versionColCount); err != nil || versionColCount == 0 {
		return 0, false, err
	}
	// If the table also has a "dirty" column it is already the golang-migrate format.
	var dirtyColCount int
	row = s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('schema_migrations') WHERE name='dirty'`)
	if err := row.Scan(&dirtyColCount); err != nil {
		return 0, false, err
	}
	if dirtyColCount > 0 {
		return 0, false, nil
	}
	row = s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`)
	if err := row.Scan(&version); err != nil {
		return 0, false, fmt.Errorf("read legacy schema version: %w", err)
	}
	return version, true, nil
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
