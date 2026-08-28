package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"youtrack_backend/internal/domain"
)

type AgileRepository struct {
	db *sql.DB
}

func NewAgileRepository(db *sql.DB) *AgileRepository {
	return &AgileRepository{db: db}
}

func (r *AgileRepository) GetUserProfile() (*domain.AgileUserProfile, error) {
	query := `
		SELECT id, name, is_demo, is_updatable, created_with_original
		FROM agile_boards
		ORDER BY id
		LIMIT 1
	`

	var id, name sql.NullString
	var isDemo, isUpdatable, createdWithOriginal bool

	err := r.db.QueryRow(query).Scan(&id, &name, &isDemo, &isUpdatable, &createdWithOriginal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.AgileUserProfile{
				CardDetailLevel: 2,
				Type:            "AgileUserProfile",
			}, nil
		}
		return nil, fmt.Errorf("failed to get agile board for user profile: %w", err)
	}

	board := &domain.AgileBoard{
		ID:                  id.String,
		IsDemo:              isDemo,
		IsUpdatable:         isUpdatable,
		CreatedWithOriginal: createdWithOriginal,
		Type:                "AgileBoard",
	}
	if name.Valid {
		board.Name = name.String
	}

	profile := &domain.AgileUserProfile{
		CardDetailLevel: 2,
		DefaultAgile:    board,
		VisitedSprints:  make([]*domain.Sprint, 0),
		Type:            "AgileUserProfile",
	}

	return profile, nil
}

func (r *AgileRepository) GetBoardExtensions(boardID string) (*domain.Extensions, error) {
	query := `
		SELECT id, name, is_demo, is_updatable, created_with_original
		FROM agile_boards
		WHERE id = $1
	`

	var id, name sql.NullString
	var isDemo, isUpdatable, createdWithOriginal bool

	err := r.db.QueryRow(query, boardID).Scan(&id, &name, &isDemo, &isUpdatable, &createdWithOriginal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get board extensions: %w", err)
	}

	board := &domain.AgileBoard{
		ID:                  id.String,
		IsDemo:              isDemo,
		IsUpdatable:         isUpdatable,
		CreatedWithOriginal: createdWithOriginal,
		Type:                "AgileBoard",
	}
	if name.Valid {
		board.Name = name.String
	}

	extensions := &domain.Extensions{
		ID: boardID,
		TimeTracking: &domain.BoardTimeTrackingData{
			ID:    boardID,
			Agile: board,
			Type:  "BoardTimeTrackingData",
		},
		Type: "Extensions",
	}

	return extensions, nil
}
