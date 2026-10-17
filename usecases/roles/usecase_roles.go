package roles

import (
	"context"
	"errors"
	"strings"

	platformerr "youtrack_backend/internal/platform/errors"
	permstorage "youtrack_backend/internal/storage/permissions"
)

const maxRoleNameLength = 255

type Repository interface {
	GetAllRoles(ctx context.Context) ([]permstorage.Role, error)
	GetRoleByID(ctx context.Context, roleID string) (*permstorage.Role, error)
	CreateRole(ctx context.Context, name string, description *string) (*permstorage.Role, error)
	UpdateRole(ctx context.Context, roleID string, name, description *string) (*permstorage.Role, error)
	DeleteRole(ctx context.Context, roleID string) error
}

type UseCase struct {
	repository Repository
}

func New(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) List(ctx context.Context) ([]permstorage.Role, error) {
	if uc.repository == nil {
		return nil, platformerr.Internal("roles.List", "role repository is not configured", nil)
	}
	return uc.repository.GetAllRoles(ctx)
}

func (uc *UseCase) Get(ctx context.Context, roleID string) (*permstorage.Role, error) {
	if strings.TrimSpace(roleID) == "" {
		return nil, platformerr.Invalid("roles.Get", "role ID is required", errors.New("empty role ID"))
	}
	if uc.repository == nil {
		return nil, platformerr.Internal("roles.Get", "role repository is not configured", nil)
	}
	return uc.repository.GetRoleByID(ctx, strings.TrimSpace(roleID))
}

func (uc *UseCase) Create(ctx context.Context, name string, description *string) (*permstorage.Role, error) {
	if uc.repository == nil {
		return nil, platformerr.Internal("roles.Create", "role repository is not configured", nil)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, platformerr.Invalid("roles.Create", "role name is required", errors.New("empty role name"))
	}
	if len(name) > maxRoleNameLength {
		return nil, platformerr.Invalid("roles.Create", "role name is too long", errors.New("role name exceeds 255 characters"))
	}
	if description != nil {
		normalized := strings.TrimSpace(*description)
		description = &normalized
	}
	return uc.repository.CreateRole(ctx, name, description)
}

func (uc *UseCase) Update(ctx context.Context, roleID string, name, description *string) (*permstorage.Role, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, platformerr.Invalid("roles.Update", "role ID is required", errors.New("empty role ID"))
	}
	if uc.repository == nil {
		return nil, platformerr.Internal("roles.Update", "role repository is not configured", nil)
	}
	if name != nil {
		normalized := strings.TrimSpace(*name)
		if normalized == "" {
			return nil, platformerr.Invalid("roles.Update", "role name is required", errors.New("empty role name"))
		}
		if len(normalized) > maxRoleNameLength {
			return nil, platformerr.Invalid("roles.Update", "role name is too long", errors.New("role name exceeds 255 characters"))
		}
		name = &normalized
	}
	return uc.repository.UpdateRole(ctx, roleID, name, description)
}

func (uc *UseCase) Delete(ctx context.Context, roleID string) error {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return platformerr.Invalid("roles.Delete", "role ID is required", errors.New("empty role ID"))
	}
	if uc.repository == nil {
		return platformerr.Internal("roles.Delete", "role repository is not configured", nil)
	}
	return uc.repository.DeleteRole(ctx, roleID)
}
