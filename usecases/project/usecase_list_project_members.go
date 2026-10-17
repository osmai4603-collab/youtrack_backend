package projectusecase

import (
	"context"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"
)

type ProjectMembersReader interface {
	GetListTeamMembersForProject(ctx context.Context, projectID string) ([]*domain.TeamMember, error)
}

type UsersFinder interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

type GroupsFinder interface {
	Get(ctx context.Context, id string) (*domain.Group, error)
}

type MemberRolesFinder interface {
	GetRoleIDsByUserForProjectScope(ctx context.Context, userID string, projectID string) ([]string, error)
	GetRoleIDsByGroupForProjectScope(ctx context.Context, groupID string, projectID string) ([]string, error)
}

type ListProjectMembersUseCase struct {
	members ProjectMembersReader
	groups  GroupsFinder
	users   UsersFinder
	roles   MemberRolesFinder
	rbacSvc rbac.IRBACService
	logger  *slog.Logger
}

func NewListProjectMembers(
	logger *slog.Logger,
	members ProjectMembersReader,
	groups GroupsFinder,
	users UsersFinder,
	roles MemberRolesFinder,
	rbac rbac.IRBACService,
) *ListProjectMembersUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ListProjectMembersUseCase{
		members: members,
		groups:  groups,
		users:   users,
		roles:   roles,
		logger:  logger,
		rbacSvc: rbac,
	}
}

type ListProjectMembersParams struct {
	UserID    string
	ProjectID string
}

type ListProjectMembersResult struct {
	ProjectID string
	Members   []*domain.TeamMember
}

func (uc *ListProjectMembersUseCase) Execute(ctx context.Context, params *ListProjectMembersParams) (*ListProjectMembersResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if params.ProjectID == "" {
		return nil, ErrProjectIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.members == nil {
		return nil, ErrRepositoryNotConfigured
	}

	access := accessRequest(params.ProjectID, params.UserID)
	if err := uc.rbacSvc.Authorize(ctx, access); err != nil {
		return nil, err
	}

	members, err := uc.members.GetListTeamMembersForProject(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}

	if members == nil {
		members = []*domain.TeamMember{}
	}

	for _, mr := range members {
		if mr.Roles == nil {
			mr.Roles = []string{}
		}
		groupID := mr.GroupID
		if groupID == "" {
			groupID = mr.MemberGroupID
		}
		if groupID != "" {
			if uc.groups != nil {
				group, err := uc.groups.Get(ctx, groupID)
				if err != nil {
					return nil, fmt.Errorf("list project members: get group %q: %w", groupID, err)
				}
				mr.Group = group
			}
			if uc.roles != nil {
				roles, err := uc.roles.GetRoleIDsByGroupForProjectScope(ctx, groupID, mr.ProjectID)
				if err != nil {
					return nil, fmt.Errorf("list project members: get roles for group %q: %w", groupID, err)
				}
				mr.Roles = roles
			}
		}
		if mr.UserID != "" {
			if uc.users != nil {
				user, err := uc.users.FindByID(ctx, mr.UserID)
				if err != nil {
					return nil, fmt.Errorf("list project members: get user %q: %w", mr.UserID, err)
				}
				mr.User = user
			}
			if uc.roles != nil {
				roles, err := uc.roles.GetRoleIDsByUserForProjectScope(ctx, mr.UserID, mr.ProjectID)
				if err != nil {
					return nil, fmt.Errorf("list project members: get roles for user %q: %w", mr.UserID, err)
				}
				mr.Roles = roles
			}
		}
	}

	return &ListProjectMembersResult{ProjectID: params.ProjectID, Members: members}, nil
}

func accessRequest(projectID string, userID string) rbac.AccessRequest {
	return rbac.AccessRequest{
		SubjectID:  userID,
		Resource:   string(perms2.EntityProject),
		ResourceID: projectID,
		Permission: perms2.PermReadProjectFull,
		ScopeType:  string(perms2.ScopeProject),
	}
}
