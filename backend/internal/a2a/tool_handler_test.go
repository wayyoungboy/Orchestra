package a2a

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
)

type toolHandlerWorkspaceRepo struct {
	repository.WorkspaceRepository
	workspace *models.Workspace
}

func (r toolHandlerWorkspaceRepo) GetByID(context.Context, string) (*models.Workspace, error) {
	return r.workspace, nil
}

func TestFileToolsUseWorkspacePathAndRejectEscapes(t *testing.T) {
	workspaceDir := t.TempDir()
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspaceDir, "notes.txt"), []byte("workspace content"), 0o600); err != nil {
		t.Fatalf("write workspace file: %v", err)
	}

	validator := filesystem.NewValidator([]string{workspaceDir})
	handler := NewToolHandler(nil, nil, nil, nil, nil, filesystem.NewBrowser(validator), validator)
	handler.SetWorkspaceRepo(toolHandlerWorkspaceRepo{workspace: &models.Workspace{ID: "workspace-id", Path: workspaceDir}})
	session := NewSession("session-id", "workspace-id", "member-id", "agent", "", nil)

	read := handler.ExecuteTool(toolUse(ToolFileRead, FileReadInput{Path: "notes.txt"}), session)
	if read.IsError || read.Content != "workspace content" {
		t.Fatalf("file read = %#v, want workspace content", read)
	}

	write := handler.ExecuteTool(toolUse(ToolFileWrite, FileWriteInput{Path: "generated/output.txt", Content: "created by agent"}), session)
	if write.IsError {
		t.Fatalf("file write failed: %#v", write)
	}
	content, err := os.ReadFile(filepath.Join(workspaceDir, "generated", "output.txt"))
	if err != nil || string(content) != "created by agent" {
		t.Fatalf("generated file = %q, %v", content, err)
	}

	escapePath, err := filepath.Rel(workspaceDir, filepath.Join(outsideDir, "secret.txt"))
	if err != nil {
		t.Fatalf("relative escape path: %v", err)
	}
	escape := handler.ExecuteTool(toolUse(ToolFileWrite, FileWriteInput{Path: escapePath, Content: "must not write"}), session)
	if !escape.IsError || !strings.Contains(escape.Content, "outside the workspace") {
		t.Fatalf("workspace escape should be rejected, got %#v", escape)
	}
	if _, err := os.Stat(filepath.Join(outsideDir, "secret.txt")); !os.IsNotExist(err) {
		t.Fatalf("escape write created a file outside the workspace: %v", err)
	}
}

func toolUse(name string, input any) *ACPMessage {
	rawInput, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	rawToolUse, err := json.Marshal(ToolUseMessage{
		Type:      TypeToolUse,
		Name:      name,
		Input:     rawInput,
		ToolUseID: "tool-use-id",
	})
	if err != nil {
		panic(err)
	}
	return &ACPMessage{Type: TypeToolUse, Content: rawToolUse}
}
