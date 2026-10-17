package permstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// GetPermissionsByRoleIDs returns permission names or objects for given role IDs.
func (r *PostgresRepository) GetPermissionsByRoleIDs(roleIDs []string, ctx context.Context) ([]string, error) {
	const op = "permissions.GetPermissionsByRoleIDs"

	if len(roleIDs) == 0 {
		result := []string{}
		requestlogger.LogStorageResult(ctx, op, result)
		return result, nil
	}

	query := `
		SELECT DISTINCT perm
		FROM roles r
		CROSS JOIN LATERAL unnest(r.permissions) AS perm
		WHERE r.id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, roleIDs)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var ref string
		if err := rows.Scan(&ref); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		if ref != "" {
			permissions = append(permissions, ref)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	requestlogger.LogStorageResult(ctx, op, permissions)
	return permissions, nil
}
