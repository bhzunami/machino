package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bhzunami/machino/internal/model"
)

func (s *Store) ListProjects(ctx context.Context, userID string) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.title, p.description, p.color, p.move_done,
       CASE WHEN pf.user_id IS NULL THEN 0 ELSE 1 END AS favorite,
       CASE WHEN p.created_by = ? THEN 1 ELSE 0 END AS is_owner,
       (SELECT COUNT(1) FROM project_members pm2 WHERE pm2.project_id = p.id) AS member_count,
       p.created_at, p.updated_at
FROM projects p
JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = ?
LEFT JOIN project_favorites pf ON pf.project_id = p.id AND pf.user_id = ?
ORDER BY CASE WHEN pf.user_id IS NULL THEN 1 ELSE 0 END, p.updated_at DESC`,
		userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var projects []model.Project
	for rows.Next() {
		var p model.Project
		var moveDone int
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Color,
			&moveDone, &p.Favorite, &p.IsOwner, &p.MemberCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		p.MoveDone = moveDone == 1
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

// GetProject returns a project if the user is a member (owner or shared), or ErrNotFound.
func (s *Store) GetProject(ctx context.Context, userID, projectID string) (model.Project, error) {
	var p model.Project
	row := s.db.QueryRowContext(ctx, `
SELECT p.id, p.title, p.description, p.color, p.move_done,
       CASE WHEN p.created_by = ? THEN 1 ELSE 0 END AS is_owner,
       (SELECT COUNT(1) FROM project_members pm2 WHERE pm2.project_id = p.id) AS member_count,
       p.created_at, p.updated_at
