package authandler

import (
	"errors"
	"log/slog"
	"net/http"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
)

func (handler *Handler) handleLogout(writer http.ResponseWriter, request *http.Request) {
	sessionID, ok := middlewares.GetSessionID(request.Context())
	if !ok {
		handler.clearCookie(writer)
		response.Error(writer, platErrors.Unauthorized("logout", "authentication required", errors.New("missing session id")))
		return
	}

	userID, _ := middlewares.GetCurrentUserID(request.Context())

	if err := handler.logout.Logout(request.Context(), sessionID, userID); err != nil {
		handler.logger.LogAttrs(request.Context(), slog.LevelError, "logout failed",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		http.Error(writer, "logout failed", http.StatusInternalServerError)
		return
	}

	handler.clearCookie(writer)
	response.JSON(writer, http.StatusOK, map[string]bool{"success": true})
}
