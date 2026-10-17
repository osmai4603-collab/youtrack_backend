package project

import (
	"context"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
)

const query = `
	SELECT p.id, p.name, p.short_name, p.archived, p.template, COALESCE(p.description, ''), COALESCE(p.organization_id, ''), p.created_by, p.leader_id, COALESCE(p.from_email, ''), COALESCE(p.reply_to_email, ''), COALESCE(p.icon_url, ''), p.starting_number, p.team_id, COALESCE(p.is_public, false), p.created_at, p.updated_at
	FROM projects p
	WHERE p.is_public = TRUE OR p.created_by = $1 OR p.leader_id = $1 OR EXISTS (
		SELECT 1 FROM group_members gm WHERE gm.group_id = p.team_id AND gm.is_team_member = TRUE AND (
			gm.user_id = $1 OR EXISTS (
				SELECT 1 FROM group_members gm2 WHERE gm2.group_id = gm.member_group_id AND gm2.user_id = $1 AND gm2.is_team_member = FALSE
			)
		)
	)
`

func (r *PostgresRepository) GetProjectByUserID(ctx context.Context, userID string, includeArchived bool) ([]*domain.Project, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, postgres.TranslateError(projectSelectOp, err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0, 16)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, postgres.TranslateError(projectSelectOp, err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(projectSelectOp, err)
	}

	logProjects(ctx, projectSelectOp, projects)
	return projects, nil
}
