package issueshandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	boardusecase "youtrack_backend/internal/usecases/board"

	"github.com/go-chi/chi/v5"
)

type moveCardRequest struct {
	State string `json:"state"`
}

func (h *Handler) handleMoveCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, platErrors.Unauthorized("board", "authentication required",
			errors.New("missing user identity")))
		return
	}

	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.Error(w, platErrors.Invalid("board", "issue id is required",
			boardusecase.ErrIssueIDRequired))
		return
	}

	var body moveCardRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, platErrors.Invalid("board", "invalid request body", err))
		return
	}

	if err := h.moveCard.Execute(r.Context(), &boardusecase.MoveCardParams{
		UserID:   userID,
		IssueID:  issueID,
		NewState: body.State,
	}); err != nil {
		h.logger.LogAttrs(r.Context(), slog.LevelWarn, "move card rejected",
			slog.String("user_id", userID),
			slog.String("issue_id", issueID),
			slog.String("new_state", body.State),
			slog.String("error", err.Error()),
		)
		writeBoardError(w, err, "failed to move card")
		return
	}

	response.JSON(w, http.StatusOK, response.Envelope{Success: true})
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
