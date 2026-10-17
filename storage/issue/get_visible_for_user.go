package issuerepository

import (
	"context"
	"errors"
	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	perms2 "youtrack_backend/internal/services/permissions_services"

	"github.com/jackc/pgx/v5"
)

const getVisibleIssueQuery = issueVisibilityCTE + issueSelectColumns + issueVisibilityPredicate + `
	AND i.id = $3
`

func (r *PostgresRepository) GetVisibleToUser(ctx context.Context, userID, issueID string) (*domain.Issue, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if issueID == "" {
		return nil, ErrIssueIDRequired
	}

	grantingRefs := grantingPermissionRefs(r.permSvc, perms2.PermReadIssue)
	found, err := scanIssue(r.db.QueryRow(ctx, getVisibleIssueQuery, userID, grantingRefs, issueID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrIssueNotVisible
		}
		return nil, postgres.TranslateError(issueGetOp, err)
	}
	logIssues(ctx, issueGetOp, []*domain.Issue{found})
	return found, nil
}
