package permstorage

import (
	"context"

	perms2 "youtrack_backend/internal/services/permissions_services"
)

// GetRolesForUser returns roles assigned to a user within the specified scope context.
func (r *PostgresRepository) GetRolesForUser(ctx context.Context, userID string, scope perms2.ScopeContext) ([]Role, error) {
	const op = "permissions.GetRolesForUser"
	if userID == "" {
		result := []Role{}
		logRoles(ctx, op, result)
		return result, nil
	}

	var roles []Role
	var err error
	switch scope.Type {
	case perms2.ScopeOrganization:
		roles, err = r.GetRolesForUserByOrganization(ctx, userID, scope.ScopeID)
	case perms2.ScopeProject:
		roles, err = r.GetRolesForUserByProject(ctx, userID, scope.ScopeID)
	case perms2.ScopeGlobal:
		fallthrough
	default:
		roles, err = r.GetRolesForUserByGlobal(ctx, userID, scope.ScopeID)
	}
	return roles, err
}
