package projectshandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	projectusecase "youtrack_backend/internal/usecases/project"
)

// opCreateProject labels the failures this route owns in the error envelope.
const opCreateProject = "projects.create"

// createProjectBodyLimit caps the request body. A project is a few short
// strings; a megabyte is already far more than any legitimate form sends, and
// the cap keeps a hostile client from making the server buffer without bound.
const createProjectBodyLimit = 1 << 20

// createProjectRequest is the writable half of a project. It deliberately omits
// the server-owned fields — id, creationTime, and the per-identity favorite
// flag — so a client cannot dictate them: the identity comes from the session
// and the rest is minted here.
type createProjectRequest struct {
	Name           string `json:"name"`
	ShortName      string `json:"shortName"`
	ProjectTypeID  string `json:"projectTypeId"`
	IconURL        string `json:"iconUrl"`
	Template       bool   `json:"template"`
	Archived       bool   `json:"archived"`
	Restricted     bool   `json:"restricted"`
	HasArticles    bool   `json:"hasArticles"`
	IsDemo         bool   `json:"isDemo"`
	Description    string `json:"description"`
	IssuesURL      string `json:"issuesUrl"`
	LeaderID       string `json:"leaderId"`
	FromEmail      string `json:"fromEmail"`
	ReplyToEmail   string `json:"replyToEmail"`
	TeamID         string `json:"teamId"`
	OrganizationID string `json:"organizationId"`
	StartingNumber int64  `json:"startingNumber"`
}

// createProjectResponse is the project as stored, in the same shape the read
// routes return. A create answers the whole entity rather than a bare id so a
// client can navigate to the new project without the follow-up GET that
// permission scoping might refuse.
type createProjectResponse = ProjectsResponse

// violations collects every field problem at once, so a client fixing a form
// learns about all of them in one round trip instead of one per submit.
func (p createProjectRequest) violations() *platErrors.ValidationError {
	invalid := platErrors.NewValidationError("project")

	name := strings.TrimSpace(p.Name)
	switch {
	case name == "":
		invalid.AddViolation("name", "is required")
	case len(name) > projectusecase.MaxProjectNameLength:
		invalid.AddViolation("name", "must be at most 255 characters")
	}

	if short := strings.TrimSpace(p.ShortName); len(short) > projectusecase.MaxProjectShortNameLength {
		invalid.AddViolation("shortName", "must be at most 50 characters")
	}

	// Zero means "not supplied" and is defaulted downstream; a negative value
	// could never be honoured, so it is rejected here rather than silently
	// becoming 1.
	if p.StartingNumber < 0 {
		invalid.AddViolation("startingNumber", "must be greater than zero")
	}

	return invalid
}

// handleCreateProject answers POST /api/projects.
//
// The route is gated on the global create-project permission, which is the only
// project permission that is global: a project cannot be scoped to itself
// before it exists. That gate is why this handler does not re-check
// authorization, and why the identity is taken from the session rather than
// the body.
func (hlr *Handler) handleCreateProject(wr http.ResponseWriter, req *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(req.Context())
	if !ok || userID == "" {
		response.Error(wr, platErrors.Unauthorized(opCreateProject, "authentication required",
			errors.New("missing user identity")))
		return
	}

	var payload createProjectRequest
	req.Body = http.MaxBytesReader(wr, req.Body, createProjectBodyLimit)
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Error(wr, platErrors.Invalid(opCreateProject, "invalid request body", err))
		return
	}

	if invalid := payload.violations(); invalid.HasViolations() {
		response.Error(wr, invalid.AsError(opCreateProject))
		return
	}

	result, err := hlr.createProject.Execute(req.Context(), &projectusecase.CreateProjectParams{
		UserID:         userID,
		Name:           payload.Name,
		ShortName:      payload.ShortName,
		ProjectTypeID:  payload.ProjectTypeID,
		IconURL:        payload.IconURL,
		Template:       payload.Template,
		Archived:       payload.Archived,
		Restricted:     payload.Restricted,
		HasArticles:    payload.HasArticles,
		IsDemo:         payload.IsDemo,
		Description:    payload.Description,
		IssuesURL:      payload.IssuesURL,
		LeaderID:       payload.LeaderID,
		FromEmail:      payload.FromEmail,
		ReplyToEmail:   payload.ReplyToEmail,
		TeamID:         payload.TeamID,
		OrganizationID: payload.OrganizationID,
		StartingNumber: payload.StartingNumber,
	})
	if err != nil {
		hlr.logger.LogAttrs(req.Context(), slog.LevelWarn, "project creation rejected",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		writeProjectsError(wr, err, "failed to create project")
		return
	}

	hlr.logger.LogAttrs(req.Context(), slog.LevelInfo, "project created",
		slog.String("user_id", userID),
		slog.String("project_id", result.Project.ID),
		slog.String("short_name", result.Project.ShortName),
	)

	response.Created(wr, createProjectResponse(newProjectParams(result.Project)))
}
