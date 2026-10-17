// Package issue provides the persistence layer for the issues an identity is
// allowed to read.
//
// Visibility is not a column on `issues`. It is inherited from the issue's
// project, which in turn is derived from the role assignments the identity
// holds, so this package resolves the whole chain while querying rather than
// trusting a caller-supplied project filter.
package issuerepository

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

var (
	ErrUserIDRequired    = domain.ErrUserIDRequired
	ErrProjectIDRequired = domain.ErrProjectIDRequired
	ErrIssueIDRequired   = domain.ErrIssueIDRequired
	ErrProjectNotVisible = domain.ErrProjectNotVisible
	ErrIssueNotVisible   = domain.ErrIssueNotVisible
)

const (
	defaultLimit = 100
	maxLimit     = 500
)

type Repository interface {
	ListVisibleToUser(ctx context.Context, userID string, query *domain.IssueQuery) ([]*domain.Issue, error)
	ListForProject(ctx context.Context, userID, projectID string, query *domain.IssueQuery) ([]*domain.Issue, error)
	GetVisibleToUser(ctx context.Context, userID, issueID string) (*domain.Issue, error)
}

func NewRepository(db postgres.DBTX) Repository {
	if db == nil {
		panic("db can not be nil")
	}
	return &PostgresRepository{
		db:      db,
		permSvc: perms2.NewService(),
	}
}

func normalizeQuery(query *domain.IssueQuery) domain.IssueQuery {
	var normalized domain.IssueQuery
	if query != nil {
		normalized = *query
		normalized.ResolveProject()
	}

	normalized.SearchQuery = strings.TrimSpace(normalized.SearchQuery)

	if _, ok := sortExpression(normalized.SortField); !ok {
		normalized.SortField = domain.IssueSortUpdated
	}

	switch {
	case normalized.Limit <= 0:
		normalized.Limit = defaultLimit
	case normalized.Limit > maxLimit:
		normalized.Limit = maxLimit
	}
	if normalized.Offset < 0 {
		normalized.Offset = 0
	}

	return normalized
}

func matchesIssueQuery(issue *domain.Issue, query domain.IssueQuery, issueTags map[string]struct{}) bool {
	if query.SearchQuery != "" {
		needle := strings.ToLower(query.SearchQuery)
		if !strings.Contains(strings.ToLower(issue.IssueKey), needle) &&
			!strings.Contains(strings.ToLower(issue.Summary), needle) {
			return false
		}
	}
	if query.State != "" && issue.State != query.State {
		return false
	}
	if query.Priority != "" && issue.Priority != query.Priority {
		return false
	}
	if query.IssueType != "" && issue.IssueType != query.IssueType {
		return false
	}
	if query.TagID != "" {
		if _, tagged := issueTags[query.TagID]; !tagged {
			return false
		}
	}
	switch query.AssigneeID {
	case "":
	case domain.UnassignedAssignee:
		if issue.AssigneeID != nil && *issue.AssigneeID != "" {
			return false
		}
	default:
		if issue.AssigneeID == nil || *issue.AssigneeID != query.AssigneeID {
			return false
		}
	}
	return true
}

// pageIssues sorts and then applies offset/limit.
func pageIssues(issues []*domain.Issue, query domain.IssueQuery) []*domain.Issue {
	sortIssues(issues, query)

	if query.Offset >= len(issues) {
		return []*domain.Issue{}
	}
	windowed := issues[query.Offset:]
	if len(windowed) > query.Limit {
		windowed = windowed[:query.Limit]
	}

	// A fresh slice: the caller must not be able to reach the backing array and
	// overwrite a neighbouring element.
	paged := make([]*domain.Issue, len(windowed))
	copy(paged, windowed)
	return paged
}

// ---------------------------------------------------------------------------
// PostgreSQL implementation
// ---------------------------------------------------------------------------

// PostgresRepository implements Repository against the `issues` table.
type PostgresRepository struct {
	db      postgres.DBTX
	permSvc perms2.IPermissionService
}

