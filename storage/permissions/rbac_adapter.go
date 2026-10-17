package permstorage

import (
	"context"
	"errors"
	"fmt"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	"youtrack_backend/internal/services/rbac"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresRBACRoleAdapter implements both rbac.RoleRepository and rbac.RBACRoleRepository
// backed by the PostgreSQL 'roles' table.
type PostgresRBACRoleAdapter struct {
	db postgres.DBTX
}

// NewPostgresRBACRoleAdapter constructs a new PostgreSQL adapter for the RBAC role engine.
func NewPostgresRBACRoleAdapter(db postgres.DBTX) *PostgresRBACRoleAdapter {
	return &PostgresRBACRoleAdapter{db: db}
}

// Ensure interface compliances at compile time.
var (
	_ rbac.RoleRepository     = (*PostgresRBACRoleAdapter)(nil)
	_ rbac.RBACRoleRepository = (*PostgresRBACRoleAdapter)(nil)
)

// GetRole retrieves a role by ID, returning rbac.ErrRoleNotFound if not found.
func (a *PostgresRBACRoleAdapter) GetRole(ctx context.Context, id string) (*rbac.Role, error) {
	const op = "rbac.PostgresRoleAdapter.GetRole"
	if a.db == nil {
		return nil, fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	query := `
		SELECT id, name, COALESCE(description, ''), immutable, permissions
		FROM roles
		WHERE id = $1
	`

	var role rbac.Role
	var permissions []string
	err := a.db.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsImmutable, &permissions,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", rbac.ErrRoleNotFound, id)
		}
		return nil, postgres.TranslateError(op, err)
	}

	role.Permissions = permissions
	if role.Permissions == nil {
		role.Permissions = []string{}
	}
	role.Metadata = make(map[string]string)

	requestlogger.LogStorageResult(ctx, op+"(id: "+id+")", role.Name)
	return &role, nil
}

// GetRolesWhereIDs fetches multiple roles matching the given IDs.
func (a *PostgresRBACRoleAdapter) GetRolesWhereIDs(ctx context.Context, ids []string) ([]*rbac.Role, error) {
	const op = "rbac.PostgresRoleAdapter.GetRolesWhereIDs"
	if a.db == nil {
		return nil, fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}
	if len(ids) == 0 {
		return []*rbac.Role{}, nil
	}

	query := `
		SELECT id, name, COALESCE(description, ''), immutable, permissions
		FROM roles
		WHERE id = ANY($1)
		ORDER BY id
	`

	rows, err := a.db.Query(ctx, query, ids)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	var result []*rbac.Role
	for rows.Next() {
		var role rbac.Role
		var permissions []string
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsImmutable, &permissions); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		role.Permissions = permissions
		if role.Permissions == nil {
			role.Permissions = []string{}
		}
		role.Metadata = make(map[string]string)
		result = append(result, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	return result, nil
}

// GetAllRoles retrieves all roles stored in PostgreSQL.
func (a *PostgresRBACRoleAdapter) GetAllRoles(ctx context.Context) ([]*rbac.Role, error) {
	const op = "rbac.PostgresRoleAdapter.GetAllRoles"
	if a.db == nil {
		return nil, fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	query := `
		SELECT id, name, COALESCE(description, ''), immutable, permissions
		FROM roles
		ORDER BY id
	`

	rows, err := a.db.Query(ctx, query)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	var result []*rbac.Role
	for rows.Next() {
		var role rbac.Role
		var permissions []string
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsImmutable, &permissions); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		role.Permissions = permissions
		if role.Permissions == nil {
			role.Permissions = []string{}
		}
		role.Metadata = make(map[string]string)
		result = append(result, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	return result, nil
}

// CreateRole inserts a new role and returns the persisted record.
func (a *PostgresRBACRoleAdapter) CreateRole(ctx context.Context, role *rbac.Role) (*rbac.Role, error) {
	const op = "rbac.PostgresRoleAdapter.CreateRole"
	if role == nil || role.Name == "" {
		return nil, fmt.Errorf("%w: role and role name must not be empty", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return nil, fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	perms := role.Permissions
	if perms == nil {
		perms = []string{}
	}

	query := `
		INSERT INTO roles (id, name, description, immutable, permissions)
		VALUES (COALESCE(NULLIF($1, ''), gen_random_uuid()::text), $2, NULLIF($3, ''), $4, $5)
		RETURNING id, name, COALESCE(description, ''), immutable, permissions
	`

	var persisted rbac.Role
	var savedPerms []string
	err := a.db.QueryRow(ctx, query,
		role.ID, role.Name, role.Description, role.IsImmutable, perms,
	).Scan(&persisted.ID, &persisted.Name, &persisted.Description, &persisted.IsImmutable, &savedPerms)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: role already exists: %s", rbac.ErrRoleAlreadyExists, role.Name)
		}
		return nil, postgres.TranslateError(op, err)
	}

	persisted.Permissions = savedPerms
	persisted.Metadata = role.Metadata
	return &persisted, nil
}

// SaveRole satisfies rbac.RBACRoleRepository.
func (a *PostgresRBACRoleAdapter) SaveRole(ctx context.Context, role *rbac.Role) error {
	_, err := a.CreateRole(ctx, role)
	return err
}

// UpdateRole updates an existing role in PostgreSQL.
func (a *PostgresRBACRoleAdapter) UpdateRole(ctx context.Context, role *rbac.Role) error {
	const op = "rbac.PostgresRoleAdapter.UpdateRole"
	if role == nil || role.ID == "" {
		return fmt.Errorf("%w: role ID is required", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	perms := role.Permissions
	if perms == nil {
		perms = []string{}
	}

	query := `
		UPDATE roles
		SET name = $2,
		    description = NULLIF($3, ''),
		    permissions = $4
		WHERE id = $1 AND immutable = FALSE
	`

	res, err := a.db.Exec(ctx, query, role.ID, role.Name, role.Description, perms)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	if res.RowsAffected() == 0 {
		// Determine whether role is immutable or doesn't exist
		var immutable bool
		checkErr := a.db.QueryRow(ctx, "SELECT immutable FROM roles WHERE id = $1", role.ID).Scan(&immutable)
		if checkErr != nil {
			if errors.Is(checkErr, pgx.ErrNoRows) {
				return fmt.Errorf("%w: %s", rbac.ErrRoleNotFound, role.ID)
			}
			return postgres.TranslateError(op, checkErr)
		}
		if immutable {
			return fmt.Errorf("%w: role %s is immutable", rbac.ErrRoleImmutable, role.ID)
		}
	}

	return nil
}

// DeleteRole deletes a role by ID. Returns error if failed.
func (a *PostgresRBACRoleAdapter) DeleteRole(ctx context.Context, roleID string) error {
	const op = "rbac.PostgresRoleAdapter.DeleteRole"
	if roleID == "" {
		return fmt.Errorf("%w: role ID is required", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	res, err := a.db.Exec(ctx, "DELETE FROM roles WHERE id = $1 AND immutable = FALSE", roleID)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	if res.RowsAffected() == 0 {
		var immutable bool
		checkErr := a.db.QueryRow(ctx, "SELECT immutable FROM roles WHERE id = $1", roleID).Scan(&immutable)
		if checkErr != nil {
			if errors.Is(checkErr, pgx.ErrNoRows) {
				return fmt.Errorf("%w: %s", rbac.ErrRoleNotFound, roleID)
			}
			return postgres.TranslateError(op, checkErr)
		}
		if immutable {
			return fmt.Errorf("%w: role %s is immutable", rbac.ErrRoleImmutable, roleID)
		}
		return nil
	}

	return nil
}
