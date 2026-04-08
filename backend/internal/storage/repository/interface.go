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

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id string) (*models.Task, error)
	ListByWorkspace(ctx context.Context, workspaceID string, status ...models.TaskStatus) ([]*models.Task, error)
	ListByAssignee(ctx context.Context, assigneeID string) ([]*models.Task, error)
	ListBySecretary(ctx context.Context, secretaryID string) ([]*models.Task, error)
	UpdateStatus(ctx context.Context, id string, status models.TaskStatus, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, key *models.APIKey) error
	GetByID(ctx context.Context, id string) (*models.APIKey, error)
	GetByProvider(ctx context.Context, provider models.APIKeyProvider) (*models.APIKey, error)
	List(ctx context.Context) ([]*models.APIKey, error)
	Update(ctx context.Context, key *models.APIKey) error
	Delete(ctx context.Context, id string) error
	DeleteByProvider(ctx context.Context, provider models.APIKeyProvider) error
}