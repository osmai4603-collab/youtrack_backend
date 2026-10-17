package projectshandler

import (
	"errors"
	"log/slog"
	"net/http"

	"youtrack_backend/internal/domain"
	issueshandler "youtrack_backend/internal/httphandlers/handles/issues"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	issueusecase "youtrack_backend/internal/usecases/issue"

	"github.com/go-chi/chi/v5"
)

const issueType = "jetbrains.jetpass.tracker.Issue"

func emptyVisibility() map[string][]string {
	return map[string][]string{"users": {}, "groups": {}}
}

type ProjectIssuesResponse struct {
	ID               string              `json:"id"`
	ProjectID        string              `json:"project_id"`
	IssueKey         string              `json:"issue_key"`
	IssueSequence    int                 `json:"issue_sequence"`
	Summary          string              `json:"summary"`
	Description      string              `json:"description"`
	State            string              `json:"state"`
	Priority         string              `json:"priority"`
	IssueType        string              `json:"issue_type"`
	AssigneeID       *string             `json:"assignee_id"`
	ReporterID       string              `json:"reporter_id"`
	SubsystemID      *string             `json:"subsystem_id"`
	FixVersions      string              `json:"fix_versions"`
	ParentID         *string             `json:"parent_id"`
	BuildID          *string             `json:"build_id"`
	AssigneeAvatar   *string             `json:"assignee_avatar_url"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
	DueDate          *string             `json:"due_date"`
	Estimation       *int                `json:"estimation"`
	SpentTime        *int                `json:"spent_time"`
	Votes            int                 `json:"votes"`
	WatchersCount    int                 `json:"watchers_count"`
	AttachmentsCount int                 `json:"attachments_count"`
	CommentsCount    int                 `json:"comments_count"`
	IsStarred        bool                `json:"is_starred"`
	Tags             []any               `json:"tags"`
	Sprints          []any               `json:"sprints"`
	IssueLinks       []any               `json:"issue_links"`
	Visibility       map[string][]string `json:"visibility"`
	Type             string              `json:"$type"`
}

func newIssueParams(issue *domain.Issue) ProjectIssuesResponse {
	if issue == nil {
		return ProjectIssuesResponse{Type: issueType, Tags: []any{}, Sprints: []any{}, IssueLinks: []any{}, Visibility: emptyVisibility()}
	}

	response := ProjectIssuesResponse{
		ID:               issue.ID,
		ProjectID:        issue.ProjectID,
		IssueKey:         issue.IssueKey,
		IssueSequence:    issue.IssueNumber,
		Summary:          issue.Summary,
		Description:      issue.Description,
		State:            issue.State,
		Priority:         issue.Priority,
		IssueType:        issue.IssueType,
		AssigneeID:       issue.AssigneeID,
		ReporterID:       issue.ReporterID,
		SubsystemID:      issue.SubsystemID,
		FixVersions:      issue.FixVersions,
		ParentID:         issue.ParentID,
		CreatedAt:        issueshandler.FormatMillis(issue.CreatedAt),
		UpdatedAt:        issueshandler.FormatMillis(issue.UpdatedAt),
		DueDate:          issueshandler.FormatOptionalMillis(issue.DueDate),
		Estimation:       issue.Estimation,
		SpentTime:        issue.SpentTime,
		Votes:            issue.Votes,
		WatchersCount:    issue.Watchers,
		AttachmentsCount: issue.Attachments,
		CommentsCount:    issue.Comments,
		IsStarred:        issue.IsStarred,
		Tags:             []any{},
		Sprints:          []any{},
		IssueLinks:       []any{},
		Visibility:       issue.Visibility,
		Type:             issueType,
	}
	if response.Visibility == nil {
		response.Visibility = emptyVisibility()
	}

	return response
}

func (hlr *Handler) handleProjectIssues(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(request.Context())
	if !ok || userID == "" {
		response.Error(writer, platErrors.Unauthorized("issues", "authentication required",
			errors.New("missing user identity")))
		return
	}

	projectID := chi.URLParam(request, "projectID")
	if projectID == "" {
		writeProjectsError(writer, issueusecase.ErrProjectIDRequired, "project id is required")
		return
	}

	query := issueshandler.IssuesQuery(request)

	result, err := hlr.projectIssues.Execute(request.Context(), &issueusecase.ListProjectIssuesParams{
		UserID:    userID,
		ProjectID: projectID,
		Query:     query,
	})
	if err != nil {
		hlr.logRejected(request, userID, "project issues listing rejected", err)
		writeProjectsError(writer, err, "failed to list issues for project")
		return
	}

	items := make([]ProjectIssuesResponse, 0, len(result.Issues))
	for _, issue := range result.Issues {
		items = append(items, newIssueParams(issue))
	}

	response.JSON(writer, http.StatusOK, items)
}

func (hlr *Handler) logRejected(request *http.Request, userID, message string, err error) {
	hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, message,
		slog.String("user_id", userID),
		slog.String("error", err.Error()),
	)
}

// issuesQuery parses the listing parameters off the request.
//
// The accepted names are the client's own (camelCase) plus the YouTrack
// REST spellings, so a request built by ProjectIdFilter.toJson() behaves the
// same as a hand-written one. Unknown values are ignored rather than rejected:
// an unrecognised sort field falls back to the default ordering, and an
// unrecognised state matches nothing. Failing the whole request instead would
// make a client upgrade able to break the list page.
//
// The project is read through requestedProjectID rather than off values
// directly, so the listing is filtered by exactly the project the authorization
// guard checked. Reading the two independently is how a guard and a query end up
// talking about different projects.

// requestedProjectID is the single answer to "which project is this request
// about?", shared by the authorization guard and the listing query.
//
// Precedence is the path parameter first, because it is the more specific
// statement of intent: letting the query string override it would let a caller
// pass the guard one project and read another. The query spellings follow, with
// the client's own `project` ahead of YouTrack's camelCase `projectId` and the
// guard-facing `projectID` last.
//
// ProjectIDAlias is left empty: this function has already folded every spelling
// into ProjectID, so IssueQuery.ResolveProject has nothing left to resolve and
// the repositories only ever see one field.

// ProjectIDExtractor adapts requestedProjectID to the middleware's fail-closed
// contract: no project named is an error, never an empty id.
//
// An empty id would be indistinguishable from "no scope was asked for", and the
// guard that consumes it has no way to widen to a broader scope on purpose — so
// the error is what stops GET /api/issues from silently checking a
// project-scoped permission against global roles.
//
// Exported so a route mounted on this router by another package names the
// project the same way the routes here do.

// parseSortField accepts the closed set and falls back to the default. It is the
// only place a caller-supplied string turns into an ordering, and it can only
// ever yield one of the five known fields.
