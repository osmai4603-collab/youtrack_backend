package project

import (
	"context"
	"time"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// addFavoriteQuery stars a project. The composite primary key makes a repeated
// star request a no-op instead of a duplicate row.
const addFavoriteQuery = `
	INSERT INTO project_favorites (user_id, project_id, created_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id, project_id) DO NOTHING
`

// removeFavoriteQuery clears a star. Deleting an absent favorite affects zero
// rows, which is the intended outcome rather than an error.
const removeFavoriteQuery = `
	DELETE FROM project_favorites
	WHERE user_id = $1 AND project_id = $2
`

// SetFavorite records or clears the identity's favorite state for a project.
func (r *PostgresRepository) SetFavorite(ctx context.Context, userID, projectID string, favorite bool) error {
	if userID == "" {
		return ErrUserIDRequired
	}
	if projectID == "" {
		return ErrProjectIDRequired
	}

	if !favorite {
		if _, err := r.db.Exec(ctx, removeFavoriteQuery, userID, projectID); err != nil {
			return postgres.TranslateError(projectFavoriteOp, err)
		}
		requestlogger.LogStorageResult(ctx, projectFavoriteOp, map[string]any{"favorite": false})
		return nil
	}

	if _, err := r.db.Exec(ctx, addFavoriteQuery, userID, projectID, time.Now().UnixMilli()); err != nil {
		return postgres.TranslateError(projectFavoriteOp, err)
	}
	requestlogger.LogStorageResult(ctx, projectFavoriteOp, map[string]any{"favorite": true})
	return nil
}
