package middlewares

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

type authorizerStub struct {
	allowed bool
	err     error
}

func (s authorizerStub) IsAllowed(context.Context, string, perms2.PermKey, perms2.ScopeContext) (bool, error) {
	return s.allowed, s.err
}

func TestRequirePermissionLogsDeniedPermission(t *testing.T) {
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logOutput, nil))
	ctx := context.WithValue(context.Background(), keyUserID, "user-123")
	ctx = requestlogger.WithContext(ctx, logger)
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	//permission := perms2.Permission{ID: "READ_PROJECT_BASIC", DisplayName: "Read Project Basic"}

	handler := RequireGlobal(authorizerStub{}, perms2.PermReadProjectBasic, func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler ran despite permission denial")
	})
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	for _, field := range []string{
		`"msg":"permission denied"`,
		`"user_id":"user-123"`,
		`"permission_id":"READ_PROJECT_BASIC"`,
		`"scope":"GLOBAL"`,
	} {
		if !strings.Contains(logOutput.String(), field) {
			t.Errorf("log output %q does not contain %q", logOutput.String(), field)
		}
	}
}
