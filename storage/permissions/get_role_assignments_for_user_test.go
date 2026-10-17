package permstorage

import (
	"testing"

	perms2 "youtrack_backend/internal/services/permissions_services"
)

func TestToRoleAssignmentPreservesScopeAndResolvesPermissions(t *testing.T) {
	orgID := "org-1"
	permSvc := perms2.NewService()
	assignment := toRoleAssignment(AssignedRole{
		ID:             "assignment-1",
		RoleID:         "role-1",
		RoleName:       "Reader",
		ScopeType:      string(perms2.ScopeOrganization),
		OrganizationID: &orgID,
		Permissions:    []string{perms2.PermReadProjectBasic, "CUSTOM_PERMISSION"},
	}, permSvc)

	if assignment.AssignmentID != "assignment-1" || assignment.RoleID != "role-1" || assignment.RoleName != "Reader" {
		t.Fatalf("role assignment identity was not preserved: %+v", assignment)
	}
	if assignment.ScopeType != string(perms2.ScopeOrganization) || assignment.OrganizationID != orgID {
		t.Fatalf("organization scope was not preserved: %+v", assignment)
	}
	if len(assignment.Permissions) != 2 || assignment.Permissions[0] != perms2.PermReadProjectBasic {
		t.Fatalf("catalog permission was not resolved: %+v", assignment.Permissions)
	}
	if assignment.Permissions[1] != "CUSTOM_PERMISSION" {
		t.Fatalf("unknown permission reference was not preserved: %+v", assignment.Permissions[1])
	}
}
