package permstorage

import (
	"context"
	"fmt"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"

	"github.com/jackc/pgx/v5"
)

// RevokeRole deletes a role assignment by assignment ID.
func (r *PostgresRepository) RevokeRole(ctx context.Context, assignmentID string) error {
	const op = "permissions.RevokeRole"

	if assignmentID == "" {
		return fmt.Errorf("%s: assignment ID is required", op)
	}

	query := `DELETE FROM assigned_roles WHERE id = $1`
	res, err := r.db.Exec(ctx, query, assignmentID)
	if err != nil {
		return postgres.TranslateError(op, err)
	}

	if res.RowsAffected() == 0 {
		return postgres.TranslateError(op, pgx.ErrNoRows)
	}

	requestlogger.LogStorageResult(ctx, op, map[string]any{"revoked": true})
	return nil
}
