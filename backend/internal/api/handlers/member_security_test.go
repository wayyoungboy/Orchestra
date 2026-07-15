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

type crossWorkspaceMemberRepo struct {
	repository.MemberRepository
	member *models.Member
}

func (r crossWorkspaceMemberRepo) GetByID(context.Context, string) (*models.Member, error) {
	return r.member, nil
}

func TestMemberHandlersRejectCrossWorkspaceMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewMemberHandler(
		crossWorkspaceMemberRepo{member: &models.Member{ID: "member-1", WorkspaceID: "workspace-b"}},
		nil,
		nil,
	)

	tests := []struct {
		name   string
		method string
		call   func(*gin.Context)
	}{
		{"get", http.MethodGet, handler.Get},
		{"update", http.MethodPut, handler.Update},
		{"delete", http.MethodDelete, handler.Delete},
		{"presence", http.MethodPost, handler.UpdatePresence},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(tt.method, "/", nil)
			ctx.Params = gin.Params{{Key: "id", Value: "workspace-a"}, {Key: "memberId", Value: "member-1"}}
			tt.call(ctx)
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
			}
		})
	}
}
