package groupshandler

import (
	"log/slog"

	"youtrack_backend/internal/httphandlers/middlewares"
	perms2 "youtrack_backend/internal/services/permissions_services"
	groupsusecase "youtrack_backend/internal/usecases/groups"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger *slog.Logger
	groups *groupsusecase.UseCase
}

func NewHandler(logger *slog.Logger, groups *groupsusecase.UseCase) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{logger: logger, groups: groups}
}

func NewRoute(parent chi.Router, handler *Handler, protector middlewares.Protector, svc middlewares.AuthorizerService) *chi.Router {

	router := parent.Route("/api/groups", func(api chi.Router) {
		api.Use(protector)
		api.Post("/", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleCreate))
		api.Get("/", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleList))
		api.Get("/{groupID}", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleGet))
		api.Post("/{groupID}", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleUpdate))
		api.Delete("/{groupID}", middlewares.RequireGlobal(svc, perms2.PermLowLevelAdminWrite, handler.handleDelete))
	})
	return &router
}
