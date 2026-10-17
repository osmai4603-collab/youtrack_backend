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

func (hlr *Handler) handleProjectByID(writer http.ResponseWriter, request *http.Request) {
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

	result, err := hlr.projectByID.Execute(request.Context(), &projectusecase.GetProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, "project fetch rejected",
			slog.String("user_id", userID),
			slog.String("project_id", projectID),
			slog.String("error", err.Error()),
		)
		writeProjectsError(writer, err, "failed to load project")
		return
	}

	response.JSON(writer, http.StatusOK, newProjectParams(result.Project))
}
