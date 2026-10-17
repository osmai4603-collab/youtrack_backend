package permstorage

import (
	"context"
	"fmt"
	"strings"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// AssignRole inserts a new role assignment into assigned_roles table.
func (r *PostgresRepository) AssignRole(ctx context.Context, assignment AssignedRole) error {
	const op = "permissions.AssignRole"

	if assignment.ID == "" {
		return fmt.Errorf("%s: assignment ID is required", op)
	}
	if assignment.RoleID == "" {
		return fmt.Errorf("%s: role ID is required", op)
	}
	if (assignment.UserID == nil) == (assignment.GroupID == nil) {
		return fmt.Errorf("%s: exactly one of user ID or group ID is required", op)
	}
	if (assignment.UserID != nil && *assignment.UserID == "") ||
		(assignment.GroupID != nil && *assignment.GroupID == "") {
		return fmt.Errorf("%s: holder ID cannot be empty", op)
	}

	scopeType := strings.ToLower(assignment.ScopeType)
	if scopeType == "" {
		scopeType = "global"
	}
	if scopeType != "global" && scopeType != "organization" && scopeType != "project" {
		return fmt.Errorf("%s: unsupported scope type %q", op, assignment.ScopeType)
	}

	switch scopeType {
	case "global":
		if assignment.ProjectID != nil || assignment.OrganizationID != nil {
			return fmt.Errorf("%s: global scope cannot have a scope ID", op)
		}
	case "project":
		if assignment.ProjectID == nil || *assignment.ProjectID == "" || assignment.OrganizationID != nil {
			return fmt.Errorf("%s: project scope requires only a project ID", op)
		}
	case "organization":
		if assignment.OrganizationID == nil || *assignment.OrganizationID == "" || assignment.ProjectID != nil {
			return fmt.Errorf("%s: organization scope requires only an organization ID", op)
		}
	}

	query := `
		INSERT INTO assigned_roles (
			id, role_id, scope_type, project_id, organization_id, user_id, group_id
		) VALUES (
			$1, $2, $3::access_scope_type, $4, $5, $6, $7
		)
	`

	_, err := r.db.Exec(ctx, query,
		assignment.ID,
		assignment.RoleID,
		strings.ToUpper(scopeType),
		assignment.ProjectID,
		assignment.OrganizationID,
		assignment.UserID,
		assignment.GroupID,
	)

	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op, map[string]any{"assigned": true})
	return nil
}
