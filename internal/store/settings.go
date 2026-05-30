package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bhzunami/machino/internal/model"
)

func (s *Store) BootstrapAppSettings(ctx context.Context, defaults model.AppSettings) error {
	defaults = normalizeAppSettings(defaults)
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO app_settings (
	id, app_domain, registration_enabled, smtp_host, smtp_port,
	smtp_username, smtp_password, smtp_from, created_at, updated_at
) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		defaults.AppDomain,
		boolToInt(defaults.RegistrationEnabled),
		defaults.SMTPHost,
		defaults.SMTPPort,
		defaults.SMTPUsername,
		defaults.SMTPPassword,
		defaults.SMTPFrom,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("bootstrap app settings: %w", err)
	}
	return nil
}

func (s *Store) AppSettings(ctx context.Context) (model.AppSettings, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT app_domain, registration_enabled, smtp_host, smtp_port, smtp_username, smtp_password, smtp_from, created_at, updated_at
FROM app_settings
WHERE id = 1`)
	settings, err := scanAppSettings(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AppSettings{}, ErrNotFound
		}
		return model.AppSettings{}, fmt.Errorf("select app settings: %w", err)
	}
	return settings, nil
}

// execerContext is satisfied by *sql.DB and *sql.Tx, allowing shared DML helpers.
type execerContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (s *Store) UpdateAppSettings(ctx context.Context, next model.AppSettings) (model.AppSettings, error) {
	next = normalizeAppSettings(next)
	if err := validateAppSettings(next); err != nil {
		return model.AppSettings{}, err
	}
	result, err := execUpdateAppSettings(ctx, s.db, next)
	if err != nil {
		return model.AppSettings{}, fmt.Errorf("update app settings: %w", err)
	}
	if err := requireAffected(result); err != nil {
		return model.AppSettings{}, err
	}
	return s.AppSettings(ctx)
}

func (s *Store) AdminCount(ctx context.Context) (int, error) {
	row := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = ?`, model.RoleAdmin)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return count, nil
}

func (s *Store) SetupRequired(ctx context.Context) (bool, error) {
	count, err := s.AdminCount(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *Store) CreateSetupAdmin(ctx context.Context, email, name, passwordHash string, settings model.AppSettings) (model.User, error) {
	email = NormalizeEmail(email)
	name = strings.TrimSpace(name)
	settings = normalizeAppSettings(settings)
	if email == "" || name == "" || passwordHash == "" {
		return model.User{}, ErrInvalidInput
	}
	if err := validateAppSettings(settings); err != nil {
		return model.User{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.User{}, fmt.Errorf("begin setup tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var adminCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = ?`, model.RoleAdmin).Scan(&adminCount); err != nil {
		return model.User{}, fmt.Errorf("count admins in setup: %w", err)
	}
	if adminCount > 0 {
		return model.User{}, ErrUnauthorized
	}

	id, err := NewID()
	if err != nil {
		return model.User{}, err
	}
	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
INSERT INTO users (id, email, name, password_hash, role, created_at, searchable)
VALUES (?, ?, ?, ?, ?, ?, 1)`,
		id, email, name, passwordHash, model.RoleAdmin, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return model.User{}, ErrEmailConflict
		}
		return model.User{}, fmt.Errorf("insert setup admin: %w", err)
	}
	if err := updateAppSettingsTx(ctx, tx, settings); err != nil {
		return model.User{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.User{}, fmt.Errorf("commit setup tx: %w", err)
	}
	committed = true

	return model.User{
		ID:         id,
		Email:      email,
		Name:       name,
		Role:       model.RoleAdmin,
		Searchable: true,
		CreatedAt:  now,
	}, nil
}

type appSettingsScanner interface {
	Scan(dest ...any) error
}

func scanAppSettings(row appSettingsScanner) (model.AppSettings, error) {
	var settings model.AppSettings
	var registrationEnabled int
	if err := row.Scan(
		&settings.AppDomain,
		&registrationEnabled,
		&settings.SMTPHost,
		&settings.SMTPPort,
		&settings.SMTPUsername,
		&settings.SMTPPassword,
		&settings.SMTPFrom,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	); err != nil {
		return model.AppSettings{}, err
	}
	settings.RegistrationEnabled = registrationEnabled == 1
	settings.SMTPPasswordSet = settings.SMTPPassword != ""
	return settings, nil
}

func updateAppSettingsTx(ctx context.Context, tx *sql.Tx, next model.AppSettings) error {
	_, err := execUpdateAppSettings(ctx, tx, next)
	if err != nil {
		return fmt.Errorf("update setup app settings: %w", err)
	}
	return nil
}

// execUpdateAppSettings runs the shared UPDATE for app_settings via any execer (db or tx).
func execUpdateAppSettings(ctx context.Context, db execerContext, next model.AppSettings) (sql.Result, error) {
	return db.ExecContext(ctx, `
UPDATE app_settings
SET app_domain = ?,
    registration_enabled = ?,
    smtp_host = ?,
    smtp_port = ?,
    smtp_username = ?,
    smtp_password = CASE WHEN ? = '' THEN smtp_password ELSE ? END,
    smtp_from = ?,
    updated_at = ?
WHERE id = 1`,
		next.AppDomain,
		boolToInt(next.RegistrationEnabled),
		next.SMTPHost,
		next.SMTPPort,
		next.SMTPUsername,
		next.SMTPPassword,
		next.SMTPPassword,
		next.SMTPFrom,
		time.Now().UTC(),
	)
}

func normalizeAppSettings(settings model.AppSettings) model.AppSettings {
	settings.AppDomain = strings.TrimSpace(settings.AppDomain)
	settings.SMTPHost = strings.TrimSpace(settings.SMTPHost)
	settings.SMTPPort = strings.TrimSpace(settings.SMTPPort)
	if settings.SMTPPort == "" {
		settings.SMTPPort = "587"
	}
	settings.SMTPUsername = strings.TrimSpace(settings.SMTPUsername)
	settings.SMTPFrom = strings.TrimSpace(settings.SMTPFrom)
	return settings
}

func validateAppSettings(settings model.AppSettings) error {
	port, err := strconv.Atoi(settings.SMTPPort)
	if err != nil || port < 1 || port > 65535 {
		return ErrInvalidInput
	}
	return nil
}
