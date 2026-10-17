package issueusecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
)

var (
	// ErrUserIDRequired is returned when the request carries no identity. It
	// aliases the domain sentinel so an error raised in the storage layer
	// still satisfies errors.Is here.
	ErrUserIDRequired = domain.ErrUserIDRequired
	// ErrRepositoryNotConfigured is returned when the use case was built
	// without an issue repository.
	ErrRepositoryNotConfigured = errors.New("issue repository is not configured")
)

// IssueReader is the data access contract needed to read the issues an identity
// may see. It is satisfied structurally by storage/issue.Repository, so no hard
// import ties this layer to the storage implementation.
type IssueReader interface {
	ListVisibleToUser(ctx context.Context, userID string, query *domain.IssueQuery) ([]*domain.Issue, error)
	ListForProject(ctx context.Context, userID, projectID string, query *domain.IssueQuery) ([]*domain.Issue, error)
	GetVisibleToUser(ctx context.Context, userID, issueID string) (*domain.Issue, error)
}

// ListIssuesUseCase resolves the issues the authenticated user is allowed to
// read.
//
// The listing is project-scoped: the route guards the request in a project scope
// and passes the project down in the query, so this use case reads it back and
// delegates to ListForProject. When no project is named it falls back to
// ListVisibleToUser, which spans every project the caller can read.
//
// Authorization is not re-checked here in either branch: the repository derives
// visibility from the caller's own role assignments, so the result can only ever
// contain issues the identity is entitled to. That is also what makes the
// unfiltered listing safe — there is no way to ask it for "all issues regardless
// of scope".
type ListIssuesUseCase struct {
	issues IssueReader
	logger *slog.Logger
}

// NewListIssues builds the list-issues use case.
func NewListIssues(logger *slog.Logger, issues IssueReader) *ListIssuesUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ListIssuesUseCase{
		issues: issues,
		logger: logger,
	}
}

type ListIssuesParams struct {
	UserID string
	Query  *domain.IssueQuery
}

type ListIssuesResult struct {
	Issues []*domain.Issue
}

func (uc *ListIssuesUseCase) Execute(ctx context.Context, params *ListIssuesParams) (*ListIssuesResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.issues == nil {
		return nil, ErrRepositoryNotConfigured
	}

	query := params.Query
	if query == nil {
		query = &domain.IssueQuery{}
	}
	// Fold any remaining project spelling into the one field the repository
	// reads, so the project this list is filtered on is the project the guard
	// authorized. Without this the guard and the query could name different
	// projects, and the answer would be about a scope nobody checked.
	query.ResolveProject()

	var issues []*domain.Issue
	var err error
	if query.ProjectID != "" {
		issues, err = uc.issues.ListForProject(ctx, params.UserID, query.ProjectID, query)
	} else {
		issues, err = uc.issues.ListVisibleToUser(ctx, params.UserID, query)
	}
	if err != nil {
		return nil, fmt.Errorf("list visible issues: %w", err)
	}

	if issues == nil {
		issues = []*domain.Issue{}
	}

	return &ListIssuesResult{Issues: issues}, nil
}
