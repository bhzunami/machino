package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"machino/internal/model"
)

func (s *Store) ListProjects(ctx context.Context, userID string) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.title, p.description, p.color,
       CASE WHEN pf.user_id IS NULL THEN 0 ELSE 1 END,
       p.created_at, p.updated_at
FROM projects p
LEFT JOIN project_favorites pf ON pf.project_id = p.id AND pf.user_id = ?
ORDER BY CASE WHEN pf.user_id IS NULL THEN 1 ELSE 0 END, p.updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Color, &p.Favorite, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (s *Store) CreateProject(ctx context.Context, userID, title, description, color string) (model.Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Project{}, ErrInvalidInput
	}
	if strings.TrimSpace(color) == "" {
		color = "#4f46e5"
	}
	id, err := NewID()
	if err != nil {
		return model.Project{}, err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO projects (id, title, description, color, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, title, strings.TrimSpace(description), color, userID, now, now)
	if err != nil {
		return model.Project{}, fmt.Errorf("insert project: %w", err)
	}
	return model.Project{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(description),
		Color:       color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) UpdateProject(ctx context.Context, userID, projectID, title, description, color string) (model.Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Project{}, ErrInvalidInput
	}
	if strings.TrimSpace(color) == "" {
		color = "#4f46e5"
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`UPDATE projects SET title=?, description=?, color=?, updated_at=? WHERE id=? AND created_by=?`,
		title, strings.TrimSpace(description), color, now, projectID, userID)
	if err != nil {
		return model.Project{}, fmt.Errorf("update project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.Project{}, ErrNotFound
	}
	return model.Project{
		ID:          projectID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Color:       color,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) DeleteProject(ctx context.Context, userID, projectID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM projects WHERE id=? AND created_by=?`, projectID, userID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetFavorite(ctx context.Context, userID, projectID string, favorite bool) error {
	if favorite {
		_, err := s.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO project_favorites (user_id, project_id) VALUES (?, ?)`,
			userID, projectID)
		if err != nil {
			return fmt.Errorf("set favorite: %w", err)
		}
		return nil
	}
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM project_favorites WHERE user_id = ? AND project_id = ?`,
		userID, projectID); err != nil {
		return fmt.Errorf("unset favorite: %w", err)
	}
	return nil
}
