package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/orchestra/backend/internal/models"
)

type sqlMemberRepo struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) MemberRepository {
	return &sqlMemberRepo{db: db}
}

func (r *sqlMemberRepo) Create(ctx context.Context, m *models.Member) error {
	query := `
		INSERT INTO members (id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	autoStart := 0
	if m.AutoStartTerminal {
		autoStart = 1
	}
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.WorkspaceID, m.Name, m.RoleType, m.RoleKey, m.Avatar,
		m.TerminalType, m.TerminalCommand, m.TerminalPath, autoStart, m.Status, m.CreatedAt.Unix(),
	)
	return err
}

func (r *sqlMemberRepo) GetByID(ctx context.Context, id string) (*models.Member, error) {
	query := `
		SELECT id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at
		FROM members WHERE id = ?
	`
	m := &models.Member{}
	var autoStart int
	var createdAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.WorkspaceID, &m.Name, &m.RoleType, &m.RoleKey, &m.Avatar,
		&m.TerminalType, &m.TerminalCommand, &m.TerminalPath, &autoStart, &m.Status, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	m.AutoStartTerminal = autoStart == 1
	m.CreatedAt = time.Unix(createdAt, 0)
	return m, nil
}

func (r *sqlMemberRepo) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Member, error) {
	query := `
		SELECT id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at
		FROM members WHERE workspace_id = ? ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.Member, 0)
	for rows.Next() {
		m := &models.Member{}
		var autoStart int
		var createdAt int64
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.Name, &m.RoleType, &m.RoleKey, &m.Avatar,
			&m.TerminalType, &m.TerminalCommand, &m.TerminalPath, &autoStart, &m.Status, &createdAt,
		); err != nil {
			return nil, err
		}
		m.AutoStartTerminal = autoStart == 1
		m.CreatedAt = time.Unix(createdAt, 0)
		members = append(members, m)
	}
	return members, nil
}

func (r *sqlMemberRepo) Update(ctx context.Context, m *models.Member) error {
	query := `
		UPDATE members SET name = ?, role_type = ?, role_key = ?, avatar = ?,
			terminal_type = ?, terminal_command = ?, terminal_path = ?, auto_start_terminal = ?, status = ?
		WHERE id = ?
	`
	autoStart := 0
	if m.AutoStartTerminal {
		autoStart = 1
	}
	_, err := r.db.ExecContext(ctx, query,
		m.Name, m.RoleType, m.RoleKey, m.Avatar,
		m.TerminalType, m.TerminalCommand, m.TerminalPath, autoStart, m.Status, m.ID,
	)
	return err
}

func (r *sqlMemberRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM members WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}