package httphandlers

import (
	"log/slog"
	"net/http"
	"time"
	authandler "youtrack_backend/internal/httphandlers/handles/auth"
	groupshandler "youtrack_backend/internal/httphandlers/handles/groups"
	issueshandler "youtrack_backend/internal/httphandlers/handles/issues"
	projectshandler "youtrack_backend/internal/httphandlers/handles/projects"
	roleshandler "youtrack_backend/internal/httphandlers/handles/roles"
	usershandler "youtrack_backend/internal/httphandlers/handles/users"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/infrastructure/health"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type YouTrackRouter struct {
	Base     *chi.Mux
	auth     *chi.Router
	users    *chi.Router
	projects *chi.Router
	issues   *chi.Router
	roles    *chi.Router
	groups   *chi.Router
}

func NewYouTrackRouter(handlers *YouTrackHandlers, healthChecker *health.HealthChecker, logger *slog.Logger, protector middlewares.Protector, authorizer middlewares.AuthorizerService) *YouTrackRouter {
	baseRoute := chi.NewRouter()
	// Global Middlewares
	baseRoute.Use(middleware.RequestID)
	baseRoute.Use(structuredLogger(logger))
	baseRoute.Use(middleware.Recoverer)
	// r.Use(i18n.Middleware)
	requestTimeout := 60 * time.Second
	baseRoute.Use(middleware.Timeout(requestTimeout))

	if healthChecker != nil {
		baseRoute.Get("/livez", healthChecker.HandleLiveness)
		baseRoute.Get("/readyz", healthChecker.HandleReadiness)
		baseRoute.Get("/healthz", healthChecker.HandleReadiness)
	}

	router := YouTrackRouter{
		Base:     baseRoute,
		auth:     authandler.NewRoute(baseRoute, handlers.Auth, protector),
		users:    usershandler.NewRoute(baseRoute, handlers.Users, protector, authorizer),
		projects: projectshandler.NewRoute(baseRoute, handlers.Projects, protector, authorizer),
		issues:   issueshandler.NewRoute(baseRoute, handlers.Issues, protector, authorizer),
		roles:    roleshandler.NewRoute(baseRoute, handlers.Roles, protector, authorizer),
		groups:   groupshandler.NewRoute(baseRoute, handlers.Groups, protector, authorizer),
	}

	return &router
}

func structuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
