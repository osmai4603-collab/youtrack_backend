package project

import (
	"context"
	"encoding/json"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const listProjectTemplatesQuery = `
	SELECT
		id,
		name,
		COALESCE(description, ''),
		COALESCE(icon_key, ''),
		COALESCE(default_fields, '{}'::JSONB)
	FROM project_templates
	ORDER BY name ASC
`

// ListTemplates returns the project templates ordered by name.
func (r *PostgresRepository) ListTemplates(ctx context.Context) ([]*domain.ProjectTemplate, error) {
	rows, err := r.db.Query(ctx, listProjectTemplatesQuery)
	if err != nil {
		return nil, postgres.TranslateError(projectTemplateOp, err)
	}
	defer rows.Close()

	templates := make([]*domain.ProjectTemplate, 0)
	for rows.Next() {
		var template domain.ProjectTemplate
		rawFields := []byte(`{}`)
		if err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.IconKey,
			&rawFields,
		); err != nil {
			return nil, postgres.TranslateError(projectTemplateOp, err)
		}
		template.DefaultFields = map[string]string{}
		// A template's fields are an opaque document; an unparseable one yields
		// no defaults rather than failing the whole listing.
		_ = json.Unmarshal(rawFields, &template.DefaultFields)
		templates = append(templates, &template)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(projectTemplateOp, err)
	}
	requestlogger.LogStorageResult(ctx, projectTemplateOp, templates)
	return templates, nil
}
