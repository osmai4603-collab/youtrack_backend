package permstorage

import (
	"context"
	"errors"
	"fmt"

	"youtrack_backend/internal/infrastructure/postgres"
	platformerr "youtrack_backend/internal/platform/errors"

	"github.com/jackc/pgx/v5"
)

var ErrImmutableRole = errors.New("immutable roles cannot be changed")

func (r *PostgresRepository) CreateRole(ctx context.Context, name string, description *string) (*Role, error) {
	var roleID string
	err := r.db.QueryRow(ctx, `
		INSERT INTO roles (name, description)
		VALUES ($1, NULLIF($2, ''))
		RETURNING id
	`, name, description).Scan(&roleID)
	if err != nil {
		return nil, postgres.TranslateError("permissions.CreateRole", err)
	}
	return r.GetRoleByID(ctx, roleID)
}

func (r *PostgresRepository) UpdateRole(ctx context.Context, roleID string, name, description *string) (*Role, error) {
	if roleID == "" {
		return nil, errors.New("role ID is required")
	}
	if name != nil && *name == "" {
		return nil, errors.New("role name is required")
	}
	result, err := r.db.Exec(ctx, `
		UPDATE roles
		SET name = COALESCE($2, name),
		    description = CASE WHEN $3 THEN NULLIF($4, '') ELSE description END
		WHERE id = $1 AND immutable = FALSE
	`, roleID, name, description != nil, description)
	if err != nil {
		return nil, postgres.TranslateError("permissions.UpdateRole", err)
	}
	if result.RowsAffected() == 0 {
		return nil, r.immutableOrMissingRole(ctx, roleID)
	}
	return r.GetRoleByID(ctx, roleID)
}

func (r *PostgresRepository) DeleteRole(ctx context.Context, roleID string) error {
	if roleID == "" {
		return errors.New("role ID is required")
	}
	result, err := r.db.Exec(ctx, "DELETE FROM roles WHERE id = $1 AND immutable = FALSE", roleID)
	if err != nil {
		return postgres.TranslateError("permissions.DeleteRole", err)
	}
	if result.RowsAffected() == 0 {
		return r.immutableOrMissingRole(ctx, roleID)
	}
	return nil
}

func (r *PostgresRepository) immutableOrMissingRole(ctx context.Context, roleID string) error {
	var immutable bool
	err := r.db.QueryRow(ctx, "SELECT immutable FROM roles WHERE id = $1", roleID).Scan(&immutable)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return platformerr.NotFound("permissions.Role", "role not found", err)
		}
		return postgres.TranslateError("permissions.Role", err)
	}
	if immutable {
		return ErrImmutableRole
	}
	return fmt.Errorf("role %q was not changed", roleID)
}
