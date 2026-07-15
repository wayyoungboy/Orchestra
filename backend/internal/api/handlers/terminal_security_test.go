package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
)

type terminalWorkspaceRepo struct {
	repository.WorkspaceRepository
	workspace *models.Workspace
}

func (r terminalWorkspaceRepo) GetByID(context.Context, string) (*models.Workspace, error) {
	return r.workspace, nil
}

type terminalMemberRepo struct {
	repository.MemberRepository
	member *models.Member
}

func (r terminalMemberRepo) GetByID(context.Context, string) (*models.Member, error) {
	return r.member, nil
}

func TestTerminalSessionConfigUsesPersistedMemberAndWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	member := &models.Member{
		ID:           "member-1",
		WorkspaceID:  "workspace-1",
		Name:         "Persisted agent",
		TerminalType: "acp",
		ACPEnabled:   true,
		ACPCommand:   "codex",
		ACPArgs:      []string{"exec"},
	}
	handler := NewTerminalHandler(
		nil,
		terminalWorkspaceRepo{workspace: &models.Workspace{ID: "workspace-1", Path: "/trusted/workspace"}},
		terminalMemberRepo{member: member},
	)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/terminal/sessions", nil)

	config, err := handler.sessionConfig(ctx, "workspace-1", "member-1")
	if err != nil {
		t.Fatalf("sessionConfig() error = %v", err)
	}
	if config.WorkspaceDir != "/trusted/workspace" || config.Member != member {
		t.Fatalf("sessionConfig() used request-controlled values: %+v", config)
	}
	if config.MemberName != member.Name || config.TerminalType != member.TerminalType {
		t.Fatalf("sessionConfig() did not preserve member configuration: %+v", config)
	}
}

func TestTerminalSessionConfigRejectsCrossWorkspaceMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTerminalHandler(
		nil,
		terminalWorkspaceRepo{workspace: &models.Workspace{ID: "workspace-1", Path: "/trusted/workspace"}},
		terminalMemberRepo{member: &models.Member{ID: "member-1", WorkspaceID: "workspace-2"}},
	)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/terminal/sessions", nil)
	if _, err := handler.sessionConfig(ctx, "workspace-1", "member-1"); err == nil {
		t.Fatal("member from another workspace should be rejected")
	}
}