// NewPostgresRepository creates a new PostgresRepository instance.
func NewPostgresRepository(db postgres.DBTX) *PostgresRepository {
	return &PostgresRepository{
		db:      db,
		permSvc: perms2.NewService(),
	}
}

const (
	issueListOp    = "GetListIssuesToUser"
	issueGetOp     = "issues.GetIssuetByID"
	issueProjectOp = "GetListIssuesForProjectID"
)

const issueVisibilityCTE = `
	WITH RECURSIVE
	membership_edge(member_type, member_id, group_id) AS (
		SELECT 'USER', gm.user_id, gm.group_id
		FROM group_members gm
		WHERE gm.member_type = 'USER' AND gm.user_id IS NOT NULL
		UNION
		SELECT 'GROUP', gm.member_group_id, gm.group_id
		FROM group_members gm
		WHERE gm.member_type = 'GROUP' AND gm.member_group_id IS NOT NULL
		UNION
		SELECT 'GROUP', g.id, g.parent_group_id
		FROM groups g
		WHERE g.parent_group_id IS NOT NULL
	),
	user_group_ids(group_id) AS (
		SELECT e.group_id
		FROM membership_edge e
		WHERE e.member_type = 'USER' AND e.member_id = $1
		UNION
		SELECT e.group_id
		FROM user_group_ids ugi
		JOIN membership_edge e ON e.member_type = 'GROUP' AND e.member_id = ugi.group_id
	),
	identity_roles AS (
		SELECT DISTINCT
			ar.role_id,
			ar.scope_type,
			COALESCE(ar.project_id, ar.organization_id, '') AS scope_id
		FROM assigned_roles ar
		WHERE ar.user_id = $1
		   OR ar.group_id IN (SELECT group_id FROM user_group_ids)
	),
	readable_projects AS (
		SELECT DISTINCT ir.scope_type, ir.scope_id
		FROM identity_roles ir
		JOIN roles r ON r.id = ir.role_id
		CROSS JOIN LATERAL unnest(r.permissions) AS perm
		WHERE perm = ANY($2)
	)
`

const issueSelectColumns = `
	SELECT
		i.id,
		i.id_readable,
		COALESCE(i.number_in_project, 0),
		i.summary,
		COALESCE(i.description, ''),
		i.state,
		i.priority,
		i.issue_type,
		i.assignee_id,
		COALESCE(i.reporter_id, ''),
		i.subsystem_id,
		COALESCE(i.fix_versions, ''),
		COALESCE((EXTRACT(EPOCH FROM i.created) * 1000)::BIGINT, 0),
		COALESCE((EXTRACT(EPOCH FROM i.updated) * 1000)::BIGINT, 0),
		i.due_date,
		i.estimation,
		i.spent_time,
		COALESCE(i.votes, 0),
		COALESCE(i.is_draft, false),
		i.project_id,
		(
			SELECT COUNT(*) FROM issue_watchers iw
			WHERE iw.issue_id = i.id
		) AS watchers_count,
		(
			SELECT COUNT(*) FROM issue_attachments a
			WHERE a.issue_id = i.id
		) AS attachments_count,
		(
			SELECT COUNT(*) FROM issue_comments c
			WHERE c.issue_id = i.id AND c.deleted = false
		) AS comments_count,
		COALESCE((
			SELECT EXISTS (
				SELECT 1 FROM issue_watchers iw
				WHERE iw.issue_id = i.id AND iw.user_id = $1 AND iw.has_star = true
			)
		), false) AS is_starred
	FROM issues i
`

const issueVisibilityPredicate = `
	JOIN projects p ON p.id = i.project_id
	WHERE EXISTS (
		SELECT 1
		FROM readable_projects rp
		WHERE rp.scope_type = 'GLOBAL'
		   OR (rp.scope_type = 'ORGANIZATION' AND rp.scope_id = COALESCE(p.organization_id, ''))
		   OR (rp.scope_type = 'PROJECT' AND rp.scope_id = p.id)
	)
`

