package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at,
			acp_enabled, acp_command, acp_args)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	autoStart := 0
	if m.AutoStartTerminal {
		autoStart = 1
	}
	acpEnabled := 0
	if m.ACPEnabled {
		acpEnabled = 1
	}
	// Serialize ACPArgs to JSON
	var acpArgsJSON interface{}
	if len(m.ACPArgs) > 0 {
		data, _ := json.Marshal(m.ACPArgs)
		acpArgsJSON = string(data)
	}
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.WorkspaceID, m.Name, m.RoleType, m.RoleKey, m.Avatar,
		m.TerminalType, m.TerminalCommand, m.TerminalPath, autoStart, m.Status, m.CreatedAt.Unix(),
		acpEnabled, m.ACPCommand, acpArgsJSON,
	)
	return err
}

func (r *sqlMemberRepo) GetByID(ctx context.Context, id string) (*models.Member, error) {
	query := `
		SELECT id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at,
			COALESCE(acp_enabled, 0), acp_command, acp_args,
			COALESCE(a2a_enabled, 0), a2a_agent_url, a2a_auth_type, a2a_auth_token
		FROM members WHERE id = ?
	`
	m := &models.Member{}
	var autoStart, acpEnabled, a2aEnabled int
	var createdAt int64
	var acpCommand, acpArgs sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.WorkspaceID, &m.Name, &m.RoleType, &m.RoleKey, &m.Avatar,
		&m.TerminalType, &m.TerminalCommand, &m.TerminalPath, &autoStart, &m.Status, &createdAt,
		&acpEnabled, &acpCommand, &acpArgs,
		&a2aEnabled, &m.A2AAgentURL, &m.A2AAuthType, &m.A2AAuthToken,
	)
	if err != nil {
		return nil, err
	}
	m.AutoStartTerminal = autoStart == 1
	m.ACPEnabled = acpEnabled == 1
	m.A2AEnabled = a2aEnabled == 1
	m.CreatedAt = time.Unix(createdAt, 0)
	if acpCommand.Valid {
		m.ACPCommand = acpCommand.String
	}
	if acpArgs.Valid && acpArgs.String != "" {
		json.Unmarshal([]byte(acpArgs.String), &m.ACPArgs)
	}
	return m, nil
}

func (r *sqlMemberRepo) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Member, error) {
	query := `
		SELECT id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at,
			COALESCE(acp_enabled, 0), acp_command, acp_args,
			COALESCE(a2a_enabled, 0), a2a_agent_url, a2a_auth_type, a2a_auth_token
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
		var autoStart, acpEnabled, a2aEnabled int
		var createdAt int64
		var acpCommand, acpArgs sql.NullString
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.Name, &m.RoleType, &m.RoleKey, &m.Avatar,
			&m.TerminalType, &m.TerminalCommand, &m.TerminalPath, &autoStart, &m.Status, &createdAt,
			&acpEnabled, &acpCommand, &acpArgs,
			&a2aEnabled, &m.A2AAgentURL, &m.A2AAuthType, &m.A2AAuthToken,
		); err != nil {
			return nil, err
		}
		m.AutoStartTerminal = autoStart == 1
		m.ACPEnabled = acpEnabled == 1
		m.A2AEnabled = a2aEnabled == 1
		m.CreatedAt = time.Unix(createdAt, 0)
		if acpCommand.Valid {
			m.ACPCommand = acpCommand.String
		}
		if acpArgs.Valid && acpArgs.String != "" {
			json.Unmarshal([]byte(acpArgs.String), &m.ACPArgs)
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *sqlMemberRepo) Update(ctx context.Context, m *models.Member) error {
	query := `
		UPDATE members SET name = ?, role_type = ?, role_key = ?, avatar = ?,
			terminal_type = ?, terminal_command = ?, terminal_path = ?, auto_start_terminal = ?, status = ?,
			acp_enabled = ?, acp_command = ?, acp_args = ?,
			a2a_enabled = ?, a2a_agent_url = ?, a2a_auth_type = ?, a2a_auth_token = ?
		WHERE id = ?
	`
	autoStart := 0
	if m.AutoStartTerminal {
		autoStart = 1
	}
	acpEnabled := 0
	if m.ACPEnabled {
		acpEnabled = 1
	}
	// Serialize ACPArgs to JSON
	var acpArgsJSON interface{}
	if len(m.ACPArgs) > 0 {
		data, _ := json.Marshal(m.ACPArgs)
		acpArgsJSON = string(data)
	}
	a2aEnabled := 0
	if m.A2AEnabled {
		a2aEnabled = 1
	}
	_, err := r.db.ExecContext(ctx, query,
		m.Name, m.RoleType, m.RoleKey, m.Avatar,
		m.TerminalType, m.TerminalCommand, m.TerminalPath, autoStart, m.Status,
		acpEnabled, m.ACPCommand, acpArgsJSON,
		a2aEnabled, m.A2AAgentURL, m.A2AAuthType, m.A2AAuthToken, m.ID,
	)
	return err
}

func (r *sqlMemberRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM members WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}