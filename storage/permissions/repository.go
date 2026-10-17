package permstorage

import (
	"context"
	"time"

	"youtrack_backend/internal/infrastructure/postgres"
	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"

	"github.com/jackc/pgx/v5"
)

type AssignedRole struct {
	ID             string     `json:"id"`
	RoleID         string     `json:"role_id"`
	ScopeType      string     `json:"scope_type"`
	ProjectID      *string    `json:"project_id"`
	OrganizationID *string    `json:"organization_id"`
	UserID         *string    `json:"user_id"`
	GroupID        *string    `json:"group_id"`
	RoleName       string     `json:"role_name"`
	Permissions    []string   `json:"permissions,omitempty"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

type Role struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Immutable   bool                `json:"immutable"`
	Permissions []perms2.Permission `json:"permissions,omitempty"`
}

type Repository interface {
	GetPermissionsForUser(ctx context.Context, userID string, scope perms2.ScopeContext) ([]perms2.ScopeRole, error)
	// GetPermissionsByRoleIDs returns permission ids (or names) for given role IDs.
	GetPermissionsByRoleIDs(roleIDs []string, ctx context.Context) ([]string, error)
	GetRolesByUserID(ctx context.Context, userID string, scope *perms2.ScopeContext) ([]string, error)
	GetRoleIDsByUserForProjectScope(ctx context.Context, userID string, projectID string) ([]string, error)
	GetRoleIDsByGroupForProjectScope(ctx context.Context, groupID string, projectID string) ([]string, error)
	// GetRolesForUser returns all roles assigned to a user within the specified scope context.
	GetRolesForUser(ctx context.Context, userID string, scope perms2.ScopeContext) ([]Role, error)
	// GetRolesForUserByOrganization returns roles assigned to a user at the organization scope.
	GetRolesForUserByOrganization(ctx context.Context, userID string, orgID string) ([]Role, error)
	// GetRolesForUserByProject returns roles assigned to a user at the project scope.
	GetRolesForUserByProject(ctx context.Context, userID string, projectID string) ([]Role, error)
	// GetRolesForUserByGlobal returns roles assigned to a user at the global scope.
	GetRolesForUserByGlobal(ctx context.Context, userID string, scopeID string) ([]Role, error)
	// GetRoleAssignmentsForUser returns a user's direct and group-inherited role assignments with scope data.
	GetRoleAssignmentsForUser(ctx context.Context, userID string) ([]rbac.RoleAssignment, error)
	AssignRole(ctx context.Context, assignment AssignedRole) error
	RevokeRole(ctx context.Context, assignmentID string) error
	GetAssignedRoles(ctx context.Context, holderType string, holderID string) ([]AssignedRole, error)
	GetAllPermissions(ctx context.Context) ([]perms2.Permission, error)
	GetAllRoles(ctx context.Context) ([]Role, error)
	GetRoleByID(ctx context.Context, roleID string) (*Role, error)
	CreateRole(ctx context.Context, name string, description *string) (*Role, error)
	UpdateRole(ctx context.Context, roleID string, name, description *string) (*Role, error)
	DeleteRole(ctx context.Context, roleID string) error
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db postgres.DBTX
}

// NewPostgresRepository creates a new PostgresRepository instance.
func NewPostgresRepository(db postgres.DBTX) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func scanRoles(rows pgx.Rows, op string) ([]Role, error) {
	defer rows.Close()

	permSvc := perms2.NewService()

	var roles []Role
	for rows.Next() {
		var role Role
		var rawPerms []string
		var description *string
		if err := rows.Scan(&role.ID, &role.Name, &description, &role.Immutable, &rawPerms); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		if description != nil {
			role.Description = *description
		}
		for _, ref := range rawPerms {
			if p := permSvc.GetPermission(ref); p != nil {
				role.Permissions = append(role.Permissions, *p)
			} else {
				role.Permissions = append(role.Permissions, perms2.Permission{ID: ref, DisplayName: ref})
			}
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	if roles == nil {
		roles = []Role{}
	}
	return roles, nil
}
