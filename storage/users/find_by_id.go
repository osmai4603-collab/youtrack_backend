package users

import (
	"context"
	"errors"
	"log/slog"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"

	"github.com/jackc/pgx/v5"
)

func (r *InMemoryRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	login, ok := r.byID[id]
	if !ok {
		r.mu.RUnlock()
		return nil, ErrUserNotFound
	}
	user := r.users[login]
	copyUser := *user
	r.mu.RUnlock()
	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, "users.FindByID("+id+")",
		slog.String("user_id", user.ID),
		slog.String("email", user.Email),
		slog.String("name", user.Name),
	)
	return &copyUser, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, ErrUserIDRequired
	}
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE id = $1`
	user, err := scanUser(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, postgres.TranslateError("users.FindByID", err)
	}
	logUsers(ctx, "FindUserByID", []*domain.User{user}, slog.String("userID", id))
	return user, nil
}
