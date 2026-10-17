package permstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// GetRolesForUserByOrganization retrieves roles assigned to a user at the organization scope.
func (r *PostgresRepository) GetRolesForUserByOrganization(ctx context.Context, userID string, orgID string) ([]Role, error) {
	const op = "permissions.GetRolesForUserByOrganization"

	if userID == "" || orgID == "" {
		result := []Role{}
		requestlogger.LogStorageResult(ctx, op, result)
		return result, nil
	}

	scope := perms2.ScopeOrganization
	rows, err := r.db.Query(ctx, getQueryRolesForUser(&scope), userID, orgID)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	roles, err := scanRoles(rows, op)
	if err == nil {
		logRoles(ctx, op, roles)
	}
	return roles, err
}
