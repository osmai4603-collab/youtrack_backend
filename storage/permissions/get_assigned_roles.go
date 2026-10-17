package permstorage

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// GetAssignedRoles retrieves role assignments for a given holderType and holderID.
//
// When holderType is "user" the result also carries roles granted through
// direct or inherited group membership. Teams are groups with group_type = TEAM
// in this schema, so their roles are included through the same membership tree.
//
// The expansion is deliberately not applied to holderType "group": a group's own
// roles are what is being asked for, and widening a group lookup would pull in
// unrelated users' roles.
func getQueryAssignedRoles() string {
	return `
		WITH RECURSIVE
		membership_edge(member_type, member_id, group_id) AS (
			SELECT 'USER', gm.user_id, gm.group_id
			FROM group_members gm
			WHERE gm.member_type = 'USER' AND gm.user_id IS NOT NULL
			UNION
			SELECT 'GROUP', gm.member_group_id, gm.group_id
			FROM group_members gm
			WHERE gm.member_type = 'GROUP' AND gm.member_group_id IS NOT NULL
			UNION
			SELECT 'GROUP', g.id, g.parent_group_id
			FROM groups g
			WHERE g.parent_group_id IS NOT NULL AND g.group_type = 'TEAM'
		),
		user_group_ids(group_id) AS (
			SELECT e.group_id
			FROM membership_edge e
			WHERE e.member_type = 'USER' AND e.member_id = $1
			UNION
			SELECT e.group_id
			FROM user_group_ids ugi
			JOIN membership_edge e ON e.member_type = 'GROUP' AND e.member_id = ugi.group_id
		)
		SELECT ar.id, ar.role_id, ar.scope_type, ar.project_id, ar.organization_id,
		       ar.user_id, ar.group_id, ar.created_at, ar.updated_at, rl.name, rl.permissions
		FROM assigned_roles ar
		JOIN roles rl ON rl.id = ar.role_id
		WHERE (
			($2 = 'USER' AND (
				ar.user_id = $1 OR ar.group_id IN (SELECT group_id FROM user_group_ids)
			))
			OR ($2 = 'GROUP' AND ar.group_id = $1)
		)
		ORDER BY ar.role_id, LOWER(ar.scope_type::text)
	`
}

func (r *PostgresRepository) GetAssignedRoles(ctx context.Context, holderType string, holderID string) ([]AssignedRole, error) {
	const op = "GetAssignedRoles"

	query := getQueryAssignedRoles()

	rows, err := r.db.Query(ctx, query, holderID, holderType)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	var assignments []AssignedRole
	for rows.Next() {
		var a AssignedRole
		if err := rows.Scan(
			&a.ID, &a.RoleID, &a.ScopeType, &a.ProjectID, &a.OrganizationID,
			&a.UserID, &a.GroupID, &a.CreatedAt, &a.UpdatedAt, &a.RoleName, &a.Permissions,
		); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		assignments = append(assignments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	logAssignedRoles(ctx, op, holderType, holderID, assignments)

	return assignments, nil
}

func logAssignedRoles(ctx context.Context, operation string, holderType, holderID string, assignments []AssignedRole) {
	if len(assignments) == 0 {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, operation,
			slog.String("holderTP", holderType),
			slog.String("holderID", holderID),
			slog.String("roles", "[]"),
		)
		return
	}
	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, operation,
		slog.String("holderTP", holderType),
		slog.String("holderID", holderID),
	)
	for index, assign := range assignments {
		var scopeID string
		if assign.OrganizationID != nil {
			scopeID = *assign.OrganizationID
		} else if assign.ProjectID != nil {
			scopeID = *assign.ProjectID
		}
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "  "+strconv.Itoa((index+1)),
			slog.String("scope", assign.ScopeType),
			slog.String("scopeID", scopeID),
			slog.String("permissions", strconv.Itoa(len(assign.Permissions))),
			slog.String("role", assign.RoleName),
		)
	}
}

func conv(a AssignedRole) string {
	var scopeID string
	if a.OrganizationID != nil {
		scopeID = *a.OrganizationID
	} else if a.ProjectID != nil {
		scopeID = *a.ProjectID
	}
	return fmt.Sprintf("scope: %s, scopeID: %s, permissions: %s, role: %s",
		a.ScopeType, scopeID, strconv.Itoa(len(a.Permissions)), a.RoleName)
}
