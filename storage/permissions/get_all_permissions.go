package permstorage

import (
	"context"

	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// GetAllPermissions returns the full permission catalog.
func (r *PostgresRepository) GetAllPermissions(ctx context.Context) ([]perms2.Permission, error) {
	const op = "permissions.GetAllPermissions"
	permissions := perms2.GetDefaultPermissions()

	requestlogger.LogStorageResult(ctx, op, permissions)
	return permissions, nil
}
