package httphandlers

import (
	"log/slog"
	authandler "youtrack_backend/internal/httphandlers/handles/auth"
	groupshandler "youtrack_backend/internal/httphandlers/handles/groups"
	issueshandler "youtrack_backend/internal/httphandlers/handles/issues"
	projectshandler "youtrack_backend/internal/httphandlers/handles/projects"
	roleshandler "youtrack_backend/internal/httphandlers/handles/roles"
	usershandler "youtrack_backend/internal/httphandlers/handles/users"
	"youtrack_backend/internal/usecases"
)

type YouTrackHandlers struct {
	Auth     *authandler.Handler
	Users    *usershandler.Handler
	Projects *projectshandler.Handler
	Issues   *issueshandler.Handler
	Roles    *roleshandler.Handler
	Groups   *groupshandler.Handler
}

func NewYouTrackHandlers(usecases *usecases.YouTrackUseCases, logger *slog.Logger) *YouTrackHandlers {
	return &YouTrackHandlers{
		Auth:  authandler.NewHandler(logger, usecases.Login, usecases.Refresh, usecases.Logout),
		Users: usershandler.NewHandler(logger, usecases.Me, usecases.Permissions),
		Projects: projectshandler.NewHandler(logger, usecases.Projects, usecases.ProjectByID,
			usecases.ProjectMembers, usecases.ProjectFavorite, usecases.BoardDetails,
			usecases.ProjectIssues, usecases.CreateProject),
		Issues: issueshandler.NewHandler(logger, usecases.Issues, usecases.IssueByID,
			usecases.ProjectIssues, usecases.MoveCard),
		Roles:  roleshandler.NewHandler(logger, usecases.Roles),
		Groups: groupshandler.NewHandler(logger, usecases.Groups),
	}
}
