package projectshandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	projectusecase "youtrack_backend/internal/usecases/project"

	"github.com/go-chi/chi/v5"
)

const (
	memberType     = "jetbrains.jetpass.tracker.UserProjectRole"
	memberUserType = "jetbrains.jetpass.tracker.User"
)

type MemberUserParamsResponse struct {
	ID   string `json:"id"`
	Type string `json:"$type"`
}

type MemberParamsResponse struct {
	ID        string                    `json:"id"`
	ProjectID string                    `json:"projectId"`
	UserID    string                    `json:"userId"`
	Roles     []string                  `json:"roles"`
	IsOwner   bool                      `json:"isOwner"`
	User      *MemberUserParamsResponse `json:"user"`
	Type      string                    `json:"$type"`
}

// handleProjectMembers answers GET /api/projects/{projectID}/members.
//
// The project is named by the path and the identity by the request context, so a
// client supplies neither. A project the identity may not read is reported as
// 404 rather than 403, matching the single-project endpoint, so the roster
// cannot be used to probe which projects exist.
func (hlr *Handler) handleProjectMembers(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(request.Context())
	if !ok || userID == "" {
		response.Error(writer, platErrors.Unauthorized("projects", "authentication required",
			errors.New("missing user identity")))
		return
	}

	projectID := strings.TrimSpace(chi.URLParam(request, "projectID"))
	if projectID == "" {
		response.Error(writer, platErrors.Invalid("projects", "project id is required",
			projectusecase.ErrProjectIDRequired))
		return
	}

	result, err := hlr.projectMembers.Execute(request.Context(), &projectusecase.ListProjectMembersParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, "project members rejected",
			slog.String("user_id", userID),
			slog.String("project_id", projectID),
			slog.String("error", err.Error()),
		)
		writeProjectsError(writer, err, "failed to list project members")
		return
	}

	response.JSON(writer, http.StatusOK, result.Members)
}
