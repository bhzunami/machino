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

func (s *Store) ListTodos(ctx context.Context, userID, projectID string) ([]model.Todo, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, project_id, title, description, due_date, priority, completed, position, created_at, updated_at
FROM todos
WHERE project_id = ? AND EXISTS (SELECT 1 FROM project_members WHERE project_id = ? AND user_id = ?)
ORDER BY position ASC, created_at ASC`, projectID, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()
	var todos []model.Todo
	for rows.Next() {
		var t model.Todo
		var due sql.NullTime
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &due, &t.Priority, &t.Completed, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		if due.Valid {
			t.DueDate = &due.Time
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todos: %w", err)
	}
	return todos, nil
}

func (s *Store) CreateTodo(ctx context.Context, userID, projectID, title, description, priority string, dueDate *time.Time) (model.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" || projectID == "" {
		return model.Todo{}, ErrInvalidInput
	}
	if priority == "" {
		priority = "normal"
	}
	id, err := NewID()
	if err != nil {
		return model.Todo{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Todo{}, fmt.Errorf("begin create todo tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx,
		`UPDATE todos SET position = position + 1 WHERE project_id = ?`, projectID); err != nil {
		return model.Todo{}, fmt.Errorf("shift todo positions: %w", err)
	}
	position := 1
	now := time.Now().UTC()
	var due any
	if dueDate != nil {
		due = dueDate.UTC()
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO todos (id, project_id, title, description, due_date, priority, completed, position, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)`,
		id, projectID, title, strings.TrimSpace(description), due, priority, position, userID, now, now); err != nil {
		return model.Todo{}, fmt.Errorf("insert todo: %w", err)
	}
	// Verify membership before committing.
	var memberCount int
	if err = tx.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM project_members WHERE project_id = ? AND user_id = ?`, projectID, userID).Scan(&memberCount); err != nil {
		err = fmt.Errorf("check project membership: %w", err)
		return model.Todo{}, err
	}
	if memberCount == 0 {
		err = ErrNotFound
		return model.Todo{}, err
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE projects SET updated_at = ? WHERE id = ?`, now, projectID); err != nil {
		return model.Todo{}, fmt.Errorf("touch project: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return model.Todo{}, fmt.Errorf("commit create todo tx: %w", err)
	}
	return model.Todo{
		ID:          id,
		ProjectID:   projectID,
		Title:       title,
		Description: strings.TrimSpace(description),
		DueDate:     dueDate,
		Priority:    priority,
		Position:    position,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) UpdateTodo(ctx context.Context, userID, todoID string, completed *bool, title, description, priority *string, dueDate **time.Time) (model.Todo, error) {
	current, err := s.todoByID(ctx, todoID)
	if err != nil {
		return model.Todo{}, err
	}
	if completed != nil {
		current.Completed = *completed
	}
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return model.Todo{}, ErrInvalidInput
		}
		current.Title = trimmed
	}
	if description != nil {
		current.Description = strings.TrimSpace(*description)
	}
	if priority != nil && strings.TrimSpace(*priority) != "" {
		current.Priority = strings.TrimSpace(*priority)
	}
	if dueDate != nil {
		current.DueDate = *dueDate
	}
	now := time.Now().UTC()
	var due any
	if current.DueDate != nil {
		due = current.DueDate.UTC()
	}
	// Include project ownership check in the UPDATE to prevent IDOR.
	result, err := s.db.ExecContext(ctx,
		`UPDATE todos SET title = ?, description = ?, due_date = ?, priority = ?, completed = ?, updated_at = ?
		 WHERE id = ? AND project_id IN (SELECT project_id FROM project_members WHERE project_id = todos.project_id AND user_id = ?)`,
		current.Title, current.Description, due, current.Priority, boolToInt(current.Completed), now, todoID, userID)
	if err != nil {
		return model.Todo{}, fmt.Errorf("update todo: %w", err)
	}
	if err := requireAffected(result); err != nil {
		return model.Todo{}, err
	}
	current.UpdatedAt = now
	return current, nil
}

func (s *Store) DeleteCompletedTodos(ctx context.Context, userID, projectID string) error {
	if projectID == "" {
		return ErrInvalidInput
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM todos WHERE project_id = ? AND completed = 1
		 AND EXISTS (SELECT 1 FROM project_members WHERE project_id = ? AND user_id = ?)`,
		projectID, projectID, userID)
	if err != nil {
		return fmt.Errorf("delete completed todos: %w", err)
	}
	return nil
}

func (s *Store) ReorderTodos(ctx context.Context, userID, projectID string, ids []string) error {
	if projectID == "" || len(ids) == 0 {
		return ErrInvalidInput
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reorder tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	// Verify membership before reordering.
	var memberCount int
	if err = tx.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM project_members WHERE project_id = ? AND user_id = ?`, projectID, userID).Scan(&memberCount); err != nil {
		return fmt.Errorf("check project membership: %w", err)
	}
	if memberCount == 0 {
		err = ErrNotFound
		return err
	}
	for i, id := range ids {
		result, execErr := tx.ExecContext(ctx,
			`UPDATE todos SET position = ?, updated_at = ? WHERE id = ? AND project_id = ?`,
			i+1, time.Now().UTC(), id, projectID)
		if execErr != nil {
			err = fmt.Errorf("update todo position: %w", execErr)
			return err
		}
		if execErr := requireAffected(result); execErr != nil {
			err = execErr
			return err
		}
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE projects SET updated_at = ? WHERE id = ?`, time.Now().UTC(), projectID); err != nil {
		return fmt.Errorf("touch reordered project: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit reorder tx: %w", err)
	}
	return nil
}

func (s *Store) todoByID(ctx context.Context, todoID string) (model.Todo, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, title, description, due_date, priority, completed, position, created_at, updated_at FROM todos WHERE id = ?`,
		todoID)
	var t model.Todo
	var due sql.NullTime
	if err := row.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &due, &t.Priority, &t.Completed, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Todo{}, ErrNotFound
		}
		return model.Todo{}, fmt.Errorf("select todo: %w", err)
	}
	if due.Valid {
		t.DueDate = &due.Time
	}
	return t, nil
}
