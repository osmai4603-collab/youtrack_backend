package users

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"

	"github.com/jackc/pgx/v5"
)

func (r *InMemoryRepository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	r.mu.RLock()
	user, ok := r.users[login]
	if !ok {
		r.mu.RUnlock()
		return nil, ErrUserNotFound
	}
	copyUser := *user
	r.mu.RUnlock()
	requestlogger.LogStorageResult(ctx, "users.FindByLogin(login: "+login+")", "email: "+copyUser.Email+", user_id: "+copyUser.ID+", login: "+copyUser.Login)
	return &copyUser, nil
}

func (r *PostgresRepository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, ErrUserNotFound
	}
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE login = $1`
	user, err := scanUser(r.db.QueryRow(ctx, query, login))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, postgres.TranslateError("users.FindByLogin("+login+")", err)
	}
	logUsers(ctx, "FindUserByLogin", []*domain.User{user}, slog.String("login", login))
	return user, nil
}
