package projectshandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	projectusecase "youtrack_backend/internal/usecases/project"
)

// projectType is the $type discriminator clients expect on project entities.
const projectType = "jetbrains.jetpass.tracker.Project"

type ParamsRequest struct {
	IncludeArchived bool `json:"includeArchived"`
}

type ProjectsResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	ShortName      string     `json:"shortName"`
	ProjectTypeID  string     `json:"projectTypeId"`
	Pinned         bool       `json:"pinned"`
	IconURL        string     `json:"iconUrl"`
	Template       bool       `json:"template"`
	Archived       bool       `json:"archived"`
	Restricted     bool       `json:"restricted"`
	HasArticles    bool       `json:"hasArticles"`
	IsDemo         bool       `json:"isDemo"`
	Description    string     `json:"description"`
	IssuesURL      string     `json:"issuesUrl"`
	LeaderID       string     `json:"leaderId"`
	CreatedBy      string     `json:"createdBy"`
	FromEmail      string     `json:"fromEmail"`
	ReplyToEmail   string     `json:"replyToEmail"`
	TeamID         string     `json:"teamId"`
	OrganizationID string     `json:"organizationId"`
	StartingNumber int64      `json:"startingNumber"`
	CreationTime   int64      `json:"creationTime"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	Type           string     `json:"$type"`
}

// newProjectParams maps a domain project onto the wire shape.
func newProjectParams(project *domain.Project) ProjectsResponse {
	if project == nil {
		return ProjectsResponse{Type: projectType}
	}
	return ProjectsResponse{
		ID:             project.ID,
		Name:           project.Name,
		ShortName:      project.ShortName,
		ProjectTypeID:  project.ProjectTypeID,
		Pinned:         project.Pinned,
		IconURL:        project.IconURL,
		Template:       project.Template,
		Archived:       project.Archived,
		Restricted:     project.Restricted,
		HasArticles:    project.HasArticles,
		IsDemo:         project.IsDemo,
		Description:    project.Description,
		IssuesURL:      project.IssuesURL,
		LeaderID:       project.LeaderID,
		CreatedBy:      project.CreatedBy,
		FromEmail:      project.FromEmail,
		ReplyToEmail:   project.ReplyToEmail,
		TeamID:         project.TeamID,
		OrganizationID: project.OrganizationID,
		StartingNumber: project.StartingNumber,
		CreationTime:   project.CreationTime,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		Type:           projectType,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (hlr *Handler) handleProjects(wr http.ResponseWriter, req *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(req.Context())
	if !ok || userID == "" {
		response.Error(wr, platErrors.Unauthorized("projects", "authentication required",
			errors.New("missing user identity")))
		return
	}

	includeArchived := projectsQuery(req).IncludeArchived

	result, err := hlr.projects.ListForUser(req.Context(), &projectusecase.ListProjectsParams{
		UserID:          userID,
		IncludeArchived: includeArchived,
	})
	if err != nil {
		hlr.logger.LogAttrs(req.Context(), slog.LevelWarn, "projects listing rejected",
			slog.String("user_id", userID),
			slog.Bool("include_archived", includeArchived),
			slog.String("error", err.Error()),
		)
		writeProjectsError(wr, err, "failed to list projects")
		return
	}

	items := make([]ProjectsResponse, 0, len(result.Projects))
	for _, project := range result.Projects {
		items = append(items, newProjectParams(project))
	}

	response.JSON(wr, http.StatusOK, items)
}

func projectsQuery(request *http.Request) ParamsRequest {
	includeArchived, err := strconv.ParseBool(request.URL.Query().Get("archived"))
	if err != nil {
		return ParamsRequest{}
	}
	return ParamsRequest{IncludeArchived: includeArchived}
}

// writeProjectsError maps a use-case / storage error to an HTTP status code.
//
// The sentinels come from the domain package and are aliased by every layer, so
// one case here catches a failure raised in storage as well as one raised in the
// use case. The write cases are only reachable from the create and update
// routes; the read routes never produce them.
func writeProjectsError(writer http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, projectusecase.ErrUserIDRequired):
		response.Error(writer, platErrors.Unauthorized("projects", "authentication required", err))
	case errors.Is(err, projectusecase.ErrProjectIDRequired):
		response.Error(writer, platErrors.Invalid("projects", "project id is required", err))
	case errors.Is(err, projectusecase.ErrProjectNameRequired):
		response.Error(writer, platErrors.Invalid("projects", "project name is required", err))
	case errors.Is(err, projectusecase.ErrStartingNumberInvalid):
		response.Error(writer, platErrors.Invalid("projects", "starting number must be greater than zero", err))
	case errors.Is(err, projectusecase.ErrConflict):
		response.Error(writer, platErrors.Conflict("projects", "a project with this short name already exists", err))
	case errors.Is(err, projectusecase.ErrProjectNotVisible):
		response.Error(writer, platErrors.NotFound("projects", "project not found", err))
	default:
		response.Error(writer, platErrors.Internal("projects", fallback, err))
	}
}
