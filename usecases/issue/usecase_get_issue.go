package issueusecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
)

var (
	// ErrIssueIDRequired is returned when the request names no issue.
	ErrIssueIDRequired = domain.ErrIssueIDRequired
	// ErrIssueNotVisible is returned when the issue does not exist or its project
	// is outside the caller's scope. It aliases the domain sentinel so the
	// handler can map it without importing the storage package.
	ErrIssueNotVisible = domain.ErrIssueNotVisible
)

// GetIssueUseCase resolves a single issue the authenticated user is allowed to
// read.
//
// As with the listings, a missing issue and an out-of-scope one share
// ErrIssueNotVisible. Distinguishing them would turn this endpoint into an
// existence oracle for issue IDs across project boundaries.
type GetIssueUseCase struct {
	issues IssueReader
	logger *slog.Logger
}

// NewGetIssue builds the get-issue use case.
func NewGetIssue(logger *slog.Logger, issues IssueReader) *GetIssueUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetIssueUseCase{
		issues: issues,
		logger: logger,
	}
}

type GetIssueParams struct {
	UserID  string
	IssueID string
}

type GetIssueResult struct {
	Issue *domain.Issue
}

func (uc *GetIssueUseCase) Execute(ctx context.Context, params *GetIssueParams) (*GetIssueResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if params.IssueID == "" {
		return nil, ErrIssueIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.issues == nil {
		return nil, ErrRepositoryNotConfigured
	}

	issue, err := uc.issues.GetVisibleToUser(ctx, params.UserID, params.IssueID)
	if err != nil {
		if errors.Is(err, domain.ErrIssueNotVisible) {
			return nil, ErrIssueNotVisible
		}
		return nil, fmt.Errorf("get issue %q: %w", params.IssueID, err)
	}
	if issue == nil {
		// Defensive: a repository returning (nil, nil) would otherwise reach the
		// handler as a 200 with a null body.
		return nil, ErrIssueNotVisible
	}

	return &GetIssueResult{Issue: issue}, nil
}
