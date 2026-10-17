package usershandler

import (
	"log/slog"
	"net/http"
	"youtrack_backend/internal/httphandlers/middlewares"
	perms2 "youtrack_backend/internal/services/permissions_services"
	permissionsusecase "youtrack_backend/internal/usecases/permissions"
	usersusecase "youtrack_backend/internal/usecases/users"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger      *slog.Logger
	users       *usersusecase.MeUseCase
	permissions *permissionsusecase.UserPermissionsUseCase
}

func NewHandler(logger *slog.Logger, users *usersusecase.MeUseCase, permissions *permissionsusecase.UserPermissionsUseCase) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:      logger,
		users:       users,
		permissions: permissions,
	}
}

func NewRoute(parent chi.Router, handler *Handler, protector middlewares.Protector, svc middlewares.AuthorizerService) *chi.Router {
	permissionRequired := func(permission perms2.PermKey, hlrFunc http.HandlerFunc) http.HandlerFunc {
		return middlewares.RequireGlobal(svc, perms2.PermReadUserDetails, hlrFunc)
	}

	router := parent.Route("/api/users", func(api chi.Router) {
		api.Use(protector)
		api.Get("/me", permissionRequired(perms2.PermReadUserDetails, handler.handleMe))
		api.Get("/me/permissions", handler.handleUserPermissions)

	})
	return &router
}
