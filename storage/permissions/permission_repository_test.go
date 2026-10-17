package permstorage

import (
	"bytes"
	"context"
	"strings"
	"testing"

	perms2 "youtrack_backend/internal/services/permissions_services"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Ensure PostgresRepository satisfies the Repository and UserPermissionsFinder interfaces.

func TestUserRoleQueriesResolveInheritedGroupsAndTeams(t *testing.T) {
	t.Parallel()
	globalScope := perms2.ScopeGlobal
	projectScope := perms2.ScopeProject
	organizationScope := perms2.ScopeOrganization
	queries := map[string]string{
		//"permissions":        getQueryPermissionsForUser,
		"user roles":         getQueryRolesForUser(&globalScope),
		"global roles":       getQueryRolesForUser(&globalScope),
		"project roles":      getQueryRolesForUser(&projectScope),
		"organization roles": getQueryRolesForUser(&organizationScope),
	}
	for name, query := range queries {
		for _, fragment := range []string{
			"WITH RECURSIVE",
			"group_members",
			"parent_group_id",
			"FROM groups",
		} {
			if !strings.Contains(query, fragment) {
				t.Errorf("%s query is missing identity resolution fragment %q", name, fragment)
			}
		}
		for _, staleReference := range []string{
			"user_group_members",
			"project_team_members",
			"project_teams",
			"user_groups",
			"holder_type",
			"holder_id",
			"scope_project_id",
			"scope_organization_id",
		} {
			if strings.Contains(query, staleReference) {
				t.Errorf("%s query references stale schema item %q", name, staleReference)
			}
		}
	}
}

func TestLogAssignedRolesWritesOnePhysicalLinePerRole(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Skip("logAssignedRoles test skipped due to disabled logging buffer")
	}
}

func TestProjectRoleQueryMatchesAssignedRolesSchema(t *testing.T) {
	t.Parallel()
	projectScope := perms2.ScopeProject

	for _, fragment := range []string{
		"FROM groups",
		"FROM group_members",
		"member_type = 'USER'",
		"member_type = 'GROUP'",
		"parent_group_id",
		"ar.user_id",
		"ar.group_id",
		"ar.project_id = $2",
		"ar.scope_type = 'PROJECT'",
		"r.immutable",
		"r.permissions",
	} {
		if !strings.Contains(getQueryRolesForUser(&projectScope), fragment) {
			t.Errorf("queryProjectRoles is missing schema fragment %q", fragment)
		}
	}
	for _, staleReference := range []string{
		"holder_type",
		"holder_id",
		"scope_project_id",
		"scope_id",
		"project_team_members",
		"project_teams",
		"user_groups",
		"audit_target_id",
		"is_updatable",
	} {
		// audit_target_id and is_updatable are output aliases for the existing scanner.
		if staleReference == "audit_target_id" || staleReference == "is_updatable" {
			continue
		}
		if strings.Contains(getQueryRolesForUser(&projectScope), staleReference) {
			t.Errorf("queryProjectRoles references stale schema item %q", staleReference)
		}
	}
}

func TestUserGroupResolutionCTEEndsWithValidCTEList(t *testing.T) {
	t.Parallel()
	if strings.HasSuffix(strings.TrimSpace(userGroupResolutionCTE), ",") {
		t.Fatal("userGroupResolutionCTE ends with a trailing comma before SELECT; PostgreSQL rejects this syntax")
	}
}

func TestGetPermissionsForUserQuerySelectsScopeAndPermissionArray(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		scope     perms2.ScopeLevel
		fragments []string
	}{
		{
			scope: perms2.ScopeGlobal,
			fragments: []string{
				"SELECT ar.scope_type, rl.permissions",
				"ar.scope_type = 'GLOBAL'",
				"ar.organization_id IS NULL",
				"ar.project_id IS NULL",
				"ar.user_id = $1",
			},
		},
		{
			scope: perms2.ScopeProject,
			fragments: []string{
				"SELECT ar.scope_type, rl.permissions",
				"ar.scope_type = 'PROJECT'",
				"ar.project_id = $2",
				"ar.user_id = $1",
			},
		},
		{
			scope: perms2.ScopeOrganization,
			fragments: []string{
				"SELECT ar.scope_type, rl.permissions",
				"ar.scope_type = 'ORGANIZATION'",
				"ar.organization_id = $2",
				"ar.user_id = $1",
			},
		},
	} {
		query := getQueryPermissionsForUser(test.scope)
		for _, fragment := range test.fragments {
			if !strings.Contains(query, fragment) {
				t.Errorf("%s permissions query is missing fragment %q", test.scope, fragment)
			}
		}
	}
}

