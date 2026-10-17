package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"

	"github.com/jackc/pgx/v5"
)

const listProjectSubsystemsQuery = `
	SELECT
		id,
		project_id,
		name,
		COALESCE(owner_id, ''),
		COALESCE(color, 0),
		COALESCE(created_at, 0),
		COALESCE(updated_at, 0)
	FROM project_subsystems
	WHERE project_id = $1
	ORDER BY name ASC
`

const getProjectSubsystemQuery = `
	SELECT
		id,
		project_id,
		name,
		COALESCE(owner_id, ''),
		COALESCE(color, 0),
		COALESCE(created_at, 0),
		COALESCE(updated_at, 0)
	FROM project_subsystems
	WHERE project_id = $1 AND id = $2
`

const createProjectSubsystemQuery = `
	INSERT INTO project_subsystems (id, project_id, name, owner_id, color, created_at, updated_at)
	VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
`

// ListSubsystems returns a project's subsystems ordered by name.
func (r *PostgresRepository) ListSubsystems(ctx context.Context, projectID string) ([]*domain.ProjectSubsystem, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}

	rows, err := r.db.Query(ctx, listProjectSubsystemsQuery, projectID)
	if err != nil {
		return nil, postgres.TranslateError(projectSubsystemOp, err)
	}
	defer rows.Close()

	subsystems := make([]*domain.ProjectSubsystem, 0)
	for rows.Next() {
		subsystem, err := scanSubsystem(rows)
		if err != nil {
			return nil, err
		}
		subsystems = append(subsystems, subsystem)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(projectSubsystemOp, err)
	}
	requestlogger.LogStorageResult(ctx, "projects.ListSubsystems", subsystems)
	return subsystems, nil
}

// GetSubsystem returns one subsystem of a project.
func (r *PostgresRepository) GetSubsystem(ctx context.Context, projectID, subsystemID string) (*domain.ProjectSubsystem, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}
	if subsystemID == "" {
		return nil, ErrSubsystemNotFound
	}

	subsystem, err := scanSubsystem(r.db.QueryRow(ctx, getProjectSubsystemQuery, projectID, subsystemID))
	if err != nil {
		// The lookup is scoped to one project, so "no such row" only means the
		// subsystem is absent from this project.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubsystemNotFound
		}
		return nil, err
	}
	requestlogger.LogStorageResult(ctx, "projects.GetSubsystem", subsystem)
	return subsystem, nil
}

// CreateSubsystem adds a subsystem to a project.
func (r *PostgresRepository) CreateSubsystem(ctx context.Context, subsystem *domain.ProjectSubsystem) error {
	if subsystem == nil || subsystem.Name == "" {
		return ErrSubsystemNameRequired
	}
	if subsystem.ProjectID == "" {
		return ErrProjectIDRequired
	}
	if subsystem.ID == "" {
		return ErrSubsystemIDRequired
	}

	now := time.Now().UnixMilli()
	createdAt := subsystem.CreatedAt
	if createdAt == 0 {
		createdAt = now
	}

	_, err := r.db.Exec(ctx, createProjectSubsystemQuery,
		subsystem.ID,
		subsystem.ProjectID,
		subsystem.Name,
		subsystem.OwnerID,
		subsystem.Color,
		createdAt,
		now,
	)
	if err != nil {
		// The (project_id, name) unique index rejects a duplicate name, which is
		// a conflict the caller can act on rather than a server fault.
		if postgres.IsUniqueViolation(err) {
			return fmt.Errorf("%s: %w", projectSubsystemOp, ErrConflict)
		}
		return postgres.TranslateError(projectSubsystemOp, err)
	}
	requestlogger.LogStorageResult(ctx, "projects.CreateSubsystem", subsystem)
	return nil
}
