package permissionsusecase

import (
	"context"
	"errors"
	"testing"

	permstorage "youtrack_backend/internal/storage/permissions"
)

// stubRoleReader is a hand-written RoleReader. A stub keeps these tests on the
// use case's own mapping — scope fields, group attribution — rather than on the
// SQL, which needs a database to exercise.
type stubRoleReader struct {
	assignments []permstorage.AssignedRole
	roles       map[string]*permstorage.Role
	names       map[string][]string
	readErr     error
}

func (s *stubRoleReader) GetAssignedRoles(context.Context, string, string) ([]permstorage.AssignedRole, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	return s.assignments, nil
}

func (s *stubRoleReader) GetRoleByID(_ context.Context, roleID string) (*permstorage.Role, error) {
	if role, ok := s.roles[roleID]; ok {
		return role, nil
	}
	return nil, errors.New("no such role")
}

func (s *stubRoleReader) GetPermissionsByRoleIDs(_ []string, _ context.Context) ([]string, error) {
	return s.names["all"], nil
}

// -----------------------------------------------------------------------------
// The scope of every reported role
// -----------------------------------------------------------------------------

func TestGet_ReportsTheScopeOfEachRole(t *testing.T) {
	// The reason this endpoint exists: a client cannot gate a control on "a role
	// with some permissions". It has to know whether the role is global, spans an
	// organization, or is confined to one project.
	stub := &stubRoleReader{
		assignments: []permstorage.AssignedRole{
			{RoleID: "r-admin", UserID: stringPointer("user-1"), ScopeType: "GLOBAL"},
			{RoleID: "r-org", UserID: stringPointer("user-1"), ScopeType: "ORGANIZATION", OrganizationID: stringPointer("org-1")},
			{RoleID: "r-dev", UserID: stringPointer("user-1"), ScopeType: "PROJECT", ProjectID: stringPointer("proj-1")},
		},
		roles: map[string]*permstorage.Role{
			"r-admin": {ID: "r-admin", Name: "System Admin"},
			"r-org":   {ID: "r-org", Name: "Org Dev"},
			"r-dev":   {ID: "r-dev", Name: "Project Dev"},
		},
		names: map[string][]string{"all": {"READ_ISSUE"}},
	}

	result, err := NewUserPermissions(nil, stub).Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	want := map[string]struct {
		scope   string
		org     string
		project string
	}{
		"r-admin": {scope: "global"},
		"r-org":   {scope: "organization", org: "org-1"},
		"r-dev":   {scope: "project", project: "proj-1"},
	}

	for _, assignment := range result.RoleAssignments {
		expected, ok := want[assignment.RoleID]
		if !ok {
			t.Fatalf("unexpected role %q in the result", assignment.RoleID)
		}
		if assignment.ScopeType != expected.scope {
			t.Errorf("%s: scope = %q, want %q", assignment.RoleID, assignment.ScopeType, expected.scope)
		}
		if assignment.OrganizationID != expected.org {
			t.Errorf("%s: organization = %q, want %q", assignment.RoleID, assignment.OrganizationID, expected.org)
		}
		if assignment.ProjectID != expected.project {
			t.Errorf("%s: project = %q, want %q", assignment.RoleID, assignment.ProjectID, expected.project)
		}
	}
	if len(result.RoleAssignments) != len(want) {
		t.Fatalf("got %d assignments, want %d", len(result.RoleAssignments), len(want))
	}
}

func TestGet_TreatsAnEmptyScopeTypeAsGlobal(t *testing.T) {
	// Be defensive for rows produced by older data/import paths that have no
	// scope type, even though the current schema makes the column non-null.
	stub := &stubRoleReader{
		assignments: []permstorage.AssignedRole{{RoleID: "r-old", UserID: stringPointer("user-1")}},
		roles:       map[string]*permstorage.Role{"r-old": {ID: "r-old", Name: "Legacy"}},
	}

	result, err := NewUserPermissions(nil, stub).Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(result.RoleAssignments) != 1 {
		t.Fatalf("got %d assignments, want 1", len(result.RoleAssignments))
	}
	if result.RoleAssignments[0].ScopeType != "global" {
		t.Fatalf("scope = %q, want global", result.RoleAssignments[0].ScopeType)
	}
}

