package permstorage

import (
	"context"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const queryRoleIDsForGroupInProject = `
	SELECT DISTINCT ar.role_id
	FROM assigned_roles ar
	WHERE ar.group_id = $1
	  AND ar.scope_type = 'PROJECT'
	  AND ar.project_id = $2
	  AND ar.organization_id IS NULL
	ORDER BY ar.role_id
`
const queryRoleIDsForUserInProject = `
	SELECT DISTINCT ar.role_id
	FROM assigned_roles ar
	WHERE ar.user_id = $1
	  AND ar.scope_type = 'PROJECT'
	  AND ar.project_id = $2
	  AND ar.organization_id IS NULL
	ORDER BY ar.role_id
`

// GetRoleIDsByUserForProjectScope returns project-scoped roles granted
// directly to a user or inherited through their groups.
func (r *PostgresRepository) GetRoleIDsByUserForProjectScope(ctx context.Context, userID, projectID string) ([]string, error) {
	const op = "permissions.GetRoleIDsByUserForProjectScope"
	if userID == "" || projectID == "" {
		return []string{}, nil
	}

	rows, err := r.db.Query(ctx, queryRoleIDsForUserInProject, userID, projectID)
	if err != nil {
		return nil, postgres.TranslateError(op, fmt.Errorf("execute role lookup query: %w", err))
	}
	defer rows.Close()

	roleIDs := make([]string, 0)
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, postgres.TranslateError(op, fmt.Errorf("scan role lookup result: %w", err))
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, fmt.Errorf("iterate role lookup results: %w", err))
	}
	return roleIDs, nil
}

// GetRoleIDsByGroupForProjectScope returns roles granted directly to a group
// at the requested project scope.
func (r *PostgresRepository) GetRoleIDsByGroupForProjectScope(ctx context.Context, groupID, projectID string) ([]string, error) {
	const op = "permissions.GetRoleIDsByGroupForProjectScope"
	if groupID == "" || projectID == "" {
		return []string{}, nil
	}

	rows, err := r.db.Query(ctx, queryRoleIDsForGroupInProject, groupID, projectID)
	if err != nil {
		return nil, postgres.TranslateError(op, fmt.Errorf("execute role lookup query: %w", err))
	}
	defer rows.Close()

	roleIDs := make([]string, 0)
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, postgres.TranslateError(op, fmt.Errorf("scan role lookup result: %w", err))
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, fmt.Errorf("iterate role lookup results: %w", err))
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
		slog.String("group_id", groupID),
		slog.String("project_id", projectID),
		slog.Any("role_ids", roleIDs),
	)
	return roleIDs, nil
}

var _ interface {
	GetRoleIDsByUserForProjectScope(context.Context, string, string) ([]string, error)
	GetRoleIDsByGroupForProjectScope(context.Context, string, string) ([]string, error)
} = (*PostgresRepository)(nil)
