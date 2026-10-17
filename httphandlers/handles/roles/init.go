package roleshandler

import (
	"log/slog"

	"youtrack_backend/internal/httphandlers/middlewares"
	perms2 "youtrack_backend/internal/services/permissions_services"
	rolesusecase "youtrack_backend/internal/usecases/roles"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger *slog.Logger
	roles  *rolesusecase.UseCase
}

func NewHandler(logger *slog.Logger, roles *rolesusecase.UseCase) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{logger: logger, roles: roles}
}

func NewRoute(parent chi.Router, handler *Handler, protector middlewares.Protector, svc middlewares.AuthorizerService) *chi.Router {

	router := parent.Route("/api/roles", func(api chi.Router) {
		api.Use(protector)
		api.Post("/", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleCreate))
		api.Get("/", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminRead, handler.handleList))
		api.Get("/{roleID}", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminRead, handler.handleGet))
		api.Post("/{roleID}", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleUpdate))
		api.Delete("/{roleID}", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleDelete))
	})
	return &router
}
