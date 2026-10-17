package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"
)

type sessionResolverStub struct {
	session     *domain.UserSession
	assignments []rbac.RoleAssignment
	err         error
}

func (s sessionResolverStub) GetSessionFromToken(string, context.Context) (*domain.UserSession, error) {
	return s.session, nil
}

func (s sessionResolverStub) GetRoleAssignmentsFromUserID(context.Context, string) ([]rbac.RoleAssignment, error) {
	return s.assignments, s.err
}

func TestAuthenticateStoresRoleAssignmentsWithScopeInContext(t *testing.T) {
	want := []rbac.RoleAssignment{{
		RoleID:         "role-1",
		RoleName:       "Organization Reader",
		ScopeType:      string(perms2.ScopeOrganization),
		OrganizationID: "org-1",
		Permissions:    []string{perms2.PermReadProjectBasic},
	}}
	resolver := sessionResolverStub{
		session: &domain.UserSession{
			SessionID: "session-1",
			UserID:    "user-1",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		assignments: want,
	}

	handler := Authenticate(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assignments, ok := GetCurrentUserRoleAssignments(r.Context())
		if !ok || len(assignments) != 1 {
			t.Fatalf("role assignments = %v, present = %v; want one assignment", assignments, ok)
		}
		got := assignments[0]
		if got.ScopeType != string(perms2.ScopeOrganization) || got.OrganizationID != "org-1" {
			t.Errorf("scope = (%q, %q), want organization org-1", got.ScopeType, got.OrganizationID)
		}
		if got.RoleID != "role-1" || len(got.Permissions) != 1 || got.Permissions[0] != perms2.PermReadProjectBasic {
			t.Errorf("assignment = %+v, want role and project-read permission", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: "token"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestAuthenticateRejectsWhenRoleAssignmentsCannotBeLoaded(t *testing.T) {
	resolver := sessionResolverStub{
		session: &domain.UserSession{
			SessionID: "session-1",
			UserID:    "user-1",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		err: errors.New("permission store unavailable"),
	}
	handler := Authenticate(resolver)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler ran despite role lookup failure")
	}))
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: "token"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}
