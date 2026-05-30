package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bhzunami/machino/internal/model"
)

func (s *Store) ListColumns(ctx context.Context, userID, projectID string) ([]model.Column, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, project_id, title, position, created_at, updated_at
FROM project_columns
WHERE project_id = ? AND EXISTS (SELECT 1 FROM project_members WHERE project_id = ? AND user_id = ?)
ORDER BY position ASC, created_at ASC`, projectID, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("list columns: %w", err)
	}
	defer rows.Close()
	cols := []model.Column{}
	for rows.Next() {
		var c model.Column
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Title, &c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate columns: %w", err)
	}
	return cols, nil
}

func (s *Store) CreateColumn(ctx context.Context, userID, projectID, title string) (model.Column, error) {
	title = strings.TrimSpace(title)
	if title == "" || projectID == "" {
		return model.Column{}, ErrInvalidInput
	}
	var memberCount int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM project_members WHERE project_id = ? AND user_id = ?`, projectID, userID,
	).Scan(&memberCount); err != nil {
		return model.Column{}, fmt.Errorf("check project membership: %w", err)
	}
	if memberCount == 0 {
		return model.Column{}, ErrNotFound
	}
	id, err := NewID()
	if err != nil {
		return model.Column{}, err
	}
	now := time.Now().UTC()
	var maxPos int
	_ = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(position), 0) FROM project_columns WHERE project_id = ?`, projectID,
	).Scan(&maxPos)
	position := maxPos + 1
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO project_columns (id, project_id, title, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, projectID, title, position, now, now,
	); err != nil {
		return model.Column{}, fmt.Errorf("insert column: %w", err)
	}
	return model.Column{
		ID:        id,
		ProjectID: projectID,
		Title:     title,
		Position:  position,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Store) UpdateColumn(ctx context.Context, userID, columnID, title string) (model.Column, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Column{}, ErrInvalidInput
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `
UPDATE project_columns SET title = ?, updated_at = ?
WHERE id = ? AND project_id IN (
  SELECT project_id FROM project_members WHERE project_id = project_columns.project_id AND user_id = ?
)`, title, now, columnID, userID)
	if err != nil {
		return model.Column{}, fmt.Errorf("update column: %w", err)
	}
	if err := requireAffected(result); err != nil {
		return model.Column{}, err
	}
	var c model.Column
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, title, position, created_at, updated_at FROM project_columns WHERE id = ?`, columnID,
	).Scan(&c.ID, &c.ProjectID, &c.Title, &c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return model.Column{}, fmt.Errorf("fetch updated column: %w", err)
	}
	return c, nil
}

func (s *Store) DeleteColumn(ctx context.Context, userID, columnID string) (string, error) {
	var projectID string
	if err := s.db.QueryRowContext(ctx,
		`SELECT project_id FROM project_columns WHERE id = ?`, columnID,
	).Scan(&projectID); err != nil {
		return "", fmt.Errorf("fetch column project: %w", err)
	}
	result, err := s.db.ExecContext(ctx, `
DELETE FROM project_columns
WHERE id = ? AND project_id IN (
  SELECT project_id FROM project_members WHERE project_id = ? AND user_id = ?
)`, columnID, projectID, userID)
	if err != nil {
		return "", fmt.Errorf("delete column: %w", err)
	}
	return projectID, requireAffected(result)
}

func (s *Store) ReorderColumns(ctx context.Context, userID, projectID string, ids []string) error {
	if projectID == "" || len(ids) == 0 {
		return ErrInvalidInput
	}
	var memberCount int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM project_members WHERE project_id = ? AND user_id = ?`, projectID, userID,
	).Scan(&memberCount); err != nil {
		return fmt.Errorf("check project membership: %w", err)
	}
	if memberCount == 0 {
		return ErrNotFound
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reorder columns tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	now := time.Now().UTC()
	for i, id := range ids {
		result, execErr := tx.ExecContext(ctx,
			`UPDATE project_columns SET position = ?, updated_at = ? WHERE id = ? AND project_id = ?`,
			i+1, now, id, projectID)
		if execErr != nil {
			err = fmt.Errorf("update column position: %w", execErr)
			return err
		}
		if execErr := requireAffected(result); execErr != nil {
			err = execErr
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit reorder columns tx: %w", err)
	}
	return nil
}
