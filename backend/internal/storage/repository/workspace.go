package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/orchestra/backend/internal/models"
)

type sqlWorkspaceRepo struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &sqlWorkspaceRepo{db: db}
}

func (r *sqlWorkspaceRepo) Create(ctx context.Context, ws *models.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, path, last_opened_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		ws.ID, ws.Name, ws.Path,
		ws.LastOpenedAt.Unix(), ws.CreatedAt.Unix(),
	)
	return err
}

func (r *sqlWorkspaceRepo) GetByID(ctx context.Context, id string) (*models.Workspace, error) {
	query := `
		SELECT id, name, path, last_opened_at, created_at
		FROM workspaces WHERE id = ?
	`
	ws := &models.Workspace{}
	var lastOpened, createdAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ws.ID, &ws.Name, &ws.Path, &lastOpened, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	ws.LastOpenedAt = time.Unix(lastOpened, 0)
	ws.CreatedAt = time.Unix(createdAt, 0)
	return ws, nil
}

func (r *sqlWorkspaceRepo) List(ctx context.Context) ([]*models.Workspace, error) {
	query := `
		SELECT id, name, path, last_opened_at, created_at
		FROM workspaces ORDER BY last_opened_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workspaces := make([]*models.Workspace, 0)
	for rows.Next() {
		ws := &models.Workspace{}
		var lastOpened, createdAt int64
		if err := rows.Scan(
			&ws.ID, &ws.Name, &ws.Path, &lastOpened, &createdAt,
		); err != nil {
			return nil, err
		}
		ws.LastOpenedAt = time.Unix(lastOpened, 0)
		ws.CreatedAt = time.Unix(createdAt, 0)
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (r *sqlWorkspaceRepo) Update(ctx context.Context, ws *models.Workspace) error {
	query := `UPDATE workspaces SET name = ?, path = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, ws.Name, ws.Path, ws.ID)
	return err
}

func (r *sqlWorkspaceRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM workspaces WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqlWorkspaceRepo) UpdateLastOpened(ctx context.Context, id string) error {
	query := `UPDATE workspaces SET last_opened_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	return err
}