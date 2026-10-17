package issueusecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
)

var (
	// ErrProjectIDRequired is returned when the request names no project.
	ErrProjectIDRequired = domain.ErrProjectIDRequired
	// ErrProjectNotVisible is returned when the project does not exist or is
	// outside the caller's scope. It aliases the domain sentinel so the handler
	// can map it without importing the storage package.
	ErrProjectNotVisible = domain.ErrProjectNotVisible
)

// ListProjectIssuesUseCase resolves the issues of a single project, gated on the
// caller being able to read that project.
//
// The project check is the point of this use case. Without it, a project-scoped
// identity could pass any project ID and receive the union of issues that happen
// to live there, so the endpoint would be a way to read across project
// boundaries. ErrProjectNotVisible is deliberately shared between "no such
// project" and "not yours" for the same reason the issue read shares one
// sentinel: the response must not confirm that the ID exists.
type ListProjectIssuesUseCase struct {
	issues IssueReader
	logger *slog.Logger
}

// NewListProjectIssues builds the list-project-issues use case.
func NewListProjectIssues(logger *slog.Logger, issues IssueReader) *ListProjectIssuesUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ListProjectIssuesUseCase{
		issues: issues,
		logger: logger,
	}
}

type ListProjectIssuesParams struct {
	UserID    string
	ProjectID string
	Query     *domain.IssueQuery
}

type ListProjectIssuesResult struct {
	ProjectID string
	Issues    []*domain.Issue
}

func (uc *ListProjectIssuesUseCase) Execute(ctx context.Context, params *ListProjectIssuesParams) (*ListProjectIssuesResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if params.ProjectID == "" {
		return nil, ErrProjectIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.issues == nil {
		return nil, ErrRepositoryNotConfigured
	}

	// The path parameter overrides any project filter in the query string.
	// Letting the query win would let a caller request project A's endpoint and
	// receive project B's issues, which is a confusing way to fail open.
	query := &domain.IssueQuery{}
	if params.Query != nil {
		copied := *params.Query
		query = &copied
	}
	query.ProjectID = params.ProjectID
	query.ProjectIDAlias = ""

	issues, err := uc.issues.ListForProject(ctx, params.UserID, params.ProjectID, query)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotVisible) {
			return nil, ErrProjectNotVisible
		}
		return nil, fmt.Errorf("list issues for project %q: %w", params.ProjectID, err)
	}

	if issues == nil {
		issues = []*domain.Issue{}
	}

	return &ListProjectIssuesResult{ProjectID: params.ProjectID, Issues: issues}, nil
}