// -----------------------------------------------------------------------------
// Group membership
// -----------------------------------------------------------------------------

func TestGet_AttributesAGroupRoleToItsGroup(t *testing.T) {
	// The repository expands group membership, so a role arriving with
	// holder_type "group" was inherited. Naming the group is what lets a client
	// explain the access and an admin find where to change it.
	stub := &stubRoleReader{
		assignments: []permstorage.AssignedRole{
			{RoleID: "r-dev", GroupID: stringPointer("group-7"), ScopeType: "PROJECT", ProjectID: stringPointer("proj-1")},
		},
		roles: map[string]*permstorage.Role{"r-dev": {ID: "r-dev", Name: "Project Dev"}},
	}

	result, err := NewUserPermissions(nil, stub).Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result.RoleAssignments[0].GroupID != "group-7" {
		t.Fatalf("group = %q, want group-7", result.RoleAssignments[0].GroupID)
	}
}

func TestGet_LeavesTheGroupEmptyForADirectAssignment(t *testing.T) {
	stub := &stubRoleReader{
		assignments: []permstorage.AssignedRole{
			{RoleID: "r-admin", UserID: stringPointer("user-1"), ScopeType: "GLOBAL"},
		},
		roles: map[string]*permstorage.Role{"r-admin": {ID: "r-admin", Name: "System Admin"}},
	}

	result, err := NewUserPermissions(nil, stub).Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result.RoleAssignments[0].GroupID != "" {
		t.Fatalf("a direct assignment must report no group, got %q", result.RoleAssignments[0].GroupID)
	}
}

// -----------------------------------------------------------------------------
// Existing behaviour that must not change
// -----------------------------------------------------------------------------

func TestGet_RequiresAUserID(t *testing.T) {
	if _, err := NewUserPermissions(nil, &stubRoleReader{}).Get(context.Background(), ""); !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}
}

func TestGet_RejectsACancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stub := &stubRoleReader{readErr: errors.New("the repository must not be reached")}
	if _, err := NewUserPermissions(nil, stub).Get(ctx, "user-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestGet_WithoutARepository(t *testing.T) {
	if _, err := NewUserPermissions(nil, nil).Get(context.Background(), "user-1"); err == nil {
		t.Fatal("expected an error from a use case built with no repository")
	}
}

func TestGet_FallsBackToTheRoleIDWhenTheNameIsUnknown(t *testing.T) {
	// A role row that cannot be read must not lose the entry: the id is enough
	// for a client to render something and for an admin to look it up.
	stub := &stubRoleReader{
		assignments: []permstorage.AssignedRole{
			{RoleID: "r-gone", UserID: stringPointer("user-1"), ScopeType: "GLOBAL"},
		},
		roles: map[string]*permstorage.Role{},
	}

	result, err := NewUserPermissions(nil, stub).Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result.RoleAssignments[0].RoleName != "r-gone" {
		t.Fatalf("role name = %q, want the role id as a fallback", result.RoleAssignments[0].RoleName)
	}
}

func TestGet_AlwaysReturnsNonNilSlices(t *testing.T) {
	// An empty list has to marshal as [] rather than null, or every client
	// special-cases a user with no roles.
	stub := &stubRoleReader{assignments: nil, roles: map[string]*permstorage.Role{}}

	result, err := NewUserPermissions(nil, stub).Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result.RoleAssignments == nil || result.OwnedProjectIDs == nil {
		t.Fatal("expected empty slices, got nil")
	}
}

func stringPointer(value string) *string {
	return &value
}
