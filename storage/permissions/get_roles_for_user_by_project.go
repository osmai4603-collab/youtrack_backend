package permstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// GetRolesForUserByProject retrieves roles assigned to a user at the project scope.
func (r *PostgresRepository) GetRolesForUserByProject(ctx context.Context, userID string, projectID string) ([]Role, error) {
	const op = "permissions.GetRolesForUserByProject"

	if userID == "" || projectID == "" {
		result := []Role{}
		requestlogger.LogStorageResult(ctx, op, result)
		return result, nil
	}

	scope := perms2.ScopeProject
	rows, err := r.db.Query(ctx, getQueryRolesForUser(&scope), userID, projectID)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	roles, err := scanRoles(rows, op)
	if err == nil {
		logRoles(ctx, op, roles)
	}
	return roles, err
}
