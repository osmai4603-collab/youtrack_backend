package permstorage

import (
	"context"
	"fmt"
	"strconv"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// GetAllRoles retrieves all roles and their permissions.
func (r *PostgresRepository) GetAllRoles(ctx context.Context) ([]Role, error) {
	const op = "permissions.GetAllRoles"

	query := `
		SELECT id, name, description, immutable, permissions
		FROM roles
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	roles, err := scanRoles(rows, op)
	if err == nil {
		logRoles(ctx, op+"()", roles)
	}
	return roles, err
}

func logRoles(ctx context.Context, operation string, roles []Role) {
	if len(roles) == 0 {
		requestlogger.LogStorageResult(ctx, operation, "[]")
		return
	}

	for _, role := range roles {
		requestlogger.LogStorageResult(ctx, operation, fmt.Sprintf(
			"{id: %s, name: %s, permissions: %s}",
			role.ID, role.Name, strconv.Itoa(len(role.Permissions)),
		))
	}
}
