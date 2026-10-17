package project

import (
	"context"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
)

const listAllProjectsQuery = `
SELECT p.id, p.name, p.short_name, p.archived, p.template, COALESCE(p.description, ''),
       COALESCE(p.organization_id, ''), COALESCE(p.created_by, ''), COALESCE(p.leader_id, ''),
       COALESCE(p.from_email, ''), COALESCE(p.reply_to_email, ''), COALESCE(p.icon_url, ''),
       p.starting_number, p.team_id, COALESCE(p.is_public, false), p.created_at, p.updated_at
FROM projects p
WHERE ($1 = TRUE OR p.archived = FALSE)
ORDER BY p.name ASC
`

// ListAll returns all projects stored in PostgreSQL without membership restrictions.
func (r *PostgresRepository) ListAll(ctx context.Context, includeArchived bool) ([]*domain.Project, error) {
	const op = "projects.ListAll"

	rows, err := r.db.Query(ctx, listAllProjectsQuery, includeArchived)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0, 16)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	// requestlogger.LogStorageResult(ctx, op, len(projects))
	logProjects(ctx, op, projects)
	return projects, nil
}

const listPublicProjectsQuery = `
SELECT p.id, p.name, p.short_name, p.archived, p.template, COALESCE(p.description, ''),
       COALESCE(p.organization_id, ''), COALESCE(p.created_by, ''), COALESCE(p.leader_id, ''),
       COALESCE(p.from_email, ''), COALESCE(p.reply_to_email, ''), COALESCE(p.icon_url, ''),
       p.starting_number, p.team_id, COALESCE(p.is_public, false), p.created_at, p.updated_at
FROM projects p
WHERE p.is_public = TRUE AND ($1 = TRUE OR p.archived = FALSE)
ORDER BY p.name ASC
`

// ListPublic returns all projects flagged as public (is_public = TRUE).
func (r *PostgresRepository) ListPublic(ctx context.Context, includeArchived bool) ([]*domain.Project, error) {
	const op = "projects.ListPublic"

	rows, err := r.db.Query(ctx, listPublicProjectsQuery, includeArchived)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0, 16)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	logProjects(ctx, op, projects)
	return projects, nil
}
