package issuerepository

import (
	"context"
	"sort"
	"strings"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

func (r *PostgresRepository) ListVisibleToUser(ctx context.Context, userID string, query *domain.IssueQuery) ([]*domain.Issue, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	normalized := normalizeQuery(query)
	grantingRefs := grantingPermissionRefs(r.permSvc, perms2.PermReadIssue)

	rows, err := r.db.Query(ctx, buildListQuery(normalized.SortField, normalized.Ascending),
		userID,
		grantingRefs,
		normalized.ProjectID,
		normalized.SearchQuery,
		normalized.State,
		normalized.Priority,
		normalized.IssueType,
		normalized.AssigneeID,
		normalized.TagID,
		normalized.Limit,
		normalized.Offset,
	)
	if err != nil {
		return nil, postgres.TranslateError(issueListOp, err)
	}
	defer rows.Close()

	issues, err := scanIssues(rows)
	if err == nil {
		logIssues(ctx, issueListOp, issues)
	}
	return issues, err
}

func buildListQuery(sortField domain.IssueSortField, ascending bool) string {
	clause := orderByClause(sortField, ascending)
	return strings.Replace(listVisibleIssuesQuery, issueOrderByPlaceholder, clause, 1)
}

func orderByClause(field domain.IssueSortField, ascending bool) string {
	resolved, ascending := resolveSort(field, ascending)

	expression, ok := sortExpression(resolved)
	if !ok {
		expression, _ = sortExpression(domain.IssueSortUpdated)
		ascending = false
	}
	if ascending {
		return expression + " ASC, i.id ASC"
	}
	return expression + " DESC, i.id ASC"
}

func resolveSort(field domain.IssueSortField, ascending bool) (domain.IssueSortField, bool) {
	if _, ok := sortExpression(field); !ok {
		field = domain.IssueSortUpdated
	}
	if field == domain.IssueSortPriority {
		return field, true
	}
	return field, ascending
}

func sortIssues(issues []*domain.Issue, query domain.IssueQuery) {
	field, ascending := resolveSort(query.SortField, query.Ascending)

	less := func(i, j *domain.Issue) bool { return false }
	switch field {
	case domain.IssueSortCreated:
		less = func(i, j *domain.Issue) bool { return i.CreatedAt < j.CreatedAt }
	case domain.IssueSortVotes:
		less = func(i, j *domain.Issue) bool { return i.Votes < j.Votes }
	case domain.IssueSortSummary:
		less = func(i, j *domain.Issue) bool { return i.Summary < j.Summary }
	case domain.IssueSortPriority:
		less = func(i, j *domain.Issue) bool {
			return priorityRank(i.Priority) < priorityRank(j.Priority)
		}
	case domain.IssueSortUpdated:
		less = func(i, j *domain.Issue) bool { return i.UpdatedAt < j.UpdatedAt }
	}

	sort.SliceStable(issues, func(a, b int) bool {
		if ascending {
			if less(issues[a], issues[b]) {
				return true
			}
			if less(issues[b], issues[a]) {
				return false
			}
		} else {
			if less(issues[b], issues[a]) {
				return true
			}
			if less(issues[a], issues[b]) {
				return false
			}
		}
		return issues[a].ID < issues[b].ID
	})
}

func priorityRank(priority string) int {
	switch priority {
	case domain.IssuePriorityShowStopper:
		return 0
	case domain.IssuePriorityCritical:
		return 1
	case domain.IssuePriorityMajor:
		return 2
	case domain.IssuePriorityNormal:
		return 3
	case domain.IssuePriorityMinor:
		return 4
	default:
		return 5
	}
}

func sortExpression(field domain.IssueSortField) (string, bool) {
	switch field {
	case domain.IssueSortCreated:
		return "i.created", true
	case domain.IssueSortPriority:
		return `CASE i.priority
			WHEN 'show-stopper' THEN 0
			WHEN 'critical'    THEN 1
			WHEN 'major'       THEN 2
			WHEN 'normal'      THEN 3
			WHEN 'minor'       THEN 4
			ELSE 5
		END`, true
	case domain.IssueSortVotes:
		return "i.votes", true
	case domain.IssueSortSummary:
		return "i.summary", true
	case domain.IssueSortUpdated:
		return "i.updated", true
	default:
		return "", false
	}
}
