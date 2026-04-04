package repository

import (
	"context"

	"github.com/orchestra/backend/internal/models"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *models.Workspace) error
	GetByID(ctx context.Context, id string) (*models.Workspace, error)
	List(ctx context.Context) ([]*models.Workspace, error)
	Update(ctx context.Context, ws *models.Workspace) error
	Delete(ctx context.Context, id string) error
	UpdateLastOpened(ctx context.Context, id string) error
}

type MemberRepository interface {
	Create(ctx context.Context, m *models.Member) error
	GetByID(ctx context.Context, id string) (*models.Member, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Member, error)
	Update(ctx context.Context, m *models.Member) error
	Delete(ctx context.Context, id string) error
}