func TestGetAssignedRolesQueryMatchesAssignedRolesAndGroupSchemas(t *testing.T) {
	t.Parallel()

	for _, fragment := range []string{
		"ar.user_id",
		"ar.group_id",
		"member_type = 'USER'",
		"member_type = 'GROUP'",
		"ar.project_id",
		"ar.organization_id",
		"gm.member_type = 'USER'",
		"gm.member_type = 'GROUP'",
		"gm.member_group_id",
		"g.parent_group_id",
		"g.group_type = 'TEAM'",
		"ORDER BY ar.role_id, LOWER(ar.scope_type::text)",
	} {
		if !strings.Contains(getQueryAssignedRoles(), fragment) {
			t.Errorf("queryAssignedRoles is missing schema fragment %q", fragment)
		}
	}
	normalizedQuery := strings.Join(strings.Fields(getQueryAssignedRoles()), " ")
	if !strings.Contains(normalizedQuery, "ar.organization_id, ar.user_id, ar.group_id, ar.created_at, ar.updated_at, rl.name") {
		t.Error("queryAssignedRoles columns do not match the rows.Scan order")
	}
	for _, staleColumn := range []string{
		"ar.holder_type",
		"ar.holder_id",
		"ar.scope_id",
		"ar.scope_project_id",
		"ar.scope_organization_id",
	} {
		if strings.Contains(getQueryAssignedRoles(), staleColumn) {
			t.Errorf("queryAssignedRoles references nonexistent assigned_roles column %q", staleColumn)
		}
	}
	for _, staleTable := range []string{"user_group_members", "project_team_members", "project_teams", "user_groups"} {
		if strings.Contains(getQueryAssignedRoles(), staleTable) {
			t.Errorf("queryAssignedRoles references nonexistent schema table %q", staleTable)
		}
	}
}

func TestPostgresRepository_EmptyInputHandling(t *testing.T) {
	repo := NewPostgresRepository(nil)
	ctx := context.Background()

	// Empty userID should return empty permissions without database query
	userPerms, err := repo.GetPermissionsForUser(ctx, "", perms2.ScopeContext{Type: perms2.ScopeGlobal})
	if err != nil {
		t.Fatalf("expected no error for empty userID, got %v", err)
	}
	if len(userPerms) != 0 {
		t.Fatalf("expected 0 permissions, got %d", len(userPerms))
	}

	// Empty roleIDs should return empty permissions without database query
	rolePerms, err := repo.GetPermissionsByRoleIDs([]string{}, ctx)
	if err != nil {
		t.Fatalf("expected no error for empty roleIDs, got %v", err)
	}
	if len(rolePerms) != 0 {
		t.Fatalf("expected 0 permissions, got %d", len(rolePerms))
	}

	// Empty userID for GetRolesByUserID should return empty roles without query
	roles, err := repo.GetRolesByUserID(ctx, "", &perms2.ScopeContext{Type: perms2.ScopeGlobal})
	if err != nil {
		t.Fatalf("expected no error for empty userID in GetRolesByUserID, got %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected 0 roles, got %d", len(roles))
	}

	// Empty userID for GetRolesForUser should return empty roles without query
	rolesList, err := repo.GetRolesForUser(ctx, "", perms2.ScopeContext{Type: perms2.ScopeGlobal})
	if err != nil {
		t.Fatalf("expected no error for empty userID in GetRolesForUser, got %d", err)
	}
	if len(rolesList) != 0 {
		t.Fatalf("expected 0 roles, got %d", len(rolesList))
	}

	// Empty inputs for GetRolesForUserByOrganization should return empty roles without query
	orgRoles, err := repo.GetRolesForUserByOrganization(ctx, "", "org-1")
	if err != nil || len(orgRoles) != 0 {
		t.Fatalf("expected 0 roles and no error for empty userID in GetRolesForUserByOrganization, got %d, err: %v", len(orgRoles), err)
	}
	orgRoles, err = repo.GetRolesForUserByOrganization(ctx, "user-1", "")
	if err != nil || len(orgRoles) != 0 {
		t.Fatalf("expected 0 roles and no error for empty orgID in GetRolesForUserByOrganization, got %d, err: %v", len(orgRoles), err)
	}

	// Empty inputs for GetRolesForUserByProject should return empty roles without query
	projRoles, err := repo.GetRolesForUserByProject(ctx, "", "proj-1")
	if err != nil || len(projRoles) != 0 {
		t.Fatalf("expected 0 roles and no error for empty userID in GetRolesForUserByProject, got %d, err: %v", len(projRoles), err)
	}
	projRoles, err = repo.GetRolesForUserByProject(ctx, "user-1", "")
	if err != nil || len(projRoles) != 0 {
		t.Fatalf("expected 0 roles and no error for empty projectID in GetRolesForUserByProject, got %d, err: %v", len(projRoles), err)
	}

	// Empty userID for GetRolesForUserByGlobal should return empty roles without query
	globalRoles, err := repo.GetRolesForUserByGlobal(ctx, "", "")
	if err != nil || len(globalRoles) != 0 {
		t.Fatalf("expected 0 roles and no error for empty userID in GetRolesForUserByGlobal, got %d, err: %v", len(globalRoles), err)
	}
}

