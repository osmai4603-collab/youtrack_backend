package permissionsusecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	perms2 "youtrack_backend/internal/services/permissions_services"
	permstorage "youtrack_backend/internal/storage/permissions"
)

var ErrUserIDRequired = errors.New("user id is required")

// RoleReader is the data access contract needed to expand a user's role
// assignments into the permission names each role grants. It is satisfied
// structurally by storage/permissions.Repository.
type RoleReader interface {
	GetAssignedRoles(ctx context.Context, holderType string, holderID string) ([]permstorage.AssignedRole, error)
	GetRoleByID(ctx context.Context, roleID string) (*permstorage.Role, error)
	GetPermissionsByRoleIDs(roleIDs []string, ctx context.Context) ([]string, error)
}

// UserRoleAssignment is a single role held by a user together with the
// permission names it grants in its own scope.
//
// ScopeType is what makes the entry meaningful to a client: without it, a
// system-admin role and a project-admin role are both "a role with some
// permissions", and a client cannot tell whether it may edit a project or every
// project. OrganizationID is carried alongside ScopeType because an
// organization role reaches every project that organization owns, which is not
// something ProjectID alone expresses.
type UserRoleAssignment struct {
	RoleID         string
	RoleName       string
	ScopeType      string
	OrganizationID string
	ProjectID      string
	GroupID        string
	UserID         string
	HolderType     string
	Permissions    []string
}

// UserPermissions is the full permission picture a client needs to gate its UI.
//
// OwnedProjectIDs is always empty. The projects table has no ownership relation
// — projects.leader_id names a project leader, which is not the same thing — so
// there is no query that could answer it. The field is kept because it is part
// of the published response, and filling it needs a schema decision rather than
// a guess. See §11.7 of docs/authorize_three_scopes_implementation_plan.md.
type UserPermissions struct {
	RoleAssignments []UserRoleAssignment
	OwnedProjectIDs []string
}

// UserPermissionsUseCase resolves the effective role assignments of a user.
type UserPermissionsUseCase struct {
	roles  RoleReader
	logger *slog.Logger
}

// NewUserPermissions builds the user-permissions use case.
func NewUserPermissions(logger *slog.Logger, roles RoleReader) *UserPermissionsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserPermissionsUseCase{roles: roles, logger: logger}
}

// Get resolves every role the user holds, directly or through a group, and the
// scope each one applies in. Assignments are reported per (role, scope) so a user
// holding the same role in two projects gets one entry per scope.
func (uc *UserPermissionsUseCase) Get(ctx context.Context, userID string) (*UserPermissions, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.roles == nil {
		return nil, errors.New("permission repository is not configured")
	}

	// The repository expands group membership for holderType "user", so this
	// covers inherited roles as well as direct ones.
	assignments, err := uc.roles.GetAssignedRoles(ctx, "USER", userID)
	if err != nil {
		return nil, fmt.Errorf("fetch assigned roles: %w", err)
	}

	result := &UserPermissions{
		RoleAssignments: []UserRoleAssignment{},
		OwnedProjectIDs: []string{},
	}
	for _, assignment := range assignments {
		roleName := assignment.RoleName
		if roleName == "" {
			if uc.roles != nil {
				if r, err := uc.roles.GetRoleByID(ctx, assignment.RoleID); err == nil && r != nil && r.Name != "" {
					roleName = r.Name
				}
			}
			if roleName == "" {
				roleName = assignment.RoleID
			}
		}
		holderID := ""
		if assignment.GroupID != nil {
			holderID = *assignment.GroupID
		}
		if assignment.UserID != nil {
			holderID = *assignment.UserID
		}
		organizationID := ""
		if assignment.OrganizationID != nil {
			organizationID = *assignment.OrganizationID
		}
		projectID := ""
		if assignment.ProjectID != nil {
			projectID = *assignment.ProjectID
		}
		groupID := ""
		if assignment.GroupID != nil {
			groupID = *assignment.GroupID
		}

		// scope_type is nullable in the schema, so an empty string here means the
		// row predates the column. It is reported as the global scope, which is
		// how the authorization query already reads it.
		scopeType := strings.ToLower(assignment.ScopeType)
		if scopeType == "" {
			scopeType = strings.ToLower(string(perms2.ScopeGlobal))
		}

		result.RoleAssignments = append(result.RoleAssignments, UserRoleAssignment{
			RoleID:         assignment.RoleID,
			RoleName:       roleName,
			ScopeType:      scopeType,
			OrganizationID: organizationID,
			ProjectID:      projectID,
			GroupID:        groupID,
			UserID:         holderID,
			HolderType:     "USER",
			Permissions:    assignment.Permissions,
		})
	}

	return result, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
