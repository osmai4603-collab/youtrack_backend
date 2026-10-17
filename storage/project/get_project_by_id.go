package project

import (
	"context"
	"errors"
	"log/slog"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"

	"github.com/jackc/pgx/v5"
)

// getVisibleProjectToUserQuery selects the project table columns in scanProject
// order and appends the caller's per-user favorite flag.
//
// Archived projects are not filtered out — a project hidden from the default
// listing is still addressable by id, matching IsVisibleToUser.
const getVisibleProjectToUserQuery = `
SELECT
	p.id,
	p.name,
	p.short_name,
	COALESCE(p.archived, false),
	COALESCE(p.template, false),
	COALESCE(p.description, ''),
	COALESCE(p.organization_id, ''),
	COALESCE(p.created_by, ''),
	COALESCE(p.leader_id, ''),
	COALESCE(p.from_email, ''),
	COALESCE(p.reply_to_email, ''),
	COALESCE(p.icon_url, ''),
	COALESCE(p.starting_number, 1),
	COALESCE(p.team_id, ''),
	COALESCE(p.is_public, false),
	p.created_at,
	p.updated_at
FROM projects p
WHERE p.id = $1 LIMIT 1
`

func (r *PostgresRepository) GetProjectByID(ctx context.Context, projectID string) (*domain.Project, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}

	project, err := scanProject(r.db.QueryRow(ctx, getVisibleProjectToUserQuery, projectID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotVisible
		}
		return nil, postgres.TranslateError(projectGetOp, err)
	}
	logProjects(ctx, projectGetOp, []*domain.Project{project}, slog.String("projectID", projectID))
	return project, nil
}
