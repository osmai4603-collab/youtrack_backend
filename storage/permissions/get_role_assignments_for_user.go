package permstorage

import (
	"context"

	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"
)

func (r *PostgresRepository) GetRoleAssignmentsForUser(ctx context.Context, userID string) ([]rbac.RoleAssignment, error) {
	if userID == "" {
		return []rbac.RoleAssignment{}, nil
	}

	assignments, err := r.GetAssignedRoles(ctx, "USER", userID)
	if err != nil {
		return nil, err
	}

	permSvc := perms2.NewService()
	result := make([]rbac.RoleAssignment, 0, len(assignments))
	for _, assignment := range assignments {
		result = append(result, toRoleAssignment(assignment, permSvc))
	}

	return result, nil
}

func toRoleAssignment(assignment AssignedRole, permSvc perms2.IPermissionService) rbac.RoleAssignment {
	role := rbac.RoleAssignment{
		AssignmentID: assignment.ID,
		RoleID:       assignment.RoleID,
		RoleName:     assignment.RoleName,
		ScopeType:    assignment.ScopeType,
		Permissions:  assignment.Permissions,
	}
	if assignment.OrganizationID != nil {
		role.OrganizationID = *assignment.OrganizationID
	}
	if assignment.ProjectID != nil {
		role.ProjectID = *assignment.ProjectID
	}
	for _, ref := range assignment.Permissions {
		if permission := permSvc.GetPermission(ref); permission != nil {
			role.Permissions = append(role.Permissions, ref)
		} else {
			role.Permissions = append(role.Permissions, ref)
		}
	}
	return role
}
