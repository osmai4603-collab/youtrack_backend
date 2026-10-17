package projectusecase

import (
	"context"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"
)

// ProjectGetter is the data access contract needed to load one project. It is
// satisfied structurally by storage/project.Repository, so no hard import ties
// this layer to the storage implementation.
type ProjectGetter interface {
	GetProjectByID(ctx context.Context, projectID string) (*domain.Project, error)
}

// GetProjectUseCase resolves a single project the authenticated identity may
// read.
//
// Authorization is not re-checked here: the repository derives visibility from
// the caller's own role assignments and returns ErrProjectNotVisible for both an
// absent project and one outside the caller's scope, so this layer can neither
// widen access nor leak which project IDs exist.
type GetProjectUseCase struct {
	projects ProjectGetter
	logger   *slog.Logger
	rbacSvc  rbac.IRBACService
}

// NewGetProject builds the get-project use case.
func NewGetProject(logger *slog.Logger, projects ProjectGetter, rbacSvc rbac.IRBACService) *GetProjectUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetProjectUseCase{
		projects: projects,
		logger:   logger,
		rbacSvc:  rbacSvc,
	}
}

type GetProjectParams struct {
	UserID    string
	ProjectID string
}

type GetProjectResult struct {
	Project *domain.Project
}

func (uc *GetProjectUseCase) Execute(ctx context.Context, params *GetProjectParams) (*GetProjectResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if params.ProjectID == "" {
		return nil, ErrProjectIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.projects == nil {
		return nil, ErrRepositoryNotConfigured
	}

	access := requestAccess(params.ProjectID, params.UserID)
	if err := uc.rbacSvc.Authorize(ctx, access); err != nil {
		return nil, err
	}

	project, err := uc.projects.GetProjectByID(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("get visible project: %w", err)
	}

	return &GetProjectResult{Project: project}, nil
}

func requestAccess(projectID string, userID string) rbac.AccessRequest {
	return rbac.AccessRequest{
		SubjectID:  userID,
		Permission: perms2.PermReadProjectBasic,
		Resource:   string(perms2.EntityProject),
		ResourceID: projectID,
		ScopeType:  string(perms2.ScopeProject),
	}
}
