package issuerepository

import (
	"context"
	"youtrack_backend/internal/domain"
)

func (r *PostgresRepository) ListForProject(ctx context.Context, userID, projectID string, query *domain.IssueQuery) ([]*domain.Issue, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}

	scoped := domain.IssueQuery{}
	if query != nil {
		scoped = *query
	}
	scoped.ProjectID = projectID

	issues, err := r.ListVisibleToUser(ctx, userID, &scoped)
	return issues, err
}
