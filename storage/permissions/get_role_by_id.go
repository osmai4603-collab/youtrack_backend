package permstorage

import (
	"context"
	"fmt"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// GetRoleByID retrieves a single role by role ID.
func (r *PostgresRepository) GetRoleByID(ctx context.Context, roleID string) (*Role, error) {
	const op = "permissions.GetRoleByID"

	query := `
		SELECT id, name, description, immutable, permissions
		FROM roles
		WHERE id = $1
	`

	var role Role
	var description *string
	var rawPerms []string
	err := r.db.QueryRow(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &description, &role.Immutable, &rawPerms,
	)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	if description != nil {
		role.Description = *description
	}

	permSvc := perms2.NewService()
	for _, ref := range rawPerms {
		if p := permSvc.GetPermission(ref); p != nil {
			role.Permissions = append(role.Permissions, *p)
		} else {
			role.Permissions = append(role.Permissions, perms2.Permission{ID: ref, DisplayName: ref})
		}
	}

	requestlogger.LogStorageResult(ctx, op+"(roleID: "+roleID+")", "{name: "+role.Name+", permissions: "+fmt.Sprint(len(role.Permissions))+"}")
	return &role, nil
}
