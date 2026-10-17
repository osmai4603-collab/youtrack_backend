package authandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"youtrack_backend/internal/httphandlers/response"
	authusecase "youtrack_backend/internal/usecases/auth"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (handler *Handler) handleRefresh(writer http.ResponseWriter, request *http.Request) {
	var params refreshRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
		http.Error(writer, "invalid request body", http.StatusUnprocessableEntity)
		return
	}

	result, err := handler.refresh.Refresh(request.Context(), &authusecase.RefreshParams{
		RefreshToken: params.RefreshToken,
	})
	if err != nil {
		handler.logger.LogAttrs(request.Context(), slog.LevelWarn, "refresh rejected",
			slog.String("error", err.Error()),
		)
		writeAuthError(writer, err)
		return
	}

	sessionID := ""
	var expiresIn int64 = 86400
	if result.Session != nil {
		sessionID = result.Session.SessionID
		expiresIn = int64(time.Until(result.Session.ExpiresAt).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
	}

	handler.setAccessCookie(writer, request, sessionID, expiresIn)

	account := newUserParams(result.User)

	response.JSON(writer, http.StatusOK, ParamsResponse{
		AccessToken:  sessionID,
		RefreshToken: sessionID,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         account,
		UserLegacy:   account,
	})
}
