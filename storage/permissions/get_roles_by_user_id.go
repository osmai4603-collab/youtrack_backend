package permstorage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"

	"github.com/jackc/pgx/v5"
)

func getQueryRolesForUser(scope *perms2.ScopeLevel) string {
	query := userGroupResolutionCTE + `
		SELECT DISTINCT r.id, r.name, COALESCE(r.description, ''), r.immutable, r.permissions
		FROM roles r
		JOIN assigned_roles ar ON ar.role_id = r.id
		WHERE (ar.user_id = $1 OR ar.group_id IN (SELECT group_id FROM user_group_ids))
	`
	if scope != nil {
		switch *scope {
		case perms2.ScopeProject:
			return query + `
				AND ar.scope_type = 'PROJECT' AND ar.project_id = $2 AND ar.organization_id IS NULL
			`
		case perms2.ScopeOrganization:
			return query + `
				AND ar.scope_type = 'ORGANIZATION' AND ar.organization_id = $2 AND ar.project_id IS NULL
			`
		case perms2.ScopeGlobal:
			return query + `
				AND ar.scope_type = 'GLOBAL' AND ar.project_id IS NULL AND ar.organization_id IS NULL
			`
		}
	}
	return query
}

func getQueryRoleIDsForUser(scope *perms2.ScopeLevel) string {
	query := userGroupResolutionCTE + `
		SELECT DISTINCT ar.role_id
		FROM assigned_roles ar
		WHERE (ar.user_id = $1 OR ar.group_id IN (SELECT group_id FROM user_group_ids))
	`
	if scope != nil {
		switch *scope {
		case perms2.ScopeProject:
			return query + `
				AND ar.scope_type = 'PROJECT' AND ar.project_id = $2 AND ar.organization_id IS NULL
			`
		case perms2.ScopeOrganization:
			return query + `
				AND ar.scope_type = 'ORGANIZATION' AND ar.organization_id = $2 AND ar.project_id IS NULL
			`
		case perms2.ScopeGlobal:
			return query + `
				AND ar.scope_type = 'GLOBAL' AND ar.project_id IS NULL AND ar.organization_id IS NULL
			`
		}
	}
	return query
}

// GetRolesByUserID returns role IDs assigned to a user in the given scope.
func (r *PostgresRepository) GetRolesByUserID(ctx context.Context, userID string, scope *perms2.ScopeContext) ([]string, error) {
	const op = "permissions.GetRolesByUserID"

	if userID == "" {
		result := []string{}
		requestlogger.LogStorageResult(ctx, op, result)
		return result, nil
	}

	executeQuery := func() (pgx.Rows, error) {

		if scope.ScopeID != "" {
			rows, err := r.db.Query(ctx, getQueryRoleIDsForUser(&scope.Type), userID, scope.ScopeID)
			return rows, err
		}
		rows, err := r.db.Query(ctx, getQueryRoleIDsForUser(&scope.Type), userID)
		return rows, err
	}

	rows, err := executeQuery()
	if err != nil {
		return nil, postgres.TranslateError(op, fmt.Errorf("execute role lookup query: %w", err))
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, postgres.TranslateError(op, fmt.Errorf("scan role lookup result: %w", err))
		}
		roles = append(roles, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, fmt.Errorf("iterate role lookup results: %w", err))
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op+"(userID: "+userID+", scope: "+scopeLabel(scope)+")",
		slog.Any("role_ids", "["+strings.Join(roles, ", ")+"]"),
	)
	requestlogger.LogStorageResult(ctx, op+"(userID: "+userID+", scope: "+scopeLabel(scope)+")", roles)
	return roles, nil
}

func scopeLabel(scope *perms2.ScopeContext) string {
	if scope == nil {
		return "GLOBAL"
	}
	if scope.Type == "" {
		return string(perms2.ScopeGlobal)
	}
	return string(scope.Type)
}
