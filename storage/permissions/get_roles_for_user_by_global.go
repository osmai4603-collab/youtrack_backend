package permstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// GetRolesForUserByGlobal retrieves roles assigned to a user at the global scope.
func (r *PostgresRepository) GetRolesForUserByGlobal(ctx context.Context, userID string, _ string) ([]Role, error) {
	const op = "permissions.GetRolesForUserByGlobal"

	if userID == "" {
		result := []Role{}
		requestlogger.LogStorageResult(ctx, op, result)
		return result, nil
	}

	scope := perms2.ScopeGlobal

	rows, err := r.db.Query(ctx, getQueryRolesForUser(&scope), userID)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	roles, err := scanRoles(rows, op)
	if err == nil {
		logRoles(ctx, op, roles)
	}
	return roles, err
}
