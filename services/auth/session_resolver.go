package auth

import (
	"context"
	"errors"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/services/rbac"
	authstorage "youtrack_backend/internal/storage/auth"
)

type roleFinder interface {
	GetRoleAssignmentsForUser(ctx context.Context, userID string) ([]rbac.RoleAssignment, error)
}

type sessionFinder interface {
	GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error)
}

var (
	ErrRoleStoreUnavailable = errors.New("role store unavailable")
	ErrUserIDRequired       = errors.New("user ID is required to resolve roles")
)

type CookieSessionResolver struct {
	sessionFinder
	roleFinder
}

func NewCookieSessionResolver(sessions authstorage.SessionRepository, roles roleFinder) *CookieSessionResolver {
	return &CookieSessionResolver{
		sessionFinder: sessions,
		roleFinder:    roles,
	}
}

func (r *CookieSessionResolver) GetSessionFromToken(token string, ctx context.Context) (*domain.UserSession, error) {
	session, err := r.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *CookieSessionResolver) GetRoleAssignmentsFromUserID(ctx context.Context, userID string) ([]rbac.RoleAssignment, error) {
	if r.roleFinder == nil {
		return nil, ErrRoleStoreUnavailable
	}
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	assignments, err := r.GetRoleAssignmentsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if assignments == nil {
		return []rbac.RoleAssignment{}, nil
	}

	return assignments, nil
}
