package app

import (
	"context"
	"strings"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/request"
	"youtrack_backend/channels/store"
)

const (
	PermissionIssueRead       = "issue.read"
	PermissionIssueWrite      = "issue.write"
	PermissionProjectRead     = "project.read"
	PermissionProjectWrite    = "project.write"
	PermissionProfileRead     = "profile.read"
	PermissionSearchRead      = "search.read"
	PermissionSavedQueryRead  = "saved_query.read"
	PermissionInboxRead       = "inbox.read"
	PermissionHubRead         = "hub.read"
	PermissionSecuritySearch  = "security_search.read"
	PermissionSubscriptionRead = "subscription.read"
	PermissionSystemConfig    = "system.config"
	PermissionSystemAdmin     = "system.admin"
)

var permissionRegistry = []string{
	PermissionIssueRead,
	PermissionIssueWrite,
	PermissionProjectRead,
	PermissionProjectWrite,
	PermissionProfileRead,
	PermissionSearchRead,
	PermissionSavedQueryRead,
	PermissionInboxRead,
	PermissionHubRead,
	PermissionSecuritySearch,
	PermissionSubscriptionRead,
	PermissionSystemConfig,
	PermissionSystemAdmin,
	"self.write",
}

var defaultRolePermissions = map[string][]string{
	"user": {
		PermissionIssueRead,
		PermissionProjectRead,
		PermissionProfileRead,
		PermissionSearchRead,
		PermissionSavedQueryRead,
		PermissionInboxRead,
		PermissionHubRead,
		PermissionSecuritySearch,
		PermissionSubscriptionRead,
		"self.write",
	},
	"project_admin": {
		PermissionIssueRead,
		PermissionIssueWrite,
		PermissionProjectRead,
		PermissionProjectWrite,
		PermissionProfileRead,
		PermissionSearchRead,
		PermissionSavedQueryRead,
		PermissionInboxRead,
		PermissionHubRead,
		PermissionSecuritySearch,
		PermissionSubscriptionRead,
	},
	"system_admin": {
		PermissionIssueRead,
		PermissionIssueWrite,
		PermissionProjectRead,
		PermissionProjectWrite,
		PermissionProfileRead,
		PermissionSearchRead,
		PermissionSavedQueryRead,
		PermissionInboxRead,
		PermissionHubRead,
		PermissionSecuritySearch,
		PermissionSubscriptionRead,
		PermissionSystemConfig,
		PermissionSystemAdmin,
	},
}

func IsKnownPermission(permission string) bool {
	for _, p := range permissionRegistry {
		if p == permission {
			return true
		}
	}
	return false
}

func RolePermissions(role string) []string {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		return nil
	}
	perms := append([]string(nil), defaultRolePermissions[role]...)
	return perms
}

type PermissionService struct{}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func (s *PermissionService) DefaultRoles() []*model.Role {
	return DefaultRoles()
}

func (s *PermissionService) DefaultPermissions() []*model.Permission {
	perms := make([]*model.Permission, 0, len(permissionRegistry))
	for _, name := range permissionRegistry {
		perms = append(perms, &model.Permission{
			ID:          name,
			Name:        name,
			Description: permissionDescription(name),
			Operation:   name,
			IsGlobal:    true,
			Type:        "Permission",
		})
	}
	return perms
}

func (s *PermissionService) RoleByName(name string) *model.Role {
	for _, role := range s.DefaultRoles() {
		if role != nil && strings.EqualFold(role.Name, name) {
			return role
		}
	}
	return nil
}

func (s *PermissionService) PermissionByName(name string) *model.Permission {
	for _, perm := range s.DefaultPermissions() {
		if perm != nil && strings.EqualFold(perm.Name, name) {
			return perm
		}
	}
	return nil
}

func (s *PermissionService) LoadRolesFromStore(ctx context.Context, admin store.AdminStore) ([]*model.Role, error) {
	if admin == nil {
		return s.DefaultRoles(), nil
	}
	roles, err := admin.Roles(ctx)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return s.DefaultRoles(), nil
	}
	return roles, nil
}

func (s *PermissionService) LoadPermissionsFromStore(ctx context.Context, admin store.AdminStore) ([]*model.Permission, error) {
	if admin == nil {
		return s.DefaultPermissions(), nil
	}
	perms, err := admin.Permissions(ctx)
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return s.DefaultPermissions(), nil
	}
	return perms, nil
}

func DefaultRoles() []*model.Role {
	roles := []*model.Role{}
	for roleName, perms := range defaultRolePermissions {
		roleModel := &model.Role{
			ID:          roleName,
			Name:        roleName,
			Description: roleName + " role",
			IsUpdatable: false,
			Immutable:   true,
			Type:        "Role",
			Permissions: make([]*model.Permission, 0, len(perms)),
		}
		for _, permName := range perms {
			roleModel.Permissions = append(roleModel.Permissions, &model.Permission{
				ID:          permName,
				Name:        permName,
				Description: permissionDescription(permName),
				Operation:   permName,
				IsGlobal:    true,
				Type:        "Permission",
			})
		}
		roles = append(roles, roleModel)
	}
	return roles
}

func permissionDescription(permission string) string {
	switch permission {
	case PermissionIssueRead:
		return "Read issues"
	case PermissionIssueWrite:
		return "Create and update issues"
	case PermissionProjectRead:
		return "Read projects"
	case PermissionProjectWrite:
		return "Manage projects"
	case PermissionProfileRead:
		return "Read user profiles"
	case PermissionSearchRead:
		return "Use search"
	case PermissionSavedQueryRead:
		return "Read saved queries"
	case PermissionInboxRead:
		return "Read inbox"
	case PermissionHubRead:
		return "Read hub services"
	case PermissionSecuritySearch:
		return "Access security search"
	case PermissionSubscriptionRead:
		return "Manage issue subscriptions"
	case PermissionSystemConfig:
		return "Manage system config"
	case PermissionSystemAdmin:
		return "Administrative access"
	case "self.write":
		return "Write self data"
	default:
		return permission
	}
}

func RolesGrantPermission(roles []string, permission string) bool {
	if !IsKnownPermission(permission) {
		return false
	}
	for _, role := range roles {
		role = strings.TrimSpace(strings.ToLower(role))
		if role == "" {
			continue
		}
		for _, p := range RolePermissions(role) {
			if p == permission {
				return true
			}
		}
	}
	return false
}

func (a *YouTrackApp) SessionHasPermission(c request.CTX, permission string) bool {
	if c == nil {
		return false
	}
	return RolesGrantPermission(c.UserRoles(), permission)
}

func (a *YouTrackApp) RequirePermission(c request.CTX, permission string) bool {
	return a.SessionHasPermission(c, permission)
}
