package postgres

import (
	"database/sql"
	"fmt"
	"youtrack_backend/internal/domain"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetByID(id string) (*domain.Project, error) {
	query := `
		SELECT id, name, short_name, description, pinned, icon_url, archived, restricted, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	project := &domain.Project{}
	var desc, iconURL sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.ShortName,
		&desc,
		&project.Pinned,
		&iconURL,
		&project.Archived,
		&project.Restricted,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if desc.Valid {
		project.Description = desc.String
	}
	if iconURL.Valid {
		project.IconURL = iconURL.String
	}
	project.Type = "Project"

	return project, nil
}

func (r *ProjectRepository) List() ([]*domain.Project, error) {
	query := `
		SELECT id, name, short_name, description, pinned, icon_url, archived, restricted, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		p := &domain.Project{}
		var desc, iconURL sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.ShortName,
			&desc,
			&p.Pinned,
			&iconURL,
			&p.Archived,
			&p.Restricted,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		if desc.Valid {
			p.Description = desc.String
		}
		if iconURL.Valid {
			p.IconURL = iconURL.String
		}
		p.Type = "Project"
		projects = append(projects, p)
	}

	return projects, nil
}

func (r *ProjectRepository) Create(project *domain.Project) error {
	query := `
		INSERT INTO projects (id, name, short_name, description, pinned, icon_url, archived, restricted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name,
		    short_name = EXCLUDED.short_name,
		    description = EXCLUDED.description,
		    pinned = EXCLUDED.pinned,
		    icon_url = EXCLUDED.icon_url,
		    archived = EXCLUDED.archived,
		    restricted = EXCLUDED.restricted,
		    updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(query,
		project.ID,
		project.Name,
		project.ShortName,
		project.Description,
		project.Pinned,
		project.IconURL,
		project.Archived,
		project.Restricted,
	)
	if err != nil {
		return fmt.Errorf("failed to create/update project: %w", err)
	}
	return nil
}
