package groups

import (
	"context"
	"errors"
	"strings"

	"youtrack_backend/internal/domain"
	platformerr "youtrack_backend/internal/platform/errors"
	groupstorage "youtrack_backend/internal/storage/groups"
)

const maxGroupNameLength = 255

type UseCase struct {
	repository groupstorage.Repository
}

func New(repository groupstorage.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) List(ctx context.Context) ([]domain.Group, error) {
	if uc.repository == nil {
		return nil, platformerr.Internal("groups.List", "group repository is not configured", nil)
	}
	return uc.repository.List(ctx)
}

func (uc *UseCase) Create(ctx context.Context, name, description *string, groupType string, parentGroupID *string, autoJoin bool) (*domain.Group, error) {
	if uc.repository == nil {
		return nil, platformerr.Internal("groups.Create", "group repository is not configured", nil)
	}
	if name == nil || strings.TrimSpace(*name) == "" {
		return nil, platformerr.Invalid("groups.Create", "group name is required", errors.New("empty group name"))
	}
	normalizedName := strings.TrimSpace(*name)
	if len(normalizedName) > maxGroupNameLength {
		return nil, platformerr.Invalid("groups.Create", "group name is too long", errors.New("group name exceeds 255 characters"))
	}
	groupType = strings.ToUpper(strings.TrimSpace(groupType))
	if groupType == "" {
		groupType = "GROUP"
	}
	if groupType != "GROUP" && groupType != "TEAM" {
		return nil, platformerr.Invalid("groups.Create", "unsupported group type", errors.New("group type must be GROUP or TEAM"))
	}
	if description != nil {
		normalized := strings.TrimSpace(*description)
		description = &normalized
	}
	if parentGroupID != nil {
		normalized := strings.TrimSpace(*parentGroupID)
		if normalized == "" {
			parentGroupID = nil
		} else {
			parentGroupID = &normalized
		}
	}
	return uc.repository.Create(ctx, &domain.Group{
		Name:          normalizedName,
		Description:   description,
		GroupType:     groupType,
		ParentGroupID: parentGroupID,
		AutoJoin:      autoJoin,
	})
}

func (uc *UseCase) Get(ctx context.Context, groupID string) (*domain.Group, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, platformerr.Invalid("groups.Get", "group ID is required", errors.New("empty group ID"))
	}
	if uc.repository == nil {
		return nil, platformerr.Internal("groups.Get", "group repository is not configured", nil)
	}
	return uc.repository.Get(ctx, groupID)
}

func (uc *UseCase) Update(ctx context.Context, groupID string, name, description *string) (*domain.Group, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, platformerr.Invalid("groups.Update", "group ID is required", errors.New("empty group ID"))
	}
	if uc.repository == nil {
		return nil, platformerr.Internal("groups.Update", "group repository is not configured", nil)
	}
	if name != nil {
		normalized := strings.TrimSpace(*name)
		if normalized == "" {
			return nil, platformerr.Invalid("groups.Update", "group name is required", errors.New("empty group name"))
		}
		if len(normalized) > maxGroupNameLength {
			return nil, platformerr.Invalid("groups.Update", "group name is too long", errors.New("group name exceeds 255 characters"))
		}
		name = &normalized
	}
	return uc.repository.Update(ctx, groupID, name, description)
}

func (uc *UseCase) Delete(ctx context.Context, groupID, successorID string) error {
	groupID = strings.TrimSpace(groupID)
	successorID = strings.TrimSpace(successorID)
	if groupID == "" || successorID == "" {
		return platformerr.Invalid("groups.Delete", "group ID and successor are required", errors.New("missing group ID or successor"))
	}
	if groupID == successorID {
		return platformerr.Invalid("groups.Delete", "successor must be a different group", groupstorage.ErrSuccessorNeeded)
	}
	if uc.repository == nil {
		return platformerr.Internal("groups.Delete", "group repository is not configured", nil)
	}
	return uc.repository.Delete(ctx, groupID, successorID)
}
