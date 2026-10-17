package projectusecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"
)

var (
	// ErrUserIDRequired is returned when the request carries no identity.
	ErrUserIDRequired = domain.ErrUserIDRequired
	// ErrRepositoryNotConfigured is returned when the use case was built without a project repository.
	ErrRepositoryNotConfigured = errors.New("project repository is not configured")
)

// ProjectReader is the data access contract needed to list the projects an identity may read.
type ProjectReader interface {
	GetProjectByUserID(ctx context.Context, userID string, includeArchived bool) ([]*domain.Project, error)
	ListAll(ctx context.Context, includeArchived bool) ([]*domain.Project, error)
	ListPublic(ctx context.Context, includeArchived bool) ([]*domain.Project, error)
}

// ListProjectsUseCase resolves the projects the authenticated user is allowed to read
// based on their RBAC authorization level:
// 1. Global READ_PROJECT_BASIC -> retrieves all projects (ListAll).
// 2. Project-scoped roles/permissions -> retrieves projects user is member of, leader of, or public projects (ListVisibleToUser).
// 3. Normal user (no project roles or global permission) -> retrieves public projects only (ListPublic).
type ListProjectsUseCase struct {
	projects   ProjectReader
	logger     *slog.Logger
	authorizer rbac.IRBACService
}

// NewListProjects builds the list-projects use case.
func NewListProjects(logger *slog.Logger, projects ProjectReader, authorizer ...rbac.IRBACService) *ListProjectsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	uc := &ListProjectsUseCase{
		projects: projects,
		logger:   logger,
	}
	if len(authorizer) > 0 {
		uc.authorizer = authorizer[0]
	}
	return uc
}

// SetAuthorizer binds the RBAC authorization service to the usecase.
func (uc *ListProjectsUseCase) SetAuthorizer(auth rbac.IRBACService) {
	uc.authorizer = auth
}

type ListProjectsParams struct {
	UserID          string
	IncludeArchived bool
}

type ListProjectsResult struct {
	Projects []*domain.Project
}

func (uc *ListProjectsUseCase) ListForUser(ctx context.Context, params *ListProjectsParams) (*ListProjectsResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.projects == nil {
		return nil, ErrRepositoryNotConfigured
	}

	assignments, hasAssignments := rbac.GetCurrentUserRoleAssignments(ctx)
	if !hasAssignments || len(assignments) == 0 {
		// Fallback for direct invocations or legacy contexts without role assignments:
		projects, err := uc.projects.GetProjectByUserID(ctx, params.UserID, params.IncludeArchived)
		if err != nil {
			return nil, fmt.Errorf("list visible projects: %w", err)
		}
		if projects == nil {
			projects = []*domain.Project{}
		}
		return &ListProjectsResult{Projects: projects}, nil
	}

	// 1. Check if the user holds global read_project_basic permission
	if uc.hasAuthorize(ctx, perms2.ScopeGlobal, params.UserID) {
		projects, err := uc.projects.ListAll(ctx, params.IncludeArchived)
		if err != nil {
			return nil, fmt.Errorf("list all projects: %w", err)
		}
		if projects == nil {
			projects = []*domain.Project{}
		}
		return &ListProjectsResult{Projects: projects}, nil
	}

	// 2. Check if the user holds any project or organization scope roles
	if uc.hasAuthorize(ctx, perms2.ScopeProject, params.UserID) {
		projects, err := uc.projects.GetProjectByUserID(ctx, params.UserID, params.IncludeArchived)
		if err != nil {
			return nil, fmt.Errorf("list visible projects: %w", err)
		}
		if projects == nil {
			projects = []*domain.Project{}
		}
		return &ListProjectsResult{Projects: projects}, nil
	}

	// 3. Normal user (no global read_project_basic, no project-scoped roles): fetch public projects only
	projects, err := uc.projects.ListPublic(ctx, params.IncludeArchived)
	if err != nil {
		return nil, fmt.Errorf("list public projects: %w", err)
	}
	if projects == nil {
		projects = []*domain.Project{}
	}
	return &ListProjectsResult{Projects: projects}, nil
}

func (uc *ListProjectsUseCase) hasAuthorize(ctx context.Context, scope perms2.ScopeLevel, userID string) bool {
	if uc.authorizer == nil {
		return false
	}
	decision := uc.authorizer.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  userID,
		Permission: perms2.PermReadProjectBasic,
		ScopeType:  string(scope),
		Resource:   string(perms2.EntityProject),
	})
	if decision.Allowed {
		uc.logger.LogAttrs(ctx, slog.LevelDebug, "access allowed", slog.String("reason", decision.Reason))
		return true
	}
	return false
}
