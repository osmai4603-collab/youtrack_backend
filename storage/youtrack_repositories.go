package storage

import (
	"youtrack_backend/internal/infrastructure/postgres"
	authstorage "youtrack_backend/internal/storage/auth"
	boardstorage "youtrack_backend/internal/storage/board"
	groupstorage "youtrack_backend/internal/storage/groups"
	issuerepository "youtrack_backend/internal/storage/issue"
	permstorage "youtrack_backend/internal/storage/permissions"
	"youtrack_backend/internal/storage/project"
	"youtrack_backend/internal/storage/users"
)

type YouTrackRepositories struct {
	Permissions permstorage.Repository
	Users       users.Repository
	Projects    project.Repository
	Issues      issuerepository.Repository
	Board       boardstorage.Repository
	Groups      groupstorage.Repository
	Tokens      authstorage.TokenDenylist
	Sessions    authstorage.SessionRepository
}

func NewRepositoriesFromPostgres(db postgres.DBTX) *YouTrackRepositories {
	var permRepo permstorage.Repository
	var groupRepo groupstorage.Repository
	if db != nil {
		permRepo = permstorage.NewPostgresRepository(db)
		groupRepo = groupstorage.NewRepository(db)
	}
	var boardRepo boardstorage.Repository
	if db != nil {
		boardRepo = boardstorage.NewRepository(db)
	}
	return &YouTrackRepositories{
		Permissions: permRepo,
		Groups:      groupRepo,
		Users:       users.NewRepository(db),
		Projects:    project.NewRepository(db),
		Issues:      issuerepository.NewRepository(db),
		Board:       boardRepo,
		Tokens:      authstorage.NewTokenDenylist(db),
		Sessions:    authstorage.NewSessionRepository(db),
	}
}
