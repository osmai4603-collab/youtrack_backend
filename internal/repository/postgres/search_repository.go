package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"youtrack_backend/internal/domain"
)

type SearchRepository struct {
	db *sql.DB
}

func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SearchAssist(query string) (*domain.SearchAssist, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is empty")
	}

	suggestions := []*domain.Suggestion{}

	// استعلام التذاكر المطابقة في قاعدة البيانات
	rows, err := r.db.Query(`
		SELECT id_readable, summary 
		FROM issues 
		WHERE summary ILIKE $1 OR id_readable ILIKE $1 
		LIMIT 5
	`, "%"+query+"%")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var idReadable, summary string
			if scanErr := rows.Scan(&idReadable, &summary); scanErr == nil {
				suggestions = append(suggestions, &domain.Suggestion{
					Option:      idReadable,
					Description: summary,
				})
			}
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("failed during rows iteration: %w", err)
		}
	}

	result := &domain.SearchAssist{
		Caret:          len(query),
		Query:          query,
		Type:           "SearchAssist",
		AST:            nil,
		SortProperties: nil,
		StyleRanges:    []*domain.StyleRange{},
		Suggestions:    suggestions,
		QueryFeatures:  analyzeQuery(query),
	}

	return result, nil
}

func analyzeQuery(query string) *domain.QueryFeatures {
	features := &domain.QueryFeatures{
		Length:      len(query),
		QueryLength: len(query),
	}

	if strings.ContainsAny(query, "*?") {
		features.ContainsWildcard = true
	}

	words := strings.Fields(query)
	features.NumberOfWords = len(words)

	return features
}
