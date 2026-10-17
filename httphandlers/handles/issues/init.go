package issueshandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	perms2 "youtrack_backend/internal/services/permissions_services"
	boardusecase "youtrack_backend/internal/usecases/board"
	issueusecase "youtrack_backend/internal/usecases/issue"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger        *slog.Logger
	issues        *issueusecase.ListIssuesUseCase
	issueByID     *issueusecase.GetIssueUseCase
	projectIssues *issueusecase.ListProjectIssuesUseCase
	moveCard      *boardusecase.MoveCardUseCase
}

func NewHandler(
	logger *slog.Logger,
	issues *issueusecase.ListIssuesUseCase,
	issueByID *issueusecase.GetIssueUseCase,
	projectIssues *issueusecase.ListProjectIssuesUseCase,
	moveCard *boardusecase.MoveCardUseCase,
) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:        logger,
		issues:        issues,
		issueByID:     issueByID,
		projectIssues: projectIssues,
	}
}

func NewRoute(parent chi.Router, handler *Handler, protector middlewares.Protector, svc middlewares.AuthorizerService) *chi.Router {
	permissionRequired := func(permission perms2.PermKey, hlrFunc http.HandlerFunc) http.HandlerFunc {
		return middlewares.RequireGlobal(svc, permission, hlrFunc)
	}

	router := parent.Route("/api/issues", func(ir chi.Router) {
		ir.Group(func(g chi.Router) {
			g.Use(protector)
			g.Get("/", permissionRequired(perms2.PermReadIssue, handler.handleIssues))
			g.Get("/{issueID}", permissionRequired(perms2.PermReadIssue, handler.handleIssueByID))
			g.Get("/{issueID}/state", permissionRequired(perms2.PermReadIssue, handler.handleMoveCard))
		})
	})
	return &router
}

func (hlr *Handler) logRejected(request *http.Request, userID, message string, err error) {
	hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, message,
		slog.String("user_id", userID),
		slog.String("error", err.Error()),
	)
}

func writeIssuesError(writer http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, issueusecase.ErrUserIDRequired):
		response.Error(writer, platErrors.Unauthorized("issues", "authentication required", err))
	case errors.Is(err, issueusecase.ErrProjectIDRequired):
		response.Error(writer, platErrors.Invalid("issues", "project id is required", err))
	case errors.Is(err, issueusecase.ErrIssueIDRequired):
		response.Error(writer, platErrors.Invalid("issues", "issue id is required", err))
	case errors.Is(err, issueusecase.ErrProjectNotVisible):
		response.Error(writer, platErrors.NotFound("issues", "project not found", err))
	case errors.Is(err, issueusecase.ErrIssueNotVisible):
		response.Error(writer, platErrors.NotFound("issues", "issue not found", err))
	default:
		response.Error(writer, platErrors.Internal("issues", fallback, err))
	}
}

func IssuesQuery(request *http.Request) *domain.IssueQuery {
	values := request.URL.Query()

	query := &domain.IssueQuery{
		SearchQuery: strings.TrimSpace(firstNonEmpty(values.Get("search"), values.Get("query"))),
		State:       values.Get("state"),
		Priority:    values.Get("priority"),
		IssueType:   firstNonEmpty(values.Get("issueType"), values.Get("issue_type")),
		AssigneeID:  values.Get("assignee"),
		TagID:       firstNonEmpty(values.Get("tag"), values.Get("tagId")),
		ProjectID:   requestedProjectID(request),
		SortField:   parseSortField(firstNonEmpty(values.Get("sort"), values.Get("sortField"))),
	}

	query.Ascending = parseBoolDefault(firstNonEmpty(values.Get("ascending"), values.Get("asc")), false)

	if limit, err := strconv.Atoi(values.Get("$top")); err == nil {
		query.Limit = limit
	} else if limit, err := strconv.Atoi(values.Get("limit")); err == nil {
		query.Limit = limit
	}

	if offset, err := strconv.Atoi(values.Get("$skip")); err == nil {
		query.Offset = offset
	} else if offset, err := strconv.Atoi(values.Get("offset")); err == nil {
		query.Offset = offset
	}

	return query
}

func parseBoolDefault(raw string, fallback bool) bool {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func requestedProjectID(request *http.Request) string {
	values := request.URL.Query()
	return firstNonEmpty(
		chi.URLParam(request, "projectID"),
		values.Get("project"),
		values.Get("projectId"),
		values.Get("projectID"),
	)
}

func ProjectIDExtractor() middlewares.IDExtractor {
	return func(request *http.Request) (string, error) {
		if projectID := requestedProjectID(request); projectID != "" {
			return projectID, nil
		}
		return "", errors.New("no project in the route or the query string: " +
			"pass ?projectID=0-0, or use GET /api/projects/{projectID}/issues")
	}
}

func parseSortField(raw string) domain.IssueSortField {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(domain.IssueSortCreated):
		return domain.IssueSortCreated
	case string(domain.IssueSortPriority):
		return domain.IssueSortPriority
	case string(domain.IssueSortVotes):
		return domain.IssueSortVotes
	case string(domain.IssueSortSummary):
		return domain.IssueSortSummary
	case string(domain.IssueSortUpdated):
		return domain.IssueSortUpdated
	default:
		return domain.IssueSortUpdated
	}
}
