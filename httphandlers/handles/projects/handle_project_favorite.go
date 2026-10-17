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

	"github.com/go-chi/chi/v5"
)

type FavoriteParamsRequest struct {
	Favorite *bool `json:"favorite"`
}

type FavoriteParamsResponse struct {
	ID       string `json:"id"`
	Favorite bool   `json:"favorite"`
}

func (hlr *Handler) handleProjectFavorite(writer http.ResponseWriter, request *http.Request) {
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

	var params FavoriteParamsRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<16)
	if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
		response.Error(writer, platErrors.Invalid("projects", "invalid request body", err))
		return
	}
	if params.Favorite == nil {
		response.Error(writer, platErrors.Invalid("projects", "the favorite field is required",
			errors.New("missing favorite")))
		return
	}

	result, err := hlr.favorites.Execute(request.Context(), &projectusecase.SetProjectFavoriteParams{
		UserID:    userID,
		ProjectID: projectID,
		Favorite:  *params.Favorite,
	})
	if err != nil {
		hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, "project favorite rejected",
			slog.String("user_id", userID),
			slog.String("project_id", projectID),
			slog.Bool("favorite", *params.Favorite),
			slog.String("error", err.Error()),
		)
		writeProjectsError(writer, err, "failed to update project favorite")
		return
	}

	response.JSON(writer, http.StatusOK, FavoriteParamsResponse{
		ID:       result.ProjectID,
		Favorite: result.Favorite,
	})
}
