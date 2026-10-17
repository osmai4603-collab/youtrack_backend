package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	requestlogger "youtrack_backend/internal/platform/logger"
	"youtrack_backend/internal/services/rbac"
)

type contextKey string

const DefaultCookieName = "youtrack_session"
const keySessionID contextKey = "sessionID"
const keyUserID contextKey = "userID"

type SessionResolver interface {
	GetSessionFromToken(token string, ctx context.Context) (*domain.UserSession, error)
	GetRoleAssignmentsFromUserID(context.Context, string) ([]rbac.RoleAssignment, error)
}

type Protector func(http.Handler) http.Handler

func Authenticate(resolver SessionResolver) Protector {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {

			if resolver == nil {
				response.Error(wr, platErrors.Internal("authenticate", "session store unavailable", errors.New("nil session store")))
				return
			}

			cookie, err := req.Cookie(DefaultCookieName)
			if err == http.ErrNoCookie {
				response.Error(wr, platErrors.Unauthorized("authenticate", "session cookie required", err))
				return
			}

			if cookie.Value == "" {
				response.Error(wr, platErrors.Unauthorized("authenticate", "session cookie required", err))
				return
			}

			session, err := resolver.GetSessionFromToken(cookie.Value, req.Context())
			if err != nil {
				response.Error(wr, platErrors.Unauthorized("authenticate", "authentication required", err))
				return
			}
			if session == nil || session.UserID == "" {
				response.Error(wr, platErrors.Unauthorized("authenticate", "session has no user", err))
				return
			}
			if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(time.Now().UTC()) {
				requestlogger.FromContext(req.Context()).LogAttrs(req.Context(), slog.LevelWarn, "session expired", slog.String("userID", session.UserID), slog.String("expiredAt", session.ExpiresAt.Format("yyyy-mm-dd hh:mm:ss")))
				response.Error(wr, platErrors.Unauthorized("authenticate", "session expired", err))
				return
			}
			assignments, err := resolver.GetRoleAssignmentsFromUserID(req.Context(), session.UserID)
			if err != nil {
				response.Error(wr, platErrors.Internal("authenticate", "failed to load user roles", err))
				return
			}
			ctx := context.WithValue(req.Context(), keyUserID, session.UserID)
			ctx = context.WithValue(ctx, keySessionID, session.SessionID)
			ctx = rbac.WithCurrentUserRoleAssignments(ctx, assignments)

			next.ServeHTTP(wr, req.WithContext(ctx))

		})
	}
}

func GetCurrentUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(keyUserID).(string)
	return userID, ok
}

func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(keySessionID).(string)
	return sessionID, ok
}

func GetCurrentUserRoleAssignments(ctx context.Context) ([]rbac.RoleAssignment, bool) {
	return rbac.GetCurrentUserRoleAssignments(ctx)
}
