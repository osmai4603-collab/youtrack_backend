package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

const opAuthorize = "authorize"

type AuthorizerService interface {
	IsAllowed(ctx context.Context, userID string, permission perms2.PermKey, scope perms2.ScopeContext) (bool, error)
}

func RequireGlobal(svc AuthorizerService, permission perms2.PermKey, next http.HandlerFunc) http.HandlerFunc {
	return RequirePermission(svc, permission, next, ScopeBinding{
		Scope:  ScopeGlobal,
		Entity: perms2.EntitySystem,
	})
}

func RequireOrganization(svc AuthorizerService, permission perms2.PermKey, next http.HandlerFunc, organizationID IDExtractor) http.HandlerFunc {
	return RequirePermission(svc, permission, next, ScopeBinding{
		Scope:   ScopeOrganization,
		Entity:  perms2.EntityOrganization,
		Extract: organizationID,
	})
}

func RequireProject(svc AuthorizerService, permission perms2.PermKey, next http.HandlerFunc, projectID IDExtractor) http.HandlerFunc {
	return RequirePermission(svc, permission, next, ScopeBinding{
		Scope:   ScopeProject,
		Entity:  perms2.EntityProject,
		Extract: projectID,
	})
}

func RequirePermission(svc AuthorizerService, permission perms2.PermKey, next http.HandlerFunc, binding ScopeBinding) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID, ok := GetCurrentUserID(req.Context())
		if !ok {
			response.Error(w, req, platErrors.Unauthorized(opAuthorize, "authentication required", errors.New("userID is not found in request")))
			return
		}

		scope := perms2.ScopeContext{Type: binding.Scope, ResourceType: binding.Entity}
		if EntityScoped(binding.Scope) {
			id, err := binding.Extract(req)
			if err != nil {
				response.Error(w, req, platErrors.Invalid(opAuthorize, fmt.Sprintf("cannot resolve %s scope for this request", binding.Scope), err))
				return
			}
			scope.ScopeID = id
		}

		isAllowed, err := svc.IsAllowed(req.Context(), userID, permission, scope)

		if err != nil {
			requestlogger.FromContext(req.Context()).LogAttrs(req.Context(), slog.LevelError, "permission check failed",
				slog.String("user_id", userID),
				slog.String("permission_id", permission),
				slog.String("scope", string(scope.Type)),
				slog.String("resource_type", string(scope.ResourceType)),
				slog.String("scope_id", scope.ScopeID),
				slog.Any("error", err),
			)
			response.Error(w, req, platErrors.Forbidden(opAuthorize, "permissions unexpected error", err))
			return
		}
		if !isAllowed {
			requestlogger.FromContext(req.Context()).LogAttrs(req.Context(), slog.LevelWarn, "permission denied",
				slog.String("user_id", userID),
				slog.String("permission_id", permission),
				slog.String("scope", string(scope.Type)),
				slog.String("resource_type", string(scope.ResourceType)),
				slog.String("scope_id", scope.ScopeID),
			)
			response.Error(w, req, platErrors.Forbidden(opAuthorize, "permissions denied", errors.New("not authorized")))
			return
		}
				

		next.ServeHTTP(w, req)
	})
}
