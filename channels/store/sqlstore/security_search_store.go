package sqlstore

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/channels/model"
)

// SecuritySearchStore هو تنفيذ لـ store.SecuritySearchStore باستخدام PostgreSQL.
type SecuritySearchStore struct {
	db *pgxpool.Pool
}

func (s *SecuritySearchStore) GetFilterFields(ctx context.Context, entityType string) ([]*model.SecurityFilterField, error) {
	query := `
		SELECT id, name, field_type
		FROM security_filter_fields
		WHERE entity_type = $1
	`
	rows, err := s.db.Query(ctx, query, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []*model.SecurityFilterField
	for rows.Next() {
		f := &model.SecurityFilterField{EntityType: entityType}
		if err := rows.Scan(&f.ID, &f.Name, &f.Type); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fields, nil
}
