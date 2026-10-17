package authandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/httphandlers/response"
	authusecase "youtrack_backend/internal/usecases/auth"
)

type ParamsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ParamsResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
	User         userParams `json:"user"`
	UserLegacy   userParams `json:"user_id"`
}

type userParams struct {
	Guest           bool   `json:"guest"`
	Banned          bool   `json:"banned"`
	Login           string `json:"login"`
	Email           string `json:"email"`
	AvatarURL       string `json:"avatarUrl"`
	Online          bool   `json:"online"`
	IsEmailVerified bool   `json:"isEmailVerified"`
	BanBadge        string `json:"banBadge"`
	CanReadProfile  bool   `json:"canReadProfile"`
	FullName        string `json:"fullName"`
	Name            string `json:"name"`
	IsLocked        bool   `json:"isLocked"`
	ID              string `json:"id"`
	Type            string `json:"$type"`
	CreatedAt       string `json:"created_at"`
	UserName        string `json:"user_name"`
	FullNameAlias   string `json:"full_name"`
	AvatarURLAlias  string `json:"avatar_url"`
	IsBanned        bool   `json:"is_banned"`
}

func newUserParams(user *domain.User) userParams {
	if user == nil {
		return userParams{}
	}
	return userParams{
		Guest:           user.Guest,
		Banned:          user.Banned,
		Login:           user.Login,
		Email:           user.Email,
		AvatarURL:       user.AvatarURL,
		Online:          user.Online,
		IsEmailVerified: user.IsEmailVerified,
		BanBadge:        user.BanBadge,
		CanReadProfile:  user.CanReadProfile,
		FullName:        user.FullName,
		Name:            user.Name,
		IsLocked:        user.IsLocked,
		ID:              user.ID,
		Type:            user.UserTypeID,
		CreatedAt:       formatCreatedAt(user.CreatedAt),
		UserName:        user.Login,
		FullNameAlias:   user.FullName,
		AvatarURLAlias:  user.AvatarURL,
		IsBanned:        user.Banned,
	}
}

func formatCreatedAt(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func (hlr *Handler) handleLogin(writer http.ResponseWriter, request *http.Request) {
	var params ParamsRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
		http.Error(writer, "invalid request body", http.StatusUnprocessableEntity)
		return
	}
	if strings.TrimSpace(params.Username) == "" || params.Password == "" {
		http.Error(writer, "invalid request body", http.StatusUnprocessableEntity)
		return
	}

	result, err := hlr.auth.TryLogin(request.Context(), &authusecase.LoginParams{
		Username: strings.ToLower(strings.TrimSpace(params.Username)),
		Password: params.Password,
	})
	if err != nil {
		hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, "login rejected",
			slog.String("username", params.Username),
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

	hlr.setAccessCookie(writer, request, sessionID, expiresIn)

	account := newUserParams(result.User)

	response.JSON(writer, http.StatusOK, ParamsResponse{
		AccessToken:  sessionID,
		RefreshToken: sessionID,
		TokenType:    "Cookie",
		ExpiresIn:    expiresIn,
		User:         account,
		UserLegacy:   account,
	})
}

func (handler *Handler) setAccessCookie(writer http.ResponseWriter, request *http.Request, value string, maxAge int64) {
	http.SetCookie(writer, &http.Cookie{
		Name:     handler.cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge),
		HttpOnly: true,
		Secure:   request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

func (handler *Handler) clearCookie(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name:     handler.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeAuthError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authusecase.ErrAccountLocked), errors.Is(err, authusecase.ErrAccountInactive):
		http.Error(writer, "account temporarily unavailable", http.StatusLocked)
	case errors.Is(err, authusecase.ErrInvalidCredentials), errors.Is(err, authusecase.ErrInvalidRefreshToken):
		http.Error(writer, "invalid credentials", http.StatusUnauthorized)
	default:
		http.Error(writer, "authentication unavailable", http.StatusInternalServerError)
	}
}