const issueFilterPredicate = `
	AND ($3 = '' OR i.project_id = $3)
	AND ($4 = '' OR i.id_readable ILIKE '%' || $4 || '%' OR i.summary ILIKE '%' || $4 || '%')
	AND ($5 = '' OR i.state = $5)
	AND ($6 = '' OR i.priority = $6)
	AND ($7 = '' OR i.issue_type = $7)
	AND (
		$8 = ''
		OR ($8 = '__unassigned__' AND (i.assignee_id IS NULL OR i.assignee_id = ''))
		OR i.assignee_id = $8
	)
	AND (
		$9 = ''
	)
`

const listVisibleIssuesQuery = issueVisibilityCTE + issueSelectColumns + issueVisibilityPredicate +
	issueFilterPredicate + `
	ORDER BY ` + issueOrderByPlaceholder + `
	LIMIT $10 OFFSET $11
`

const issueOrderByPlaceholder = "__ISSUE_ORDER_BY__"

func scanIssues(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]*domain.Issue, error) {
	issues := make([]*domain.Issue, 0, 16)
	for rows.Next() {
		found, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, found)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return issues, nil
}

func scanIssue(row interface{ Scan(...any) error }) (*domain.Issue, error) {
	var (
		found       domain.Issue
		dueDate     *int64
		estimation  *int
		spentTime   *int
		isStarred   bool
		watchers    int
		attachments int
		comments    int
	)

	if err := row.Scan(
		&found.ID,
		&found.IssueKey,
		&found.IssueNumber,
		&found.Summary,
		&found.Description,
		&found.State,
		&found.Priority,
		&found.IssueType,
		&found.AssigneeID,
		&found.ReporterID,
		&found.SubsystemID,
		&found.FixVersions,
		&found.CreatedAt,
		&found.UpdatedAt,
		&dueDate,
		&estimation,
		&spentTime,
		&found.Votes,
		&found.IsDraft,
		&found.ProjectID,
		&watchers,
		&attachments,
		&comments,
		&isStarred,
	); err != nil {
		return nil, err
	}

	found.DueDate = dueDate
	found.Estimation = estimation
	found.SpentTime = spentTime
	found.IsStarred = isStarred
	found.Watchers = watchers
	found.Attachments = attachments
	found.Comments = comments
	found.Tags = []map[string]any{}
	found.Sprints = []map[string]any{}
	found.Links = []map[string]any{}
	found.Visibility = map[string][]string{"users": {}, "groups": {}}

	return &found, nil
}

func grantingPermissionRefs(permSvc perms2.IPermissionService, target perms2.PermKey) []string {
	if permSvc == nil {
		permSvc = perms2.NewService()
	}
	seen := make(map[string]struct{})
	var refs []string
	addRef := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			refs = append(refs, s)
		}
	}

	addRef(string(target))
	targetPerm := permSvc.GetPermission(target)
	if targetPerm != nil {
		addRef(targetPerm.DisplayName)
	}

	for _, p := range permSvc.GetAllPermissions() {
		implied := permSvc.ResolveImplied([]perms2.PermKey{p.ID})
		if slices.Contains(implied, target) {
			addRef(string(p.ID))
			addRef(p.DisplayName)
		}
	}

	return refs
}

func logIssues(ctx context.Context, op string, projects []*domain.Issue, args ...slog.Attr) {
	if len(projects) == 0 {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
			//slog.String("issues", "[]"),
			args...,
		)
		return
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op, args...)
	for index, issue := range projects {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "  "+strconv.Itoa((index+1)),
			slog.String("id", issue.ID),
			slog.String("projectID", issue.ProjectID),
			slog.String("issueKey", issue.IssueKey),
			slog.String("priority", issue.Priority),
			slog.String("state", issue.State),
		)
	}
}
