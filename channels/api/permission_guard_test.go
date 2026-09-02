package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/app/platform"
)

func TestPermissionGuardRequiresRolePermission(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("platform.New() error = %v", err)
	}

	server := app.NewServerWithOptions(ps)
	a := app.New(app.ServerConnector(server.Channels()))

	claims := jwt.MapClaims{
		"sub":   "u-1",
		"roles": []string{"user"},
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token error = %v", err)
	}

	var statusCode int
	h := &Handler{
		App:               a,
		RequireSession:    true,
		RequirePermission: "project.read",
		HandleFunc: func(c *Context, w http.ResponseWriter, r *http.Request) {
			statusCode = http.StatusOK
			w.WriteHeader(http.StatusOK)
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed permission, got %d body=%s", resp.Code, resp.Body.String())
	}
	if statusCode != http.StatusOK {
		t.Fatalf("handler was not executed")
	}
}

func TestIssueAndAdminRoutesEnforceRequiredPermissions(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("platform.New() error = %v", err)
	}

	server := app.NewServerWithOptions(ps)
	Init(server)

	claims := jwt.MapClaims{
		"sub":   "u-1",
		"roles": []string{"profile.read"},
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/issues", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp := httptest.NewRecorder()
	server.Router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing issue.read permission, got %d body=%s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp = httptest.NewRecorder()
	server.Router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing system.admin permission, got %d body=%s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/search/assist", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp = httptest.NewRecorder()
	server.Router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing search.read permission, got %d body=%s", resp.Code, resp.Body.String())
	}
}
