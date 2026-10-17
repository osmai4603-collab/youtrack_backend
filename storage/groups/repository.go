package groups

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound        = errors.New("group not found")
	ErrIDRequired      = errors.New("group ID is required")
	ErrNameRequired    = errors.New("group name is required")
	ErrSuccessorNeeded = errors.New("a different successor group is required")
	ErrSystemGroup     = errors.New("system groups cannot be deleted")
)

type Repository interface {
	List(ctx context.Context) ([]domain.Group, error)
	Create(ctx context.Context, group *domain.Group) (*domain.Group, error)
	Get(ctx context.Context, id string) (*domain.Group, error)
	Update(ctx context.Context, groupID string, name, description *string) (*domain.Group, error)
	Delete(ctx context.Context, groupID, successorID string) error
}

type PostgresRepository struct {
	db postgres.DBTX
}

func NewRepository(db postgres.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, group *domain.Group) (*domain.Group, error) {
	if group == nil || group.Name == "" {
		return nil, ErrNameRequired
	}
	groupType := group.GroupType
	if groupType == "" {
		groupType = "GROUP"
	}
	if groupType != "GROUP" && groupType != "TEAM" {
		return nil, fmt.Errorf("unsupported group type %q", groupType)
	}

	var groupID string
	err := r.db.QueryRow(ctx, `
		INSERT INTO groups (name, description, group_type, parent_group_id, auto_join)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5)
		RETURNING id
	`, group.Name, group.Description, groupType, group.ParentGroupID, group.AutoJoin).Scan(&groupID)
	if err != nil {
		return nil, postgres.TranslateError("groups.Create", err)
	}
	return r.Get(ctx, groupID)
}

const groupSelect = `
	SELECT id, name, description, COALESCE(group_type, 'GROUP'), parent_group_id, users_count, icon,
	       all_users_group, registered_users_group, auto_join, created_at, updated_at
	FROM groups
	WHERE id = $1
`

