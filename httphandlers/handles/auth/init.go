package authandler

import (
	"log/slog"
	"youtrack_backend/internal/httphandlers/middlewares"
	authusecase "youtrack_backend/internal/usecases/auth"

	"github.com/go-chi/chi/v5"
)

func NewRoute(router chi.Router, handler *Handler, protect middlewares.Protector) *chi.Router {
	router.Route("/api/auth", func(r chi.Router) {
		r.Post("/login", handler.handleLogin)
		r.Post("/refresh", handler.handleRefresh)

		r.Group(func(pr chi.Router) {
			pr.Use(protect)
			pr.Post("/logout", handler.handleLogout)
		})
	})
	return &router
}

const defaultCookieName = middlewares.DefaultCookieName

type Handler struct {
	logger     *slog.Logger
	auth       *authusecase.AuthLoginUseCase
	refresh    *authusecase.AuthRefreshUseCase
	logout     *authusecase.AuthLogoutUseCase
	cookieName string
}

func NewHandler(logger *slog.Logger, login *authusecase.AuthLoginUseCase, refresh *authusecase.AuthRefreshUseCase, logout *authusecase.AuthLogoutUseCase) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:     logger,
		auth:       login,
		refresh:    refresh,
		logout:     logout,
		cookieName: defaultCookieName,
	}
}