FROM projects p
JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = ?
WHERE p.id = ?`, userID, userID, projectID)
	var moveDone int
	if err := row.Scan(&p.ID, &p.Title, &p.Description, &p.Color,
		&moveDone, &p.IsOwner, &p.MemberCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Project{}, ErrNotFound
		}
		return model.Project{}, fmt.Errorf("get project: %w", err)
	}
	p.MoveDone = moveDone == 1
	return p, nil
}

func (s *Store) CreateProject(ctx context.Context, userID, title, description, color string, moveDone bool) (model.Project, error) {
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Project{}, fmt.Errorf("begin create project tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO projects (id, title, description, color, move_done, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, title, strings.TrimSpace(description), color, boolToInt(moveDone), userID, now, now); err != nil {
		return model.Project{}, fmt.Errorf("insert project: %w", err)
	}
	// Add creator as owner in project_members.
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role, joined_at) VALUES (?, ?, 'owner', ?)`,
		id, userID, now); err != nil {
		return model.Project{}, fmt.Errorf("insert project owner member: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return model.Project{}, fmt.Errorf("commit create project tx: %w", err)
	}
	return model.Project{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(description),
		Color:       color,
		MoveDone:    moveDone,
		IsOwner:     true,
		MemberCount: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) UpdateProject(ctx context.Context, userID, projectID, title, description, color string, moveDone bool) (model.Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Project{}, ErrInvalidInput
	}
	if strings.TrimSpace(color) == "" {
		color = "#4f46e5"
	}
	now := time.Now().UTC()
	// Only the owner (created_by) can update project metadata.
	res, err := s.db.ExecContext(ctx,
		`UPDATE projects SET title=?, description=?, color=?, move_done=?, updated_at=? WHERE id=? AND created_by=?`,
		title, strings.TrimSpace(description), color, boolToInt(moveDone), now, projectID, userID)
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
		MoveDone:    moveDone,
		IsOwner:     true,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) DeleteProject(ctx context.Context, userID, projectID string) error {
	// Only the owner (created_by) can delete a project.
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
	// Any member can favorite — verify membership via GetProject.
	if _, err := s.GetProject(ctx, userID, projectID); err != nil {
		return err
	}
	if favorite {
		if _, err := s.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO project_favorites (user_id, project_id) VALUES (?, ?)`,
			userID, projectID); err != nil {
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

// ListMembers returns all members of a project. Any member may call this.
func (s *Store) ListMembers(ctx context.Context, userID, projectID string) ([]model.ProjectMember, error) {
	if _, err := s.GetProject(ctx, userID, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT u.id, u.name, u.email, pm.role, pm.joined_at
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = ?
ORDER BY CASE WHEN pm.role = 'owner' THEN 0 ELSE 1 END, pm.joined_at ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()
	var members []model.ProjectMember
	for rows.Next() {
		var m model.ProjectMember
		if err := rows.Scan(&m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// AddMember invites a user (by email) to a project. Only the owner may do this.
func (s *Store) AddMember(ctx context.Context, ownerID, projectID, email string) (model.ProjectMember, error) {
	email = NormalizeEmail(email)
	if email == "" || projectID == "" {
		return model.ProjectMember{}, ErrInvalidInput
	}
	// Verify requester is the project owner.
	proj, err := s.GetProject(ctx, ownerID, projectID)
	if err != nil {
		return model.ProjectMember{}, err
	}
	if !proj.IsOwner {
		return model.ProjectMember{}, ErrUnauthorized
	}
	// Look up the invitee.
	var invitee model.User
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, name, email FROM users WHERE email = ?`, email).
		Scan(&invitee.ID, &invitee.Name, &invitee.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ProjectMember{}, ErrNotFound
		}
		return model.ProjectMember{}, fmt.Errorf("lookup user: %w", err)
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO project_members (project_id, user_id, role, joined_at) VALUES (?, ?, 'member', ?)`,
		projectID, invitee.ID, now); err != nil {
		return model.ProjectMember{}, fmt.Errorf("insert member: %w", err)
	}
	// Read back actual role (may already have been owner).
	var m model.ProjectMember
	if err := s.db.QueryRowContext(ctx,
		`SELECT u.id, u.name, u.email, pm.role, pm.joined_at
 FROM project_members pm JOIN users u ON u.id = pm.user_id
 WHERE pm.project_id = ? AND pm.user_id = ?`, projectID, invitee.ID).
		Scan(&m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt); err != nil {
		return model.ProjectMember{}, fmt.Errorf("read new member: %w", err)
	}
	return m, nil
}

// AddMemberByUserID adds a user (by ID) to a project. Only the owner may do this.
func (s *Store) AddMemberByUserID(ctx context.Context, ownerID, projectID, inviteeID string) (model.ProjectMember, error) {
	if inviteeID == "" || projectID == "" {
		return model.ProjectMember{}, ErrInvalidInput
	}
	proj, err := s.GetProject(ctx, ownerID, projectID)
	if err != nil {
		return model.ProjectMember{}, err
	}
	if !proj.IsOwner {
		return model.ProjectMember{}, ErrUnauthorized
	}
	var invitee model.User
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, name, email FROM users WHERE id = ?`, inviteeID).
		Scan(&invitee.ID, &invitee.Name, &invitee.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ProjectMember{}, ErrNotFound
		}
		return model.ProjectMember{}, fmt.Errorf("lookup user by id: %w", err)
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO project_members (project_id, user_id, role, joined_at) VALUES (?, ?, 'member', ?)`,
		projectID, invitee.ID, now); err != nil {
		return model.ProjectMember{}, fmt.Errorf("insert member by id: %w", err)
	}
	var m model.ProjectMember
	if err := s.db.QueryRowContext(ctx,
		`SELECT u.id, u.name, u.email, pm.role, pm.joined_at
 FROM project_members pm JOIN users u ON u.id = pm.user_id
 WHERE pm.project_id = ? AND pm.user_id = ?`, projectID, invitee.ID).
		Scan(&m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt); err != nil {
		return model.ProjectMember{}, fmt.Errorf("read new member by id: %w", err)
	}
	return m, nil
}

// RemoveMember removes a member from a project.
// Only the owner may remove others; any member may remove themselves (leave).
func (s *Store) RemoveMember(ctx context.Context, requesterID, projectID, memberUserID string) error {
	proj, err := s.GetProject(ctx, requesterID, projectID)
	if err != nil {
		return err
	}
	// Must be owner to remove others; anyone can remove themselves.
	if requesterID != memberUserID && !proj.IsOwner {
		return ErrUnauthorized
	}
	// Cannot remove the owner.
	if memberUserID == requesterID && proj.IsOwner {
		return fmt.Errorf("owner cannot leave their own project: %w", ErrUnauthorized)
	}
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id = ? AND user_id = ? AND role != 'owner'`,
		projectID, memberUserID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