func (r *PostgresRepository) Get(ctx context.Context, id string) (*domain.Group, error) {
	if id == "" {
		return nil, ErrIDRequired
	}

	var group domain.Group
	if err := r.db.QueryRow(ctx, groupSelect, id).Scan(
		&group.ID, &group.Name, &group.Description, &group.GroupType,
		&group.ParentGroupID, &group.UsersCount, &group.Icon,
		&group.AllUsersGroup, &group.RegisteredUsersGroup, &group.AutoJoin,
		&group.CreatedAt, &group.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, postgres.TranslateError("groups.Get", err)
	}

	users, err := r.getUsers(ctx, id)
	if err != nil {
		return nil, err
	}
	group.Users = users
	group.UsersCount = int64(len(users))
	requestlogger.LogStorageResult(ctx, "groups.Get", map[string]any{
		"id":          group.ID,
		"users_count": group.UsersCount,
	})
	return &group, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]domain.Group, error) {
	const op = "groups.List"
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, COALESCE(group_type, 'GROUP'), parent_group_id, users_count, icon,
		       all_users_group, registered_users_group, auto_join, created_at, updated_at
		FROM groups
		ORDER BY name, id
	`)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	groups := make([]domain.Group, 0)
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.GroupType,
			&group.ParentGroupID, &group.UsersCount, &group.Icon,
			&group.AllUsersGroup, &group.RegisteredUsersGroup, &group.AutoJoin,
			&group.CreatedAt, &group.UpdatedAt,
		); err != nil {
			rows.Close()
			return nil, postgres.TranslateError(op, err)
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, postgres.TranslateError(op, err)
	}
	rows.Close()

	for i := range groups {
		users, err := r.getUsers(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Users = users
		groups[i].UsersCount = int64(len(users))
	}
	logGroups(ctx, op, groups)
	return groups, nil
}

func (r *PostgresRepository) getUsers(ctx context.Context, groupID string) ([]domain.User, error) {
	const query = `
		WITH RECURSIVE group_tree(id) AS (
			SELECT id FROM groups WHERE id = $1
			UNION
			SELECT gm.member_group_id
			FROM group_members gm
			JOIN group_tree gt ON gt.id = gm.group_id
			WHERE gm.member_type = 'GROUP' AND gm.member_group_id IS NOT NULL
		),
		group_user_ids(user_id) AS (
			SELECT DISTINCT gm.user_id
			FROM group_members gm
			JOIN group_tree gt ON gt.id = gm.group_id
			WHERE gm.member_type = 'USER' AND gm.user_id IS NOT NULL
		)
		SELECT u.id, u.login, COALESCE(u.email, ''), COALESCE(u.full_name, ''),
		       COALESCE(u.name, ''), COALESCE(u.avatar_url, ''),
		       COALESCE(u.type::text, ''), COALESCE(u.guest, false),
		       COALESCE(u.online, false), COALESCE(u.banned, false),
		       COALESCE(u.ring_id, ''), COALESCE(u.is_locked, false),
		       COALESCE(u.created_at, u.creation_time, 'epoch'::timestamptz)
		FROM users u
		JOIN group_user_ids gm ON gm.user_id = u.id
		ORDER BY u.login
	`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, postgres.TranslateError("groups.GetUsers", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Login, &user.Email, &user.FullName, &user.Name,
			&user.AvatarURL, &user.UserTypeID, &user.Guest, &user.Online,
			&user.Banned, &user.RingID, &user.IsLocked, &user.CreatedAt,
		); err != nil {
			return nil, postgres.TranslateError("groups.GetUsers", err)
		}
		if user.Name == "" {
			user.Name = user.FullName
		}
		if user.FullName == "" {
			user.FullName = user.Name
		}
		user.CanReadProfile = true
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError("groups.GetUsers", err)
	}
	return users, nil
}

func (r *PostgresRepository) Update(ctx context.Context, groupID string, name, description *string) (*domain.Group, error) {
	if groupID == "" {
		return nil, ErrIDRequired
	}
	if name != nil && *name == "" {
		return nil, ErrNameRequired
	}
	if name == nil && description == nil {
		return r.Get(ctx, groupID)
	}

	result, err := r.db.Exec(ctx, `
		UPDATE groups
		SET name = COALESCE($2, name),
		    description = CASE WHEN $3 THEN NULLIF($4, '') ELSE description END,
		    updated_at = $5
		WHERE id = $1 AND all_users_group = FALSE AND registered_users_group = FALSE
	`, groupID, name, description != nil, description, time.Now().UTC())
	if err != nil {
		return nil, postgres.TranslateError("groups.Update", err)
	}
	if result.RowsAffected() == 0 {
		return nil, r.groupUpdateNotFoundOrSystem(ctx, groupID)
	}
	return r.Get(ctx, groupID)
}

func (r *PostgresRepository) groupUpdateNotFoundOrSystem(ctx context.Context, groupID string) error {
	var systemGroup bool
	if err := r.db.QueryRow(ctx,
		"SELECT all_users_group OR registered_users_group FROM groups WHERE id = $1", groupID,
	).Scan(&systemGroup); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return postgres.TranslateError("groups.Update", err)
	}
	if systemGroup {
		return ErrSystemGroup
	}
	return fmt.Errorf("group update failed for %q", groupID)
}

type transactionBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

func (r *PostgresRepository) Delete(ctx context.Context, groupID, successorID string) error {
	if groupID == "" {
		return ErrIDRequired
	}
	if successorID == "" || successorID == groupID {
		return ErrSuccessorNeeded
	}

	beginner, ok := r.db.(transactionBeginner)
	if !ok {
		return errors.New("group deletion requires a transactional database connection")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var allUsers, registeredUsers bool
	if err := tx.QueryRow(ctx,
		"SELECT all_users_group, registered_users_group FROM groups WHERE id = $1", groupID,
	).Scan(&allUsers, &registeredUsers); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return postgres.TranslateError("groups.Delete", err)
	}
	if allUsers || registeredUsers {
		return ErrSystemGroup
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, member_type, is_team_member)
		SELECT $2, user_id, 'USER', is_team_member
		FROM group_members
		WHERE group_id = $1 AND member_type = 'USER'
		ON CONFLICT (group_id, user_id) WHERE user_id IS NOT NULL DO NOTHING
	`, groupID, successorID); err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO group_members (group_id, member_group_id, member_type, is_team_member)
		SELECT $2, member_group_id, 'GROUP', TRUE
		FROM group_members
		WHERE group_id = $1 AND member_type = 'GROUP'
		ON CONFLICT (group_id, member_group_id) WHERE member_group_id IS NOT NULL DO NOTHING
	`, groupID, successorID); err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO group_members (group_id, member_group_id, member_type, is_team_member)
		SELECT group_id, $2, 'GROUP', TRUE
		FROM group_members
		WHERE member_group_id = $1 AND member_type = 'GROUP' AND group_id <> $2
		ON CONFLICT (group_id, member_group_id) WHERE member_group_id IS NOT NULL DO NOTHING
	`, groupID, successorID); err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	if _, err := tx.Exec(ctx, "UPDATE assigned_roles SET group_id = $2 WHERE group_id = $1", groupID, successorID); err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	if _, err := tx.Exec(ctx, "UPDATE groups SET parent_group_id = $2 WHERE parent_group_id = $1", groupID, successorID); err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}

	result, err := tx.Exec(ctx, "DELETE FROM groups WHERE id = $1", groupID)
	if err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return postgres.TranslateError("groups.Delete", err)
	}
	requestlogger.LogStorageResult(ctx, "groups.Delete", map[string]any{"deleted": true})
	return nil
}

func logGroups(ctx context.Context, op string, groups []domain.Group, args ...slog.Attr) {
	if len(groups) == 0 {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
			//slog.String("issues", "[]"),
			args...,
		)
		return
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op, args...)
	for index, gp := range groups {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "  "+strconv.Itoa((index+1)),
			slog.String("id", gp.ID),
			slog.String("name", gp.Name),
			slog.String("type", gp.GroupType),
		)
	}
}
