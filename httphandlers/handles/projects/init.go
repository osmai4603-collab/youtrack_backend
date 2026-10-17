package projectshandler

import (
	"log/slog"
	"net/http"

	"youtrack_backend/internal/httphandlers/middlewares"
	perms2 "youtrack_backend/internal/services/permissions_services"
	boardusecase "youtrack_backend/internal/usecases/board"
	issueusecase "youtrack_backend/internal/usecases/issue"
	projectusecase "youtrack_backend/internal/usecases/project"

	"github.com/go-chi/chi/v5"
)

// Handler wires the projects HTTP endpoints to their use cases.
type Handler struct {
	logger         *slog.Logger
	projects       *projectusecase.ListProjectsUseCase
	projectByID    *projectusecase.GetProjectUseCase
	projectMembers *projectusecase.ListProjectMembersUseCase
	favorites      *projectusecase.SetProjectFavoriteUseCase
	boardDetails   *boardusecase.GetBoardDetailsUseCase
	projectIssues  *issueusecase.ListProjectIssuesUseCase
	createProject  *projectusecase.CreateProjectUseCase
}

// NewHandler builds the projects HTTP handler.
func NewHandler(
	logger *slog.Logger,
	projects *projectusecase.ListProjectsUseCase,
	projectByID *projectusecase.GetProjectUseCase,
	projectMembers *projectusecase.ListProjectMembersUseCase,
	favorites *projectusecase.SetProjectFavoriteUseCase,
	boardDetails *boardusecase.GetBoardDetailsUseCase,
	projectIssues *issueusecase.ListProjectIssuesUseCase,
	createProject *projectusecase.CreateProjectUseCase,
) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:         logger,
		projects:       projects,
		projectByID:    projectByID,
		projectMembers: projectMembers,
		favorites:      favorites,
		boardDetails:   boardDetails,
		projectIssues:  projectIssues,
		createProject:  createProject,
	}
}

func NewRoute(parent chi.Router, handler *Handler, protector middlewares.Protector, svc middlewares.AuthorizerService) *chi.Router {
	projectRequired := func(permission perms2.PermKey, hlrFunc http.HandlerFunc) http.HandlerFunc {
		return middlewares.RequireProject(svc, permission, hlrFunc, middlewares.ParamOrQuery("projectID"))
	}

	router := parent.Route("/api/projects", func(pr chi.Router) {
		pr.Group(func(g chi.Router) {
			g.Use(protector)
			g.Get("/", handler.handleProjects)
			g.Get("/{projectID}", projectRequired(perms2.PermReadProjectBasic, handler.handleProjectByID))
			g.Get("/{projectID}/members", projectRequired(perms2.PermReadProjectBasic, handler.handleProjectMembers))
			g.Get("/{projectID}/issues", projectRequired(perms2.PermReadIssue, handler.handleProjectIssues))
			g.Get("/{projectID}/board", projectRequired(perms2.PermReadProjectBasic, handler.HandleGetBoard))
			g.Put("/{projectID}/favorite", projectRequired(perms2.PermReadProjectBasic, handler.handleProjectFavorite))
			g.Post("/", middlewares.RequireGlobal(svc, perms2.PermCreateProject, handler.handleCreateProject))
		})
	})
	return &router
}
