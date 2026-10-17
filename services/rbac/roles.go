package rbac

import (
	"context"
	"fmt"
	"sort"
)

// RBACRoleRepository abstracts persistent storage for RBAC role definitions.
// Implementations may use PostgreSQL, Redis, or any other backing store.
type RBACRoleRepository interface {
	// SaveRole persists a role (insert). Returns ErrRoleAlreadyExists if the role ID is taken.
	SaveRole(ctx context.Context, role *Role) error

	// GetRole retrieves a role by its ID. Returns ErrRoleNotFound if absent.
	GetRole(ctx context.Context, id RoleID) (*Role, error)

	// GetAllRoles returns all stored roles.
	GetAllRoles(ctx context.Context) ([]*Role, error)

	// UpdateRole replaces an existing role. Returns ErrRoleNotFound if absent.
	UpdateRole(ctx context.Context, role *Role) error

	// DeleteRole removes a role by its ID. Returns ErrRoleNotFound if absent.
	DeleteRole(ctx context.Context, id RoleID) error
}

// RoleGraph is a registry of independent role definitions backed by a repository.
type RoleGraph struct {
	repo RBACRoleRepository
}

// NewRoleGraph creates a role registry.
// If repo is nil, an in-memory store is used automatically.
func NewRoleGraph(repo RBACRoleRepository) *RoleGraph {
	if repo == nil {
		repo = newInMemoryRBACRoleRepository()
	}
	return &RoleGraph{repo: repo}
}

// AddRole inserts a new independent role into the registry.
func (g *RoleGraph) AddRole(role *Role) error {
	if role == nil || role.ID == "" {
		return fmt.Errorf("%w: role and role ID must not be empty", ErrInvalidInput)
	}

	ctx := context.Background()
	if err := g.repo.SaveRole(ctx, cloneRole(role)); err != nil {
		return err
	}
	return nil
}

// GetRole retrieves a copy of the specified role by ID.
func (g *RoleGraph) GetRole(id RoleID) (*Role, error) {
	ctx := context.Background()

	role, err := g.repo.GetRole(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, id)
	}
	return cloneRole(role), nil
}

// GetAllRoles returns copies of all registered roles sorted by ID.
func (g *RoleGraph) GetAllRoles() []*Role {
	ctx := context.Background()

	roles, err := g.repo.GetAllRoles(ctx)
	if err != nil {
		return nil
	}

	res := make([]*Role, 0, len(roles))
	for _, role := range roles {
		res = append(res, cloneRole(role))
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

// UpdateRole updates an existing role. Immutable roles cannot be altered.
func (g *RoleGraph) UpdateRole(updated *Role) error {
	if updated == nil || updated.ID == "" {
		return fmt.Errorf("%w: role and role ID must not be empty", ErrInvalidInput)
	}

	ctx := context.Background()

	existing, err := g.repo.GetRole(ctx, updated.ID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRoleNotFound, updated.ID)
	}
	if existing.IsImmutable {
		return fmt.Errorf("%w: cannot update immutable role %s", ErrRoleImmutable, updated.ID)
	}

	if err := g.repo.UpdateRole(ctx, cloneRole(updated)); err != nil {
		return fmt.Errorf("rbac: failed to persist role update: %w", err)
	}
	return nil
}

// DeleteRole removes a role from the registry. Immutable roles cannot be deleted.
func (g *RoleGraph) DeleteRole(id RoleID) error {
	ctx := context.Background()

	role, err := g.repo.GetRole(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRoleNotFound, id)
	}
	if role.IsImmutable {
		return fmt.Errorf("%w: cannot delete immutable role %s", ErrRoleImmutable, id)
	}

	if err := g.repo.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("rbac: failed to delete role: %w", err)
	}
	return nil
}

// GetEffectivePermissions returns only permissions explicitly granted to a role.
func (g *RoleGraph) GetEffectivePermissions(roleID RoleID) ([]string, error) {
	ctx := context.Background()

	role, err := g.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, roleID)
	}

	permissionSet := make(map[string]struct{}, len(role.Permissions))
	for _, permission := range role.Permissions {
		permissionSet[permission] = struct{}{}
	}
	permissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)
	return permissions, nil
}

// HasPermission checks whether a role has a permission explicitly granted to it.
func (g *RoleGraph) HasPermission(roleID RoleID, targetPermission string) (bool, error) {
	permissions, err := g.GetEffectivePermissions(roleID)
	if err != nil {
		return false, err
	}
	for _, permission := range permissions {
		if permission == targetPermission {
			return true, nil
		}
	}
	return false, nil
}

// Clone creates an isolated, deep copy of the role registry.
func (g *RoleGraph) Clone() *RoleGraph {
	allRoles := g.GetAllRoles()

	clone := NewRoleGraph(nil)
	for _, role := range allRoles {
		_ = clone.AddRole(role)
	}
	return clone
}

func cloneRole(role *Role) *Role {
	if role == nil {
		return nil
	}
	clone := *role
	if role.Permissions != nil {
		clone.Permissions = append([]string(nil), role.Permissions...)
	}
	if role.Metadata != nil {
		clone.Metadata = make(map[string]string, len(role.Metadata))
		for key, value := range role.Metadata {
			clone.Metadata[key] = value
		}
	}
	return &clone
}