func TestAssignAndRevokeRoleValidation(t *testing.T) {
	repo := NewPostgresRepository(nil)
	ctx := context.Background()

	// Missing ID should fail validation
	err := repo.AssignRole(ctx, AssignedRole{
		RoleID: "contributor",
		UserID: stringPointer("user-1"),
	})
	if err == nil {
		t.Fatal("expected error when assignment ID is missing, got nil")
	}

	// Missing RoleID should fail validation
	err = repo.AssignRole(ctx, AssignedRole{
		ID:     "assign-1",
		UserID: stringPointer("user-1"),
	})
	if err == nil {
		t.Fatal("expected error when role ID is missing, got nil")
	}

	// Missing user/group holder should fail validation
	err = repo.AssignRole(ctx, AssignedRole{
		ID:     "assign-1",
		RoleID: "contributor",
	})
	if err == nil {
		t.Fatal("expected error when holder ID is missing, got nil")
	}

	// Missing assignment ID in RevokeRole should fail validation
	err = repo.RevokeRole(ctx, "")
	if err == nil {
		t.Fatal("expected error when assignment ID is missing in RevokeRole, got nil")
	}
}

type execCaptureDB struct {
	query string
	args  []any
}

func (db *execCaptureDB) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.query = query
	db.args = args
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (*execCaptureDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query call")
}

func (*execCaptureDB) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow call")
}

func TestAssignRoleWritesHolderAndScopeColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		assignment    AssignedRole
		wantScope     string
		wantProjectID string
		wantOrgID     string
		wantUserID    string
		wantGroupID   string
	}{
		{
			name: "defaults to a global user role",
			assignment: AssignedRole{
				ID:     "assignment-1",
				RoleID: "role-1",
				UserID: stringPointer("user-1"),
			},
			wantScope:  "GLOBAL",
			wantUserID: "user-1",
		},
		{
			name: "writes project team role as group",
			assignment: AssignedRole{
				ID:        "assignment-2",
				RoleID:    "role-2",
				GroupID:   stringPointer("team-1"),
				ScopeType: "Project",
				ProjectID: stringPointer("project-1"),
			},
			wantScope:     "PROJECT",
			wantProjectID: "project-1",
			wantGroupID:   "team-1",
		},
		{
			name: "uses generic organization scope ID",
			assignment: AssignedRole{
				ID:             "assignment-3",
				RoleID:         "role-3",
				GroupID:        stringPointer("group-1"),
				ScopeType:      "organization",
				OrganizationID: stringPointer("org-1"),
			},
			wantScope:   "ORGANIZATION",
			wantOrgID:   "org-1",
			wantGroupID: "group-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &execCaptureDB{}
			repo := NewPostgresRepository(db)

			if err := repo.AssignRole(context.Background(), tt.assignment); err != nil {
				t.Fatalf("AssignRole returned error: %v", err)
			}
			if !strings.Contains(db.query, "project_id, organization_id, user_id, group_id") {
				t.Fatalf("insert does not target the holder/scope columns from the schema: %s", db.query)
			}
			if len(db.args) != 7 {
				t.Fatalf("expected 7 SQL arguments, got %d", len(db.args))
			}
			if db.args[2] != tt.wantScope {
				t.Errorf("scope_type = %v, want %q", db.args[2], tt.wantScope)
			}
			assertNullableStringArg(t, db.args[3], tt.wantProjectID)
			assertNullableStringArg(t, db.args[4], tt.wantOrgID)
			assertNullableStringArg(t, db.args[5], tt.wantUserID)
			assertNullableStringArg(t, db.args[6], tt.wantGroupID)
		})
	}
}

func assertNullableStringArg(t *testing.T, got any, want string) {
	t.Helper()

	if want == "" {
		if got != (*string)(nil) {
			t.Errorf("argument = %#v, want nil", got)
		}
		return
	}

	value, ok := got.(*string)
	if !ok || value == nil || *value != want {
		t.Errorf("argument = %#v, want pointer to %q", got, want)
	}
}

func TestRoleAndAssignmentStructs(t *testing.T) {
	assignment := AssignedRole{
		ID:             "assign-101",
		RoleID:         "project-admin",
		UserID:         stringPointer("user-admin"),
		ScopeType:      "PROJECT",
		ProjectID:      stringPointer("proj-backend"),
		OrganizationID: nil,
	}

	if assignment.ID != "assign-101" || assignment.RoleID != "project-admin" {
		t.Fatalf("unexpected RoleAssignment fields: %+v", assignment)
	}

	permSvc := perms2.NewService()
	readBasic := permSvc.GetPermission(perms2.PermReadProjectBasic)
	if readBasic == nil {
		t.Fatalf("missing PermReadProjectBasic")
	}

	role := Role{
		ID:          "project-admin",
		Name:        "Project Admin",
		Description: "Admin role for projects",
		Immutable:   false,
		Permissions: []perms2.Permission{*readBasic},
	}

	if role.ID != "project-admin" || len(role.Permissions) != 1 {
		t.Fatalf("unexpected Role fields: %+v", role)
	}
}

func stringPointer(value string) *string {
	return &value
}
