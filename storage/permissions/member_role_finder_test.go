package permstorage

import (
	"context"
	"strings"
	"testing"
)

func TestMemberRoleFinderQueriesUseProjectScopedAssignments(t *testing.T) {
	for _, fragment := range []string{
		"ar.group_id = $1",
		"ar.scope_type = 'PROJECT'",
		"ar.project_id = $2",
		"ar.organization_id IS NULL",
		"SELECT DISTINCT ar.role_id",
	} {
		if !strings.Contains(queryRoleIDsForGroupInProject, fragment) {
			t.Errorf("group role query is missing %q", fragment)
		}
	}
}

func TestMemberRoleFinderEmptyInputReturnsEmptyRoles(t *testing.T) {
	repo := NewPostgresRepository(nil)
	ctx := context.Background()

	userRoles, err := repo.GetRoleIDsByUserForProjectScope(ctx, "", "project-1")
	if err != nil || userRoles == nil || len(userRoles) != 0 {
		t.Errorf("empty user ID returned roles %v and error %v, want a non-nil empty slice", userRoles, err)
	}

	groupRoles, err := repo.GetRoleIDsByGroupForProjectScope(ctx, "group-1", "")
	if err != nil || groupRoles == nil || len(groupRoles) != 0 {
		t.Errorf("empty project ID returned roles %v and error %v, want a non-nil empty slice", groupRoles, err)
	}
}
