package usecases

import (
	"log/slog"
	"time"

	"youtrack_backend/internal/services/rbac"
	"youtrack_backend/internal/storage"
	authusecase "youtrack_backend/internal/usecases/auth"
	boardusecase "youtrack_backend/internal/usecases/board"
	groupusecase "youtrack_backend/internal/usecases/groups"
	issueusecase "youtrack_backend/internal/usecases/issue"
	permissionsusecase "youtrack_backend/internal/usecases/permissions"
	projectusecase "youtrack_backend/internal/usecases/project"
	roleusecase "youtrack_backend/internal/usecases/roles"
	usersusecase "youtrack_backend/internal/usecases/users"
)

type YouTrackUseCases struct {
	Login           *authusecase.AuthLoginUseCase
	Refresh         *authusecase.AuthRefreshUseCase
	Logout          *authusecase.AuthLogoutUseCase
	Me              *usersusecase.MeUseCase
	Permissions     *permissionsusecase.UserPermissionsUseCase
	Projects        *projectusecase.ListProjectsUseCase
	ProjectByID     *projectusecase.GetProjectUseCase
	ProjectMembers  *projectusecase.ListProjectMembersUseCase
	ProjectFavorite *projectusecase.SetProjectFavoriteUseCase
	CreateProject   *projectusecase.CreateProjectUseCase
	Issues          *issueusecase.ListIssuesUseCase
	IssueByID       *issueusecase.GetIssueUseCase
	ProjectIssues   *issueusecase.ListProjectIssuesUseCase
	BoardDetails    *boardusecase.GetBoardDetailsUseCase
	MoveCard        *boardusecase.MoveCardUseCase
	Roles           *roleusecase.UseCase
	Groups          *groupusecase.UseCase
	RBAC            rbac.IRBACService
}

func (u *YouTrackUseCases) SetRBAC(svc rbac.IRBACService) {
	u.RBAC = svc
	if u.Projects != nil {
		u.Projects.SetAuthorizer(svc)
	}
}

func New(
	repos *storage.YouTrackRepositories,
	sessionTTL time.Duration,
	logger *slog.Logger,
	groups projectusecase.GroupsFinder,
	users projectusecase.UsersFinder,
	roles projectusecase.MemberRolesFinder,
	rbac rbac.IRBACService,
) *YouTrackUseCases {
	if logger == nil {
		logger = slog.Default()
	}

	return &YouTrackUseCases{
		Login:          authusecase.NewLogin(logger, repos.Users, repos.Sessions, repos.Permissions, sessionTTL),
		Refresh:        authusecase.NewRefresh(logger, repos.Users, repos.Sessions, repos.Permissions, sessionTTL),
		Logout:         authusecase.NewLogout(repos.Sessions),
		Me:             usersusecase.NewMe(logger, repos.Users),
		Permissions:    permissionsusecase.NewUserPermissions(logger, repos.Permissions),
		Roles:          roleusecase.New(repos.Permissions),
		Groups:         groupusecase.New(repos.Groups),
		Projects:       projectusecase.NewListProjects(logger, repos.Projects),
		ProjectByID:    projectusecase.NewGetProject(logger, repos.Projects, rbac),
		ProjectMembers: projectusecase.NewListProjectMembers(logger, repos.Projects, groups, users, roles, rbac),
		//ProjectFavorite: projectusecase.NewSetProjectFavorite(logger, repos.Projects),
		CreateProject: projectusecase.NewCreateProject(logger, repos.Projects),
		Issues:        issueusecase.NewListIssues(logger, repos.Issues),
		IssueByID:     issueusecase.NewGetIssue(logger, repos.Issues),
		ProjectIssues: issueusecase.NewListProjectIssues(logger, repos.Issues),
		BoardDetails:  boardusecase.NewGetBoardDetails(logger, repos.Board),
		MoveCard:      boardusecase.NewMoveCard(logger, repos.Board),
		RBAC:          rbac,
	}
}
