package project

import (
	"context"
	"fmt"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// createProjectQuery inserts a project. `pinned` is intentionally absent: it is
// the legacy global flag and is no longer read, since favorites are per identity.
const createProjectQuery = `
	INSERT INTO projects (
		id, name, short_name, archived, template, description, organization_id,
		created_by, leader_id, from_email, reply_to_email, icon_url,
		starting_number, team_id, created_at, updated_at
	)
	VALUES (
		$1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''),
		NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''),
		NULLIF($12, ''), $13, NULLIF($14, ''), $15, $16
	)
`

// updateProjectQuery rewrites every mutable field of a project.
//
// This is a full write rather than a partial patch on purpose: with a partial
// update the store cannot tell a caller that means "clear the description" from
// one that forgot the field, so writing "" and false would be impossible.
// Callers read, merge and write the whole record.
//
// id, created_by, created_at, and the favorite flag are excluded: the first
// three are immutable, and favorites are per identity rather than a column.
const updateProjectQuery = `
	UPDATE projects
	SET name           = $2,
		short_name      = $3,
		archived        = $4,
		template        = $5,
		description     = NULLIF($6, ''),
		organization_id = NULLIF($7, ''),
		leader_id       = NULLIF($8, ''),
		from_email      = NULLIF($9, ''),
		reply_to_email  = NULLIF($10, ''),
		icon_url        = NULLIF($11, ''),
		starting_number = $12,
		team_id         = NULLIF($13, ''),
		updated_at      = $14
	WHERE id = $1
`

const deleteProjectQuery = `DELETE FROM projects WHERE id = $1`

// Create inserts a new project.
func (r *PostgresRepository) Create(ctx context.Context, project *domain.Project) error {
	if project == nil || project.Name == "" {
		return ErrProjectNameRequired
	}
	if project.ID == "" {
		return ErrProjectIDRequired
	}

	now := time.Now().UTC()
	createdAt := now
	if project.CreatedAt != nil {
		createdAt = *project.CreatedAt
	}
	startingNumber := project.StartingNumber
	if startingNumber == 0 {
		startingNumber = 1
	}

	_, err := r.db.Exec(ctx, createProjectQuery,
		project.ID,
		project.Name,
		project.ShortName,
		project.Archived,
		project.Template,
		project.Description,
		project.OrganizationID,
		project.CreatedBy,
		project.LeaderID,
		project.FromEmail,
		project.ReplyToEmail,
		project.IconURL,
		startingNumber,
		project.TeamID,
		createdAt,
		now,
	)
	if err != nil {
		// A reused ID or short name is a caller mistake, not an outage, so it is
		// reported as a conflict instead of the raw translated database error.
		if postgres.IsUniqueViolation(err) {
			return fmt.Errorf("%s: %w", projectWriteOp, ErrConflict)
		}
		return postgres.TranslateError(projectWriteOp, err)
	}
	requestlogger.LogStorageResult(ctx, "projects.Create", project)
	return nil
}

// Update rewrites the mutable fields of an existing project.
func (r *PostgresRepository) Update(ctx context.Context, project *domain.Project) error {
	if project == nil || project.Name == "" {
		return ErrProjectNameRequired
	}
	if project.ID == "" {
		return ErrProjectIDRequired
	}

	startingNumber := project.StartingNumber
	if startingNumber == 0 {
		startingNumber = 1
	}

	result, err := r.db.Exec(ctx, updateProjectQuery,
		project.ID,
		project.Name,
		project.ShortName,
		project.Archived,
		project.Template,
		project.Description,
		project.OrganizationID,
		project.LeaderID,
		project.FromEmail,
		project.ReplyToEmail,
		project.IconURL,
		startingNumber,
		project.TeamID,
		time.Now().UTC(),
	)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return fmt.Errorf("%s: %w", projectWriteOp, ErrConflict)
		}
		return postgres.TranslateError(projectWriteOp, err)
	}
	// No row means the project is gone. Reporting it as not visible keeps an
	// update from revealing whether the id ever existed.
	if result.RowsAffected() == 0 {
		return ErrProjectNotVisible
	}
	requestlogger.LogStorageResult(ctx, "projects.Update", project)
	return nil
}

// Delete removes a project; dependents go with it via ON DELETE CASCADE.
func (r *PostgresRepository) Delete(ctx context.Context, projectID string) error {
	if projectID == "" {
		return ErrProjectIDRequired
	}

	result, err := r.db.Exec(ctx, deleteProjectQuery, projectID)
	if err != nil {
		return postgres.TranslateError(projectWriteOp, err)
	}
	if result.RowsAffected() == 0 {
		return ErrProjectNotVisible
	}
	requestlogger.LogStorageResult(ctx, "projects.Delete", map[string]any{"deleted": true})
	return nil
}
