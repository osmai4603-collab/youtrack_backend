package permstorage

import (
	"context"
	"log/slog"
	"strings"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

func getQueryPermissionsForUser(scope perms2.ScopeLevel) string {
	query := userGroupResolutionCTE + `
		SELECT ar.scope_type, rl.permissions, rl.name
		FROM assigned_roles ar
		JOIN roles rl ON rl.id = ar.role_id
		WHERE (ar.user_id = $1 OR ar.group_id IN (SELECT group_id FROM user_group_ids))
	`
	if scope == perms2.ScopeProject {
		return query + `
			AND ar.scope_type = 'PROJECT' AND ar.project_id = $2
		`
	}
	if scope == perms2.ScopeOrganization {
		return query + `
			AND ar.scope_type = 'ORGANIZATION' AND ar.organization_id = $2
		`
	}
	if scope == perms2.ScopeGlobal {
		return query + `
			AND ar.scope_type = 'GLOBAL' AND ar.organization_id IS NULL AND ar.project_id IS NULL
		`
	}
	return query
}

func (r *PostgresRepository) GetPermissionsForUser(ctx context.Context, userID string, scope perms2.ScopeContext) ([]perms2.ScopeRole, error) {
	const op = "permissions.GetPermissionsForUser"

	if userID == "" {
		result := []perms2.ScopeRole{}
		requestlogger.LogStorageResult(ctx, op, result)
		return result, nil
	}

	scopeType := scope.Type
	if scopeType == "" {
		scopeType = perms2.ScopeGlobal
	}

	args := []any{userID}
	if scopeType == perms2.ScopeProject || scopeType == perms2.ScopeOrganization {
		args = append(args, scope.ScopeID)
	}

	rows, err := r.db.Query(ctx, getQueryPermissionsForUser(scopeType), args...)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	databasePermissions := make([]string, 0)
	roles := make([]perms2.ScopeRole, 0)
	permSvc := perms2.NewService()

	for rows.Next() {
		var roleScope string
		var refs []string
		var roleName string
		if err := rows.Scan(&roleScope, &refs, &roleName); err != nil {
			return nil, postgres.TranslateError(op, err)
		}

		rolePermissions := make([]perms2.Permission, 0, len(refs))
		seen := make(map[string]bool)
		for _, ref := range refs {
			databasePermissions = append(databasePermissions, ref)
			if ref == "" {
				continue
			}

			permission := perms2.Permission{ID: ref, DisplayName: ref}
			if p := permSvc.GetPermission(ref); p != nil {
				permission = *p
			}

			if permission.Scope == perms2.ScopeGlobal && roleScope != string(perms2.ScopeGlobal) {
				continue
			}

			if permission.ID != "" && seen[permission.ID] || permission.DisplayName != "" && seen[permission.DisplayName] {
				continue
			}
			if permission.ID != "" {
				seen[permission.ID] = true
			}
			if permission.DisplayName != "" {
				seen[permission.DisplayName] = true
			}
			rolePermissions = append(rolePermissions, permission)
		}

		roles = append(roles, perms2.ScopeRole{
			ScopeType:   perms2.ScopeLevel(roleScope),
			RoleName:    roleName,
			Permissions: rolePermissions,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op+"(userID: "+userID+", scope: "+string(scopeType)+", scopeID: "+scope.ScopeID+")",
		slog.Any("permission_roles", "["+strings.Join(extractRoleNames(roles), ", ")+"]"),
	)

	return roles, nil
}

func extractRoleNames(roles []perms2.ScopeRole) []string {
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.RoleName)
	}
	return roleNames
}
