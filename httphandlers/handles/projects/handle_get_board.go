package projectshandler

import (
	"errors"
	"net/http"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	boardusecase "youtrack_backend/internal/usecases/board"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleGetBoard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, platErrors.Unauthorized("board", "authentication required",
			errors.New("missing user identity")))
		return
	}

	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.Error(w, platErrors.Invalid("board", "project id is required",
			boardusecase.ErrProjectIDRequired))
		return
	}

	sprintID := r.URL.Query().Get("sprintID")

	data, err := h.boardDetails.Execute(r.Context(), &boardusecase.GetBoardDetailsParams{
		UserID:    userID,
		ProjectID: projectID,
		SprintID:  sprintID,
	})
	if err != nil {
		h.logger.Warn("board details rejected",
			"user_id", userID,
			"project_id", projectID,
			"error", err.Error(),
		)
		writeBoardError(w, err, "failed to load board")
		return
	}

	response.JSON(w, http.StatusOK, data)
}

// writeBoardError maps a use-case / storage error to an HTTP status code.
func writeBoardError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, boardusecase.ErrUserIDRequired):
		response.Error(w, platErrors.Unauthorized("board", "authentication required", err))
	case errors.Is(err, boardusecase.ErrProjectIDRequired):
		response.Error(w, platErrors.Invalid("board", "project id is required", err))
	case errors.Is(err, boardusecase.ErrIssueIDRequired):
		response.Error(w, platErrors.Invalid("board", "issue id is required", err))
	case errors.Is(err, boardusecase.ErrInvalidState):
		response.Error(w, platErrors.Invalid("board", "invalid issue state", err))
	case errors.Is(err, boardusecase.ErrIssueNotVisible):
		response.Error(w, platErrors.NotFound("board", "issue not found", err))
	default:
		response.Error(w, platErrors.Internal("board", fallback, err))
	}
}
