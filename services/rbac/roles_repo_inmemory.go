package rbac

import (
	"context"
	"fmt"
	"sync"
)

// inMemoryRBACRoleRepository is a thread-safe in-memory implementation
// of RBACRoleRepository. Used as the default when no external repository is provided.
type inMemoryRBACRoleRepository struct {
	mu    sync.RWMutex
	roles map[RoleID]*Role
}

func newInMemoryRBACRoleRepository() *inMemoryRBACRoleRepository {
	return &inMemoryRBACRoleRepository{
		roles: make(map[RoleID]*Role),
	}
}

func (r *inMemoryRBACRoleRepository) SaveRole(_ context.Context, role *Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[role.ID]; exists {
		return fmt.Errorf("%w: %s", ErrRoleAlreadyExists, role.ID)
	}
	r.roles[role.ID] = cloneRole(role)
	return nil
}

func (r *inMemoryRBACRoleRepository) GetRole(_ context.Context, id RoleID) (*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, exists := r.roles[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, id)
	}
	return cloneRole(role), nil
}

func (r *inMemoryRBACRoleRepository) GetAllRoles(_ context.Context) ([]*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Role, 0, len(r.roles))
	for _, role := range r.roles {
		result = append(result, cloneRole(role))
	}
	return result, nil
}

func (r *inMemoryRBACRoleRepository) UpdateRole(_ context.Context, role *Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[role.ID]; !exists {
		return fmt.Errorf("%w: %s", ErrRoleNotFound, role.ID)
	}
	r.roles[role.ID] = cloneRole(role)
	return nil
}

func (r *inMemoryRBACRoleRepository) DeleteRole(_ context.Context, id RoleID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[id]; !exists {
		return fmt.Errorf("%w: %s", ErrRoleNotFound, id)
	}
	delete(r.roles, id)
	return nil
}
