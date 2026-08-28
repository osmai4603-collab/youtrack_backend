package postgres

import (
	"database/sql"
	"fmt"

	"youtrack_backend/internal/domain"
)

type TimeTrackingRepository struct {
	db *sql.DB
}

func NewTimeTrackingRepository(db *sql.DB) *TimeTrackingRepository {
	return &TimeTrackingRepository{db: db}
}

func (r *TimeTrackingRepository) GetAttributePrototypes() ([]*domain.AttributePrototype, error) {
	return []*domain.AttributePrototype{}, nil
}

func (r *TimeTrackingRepository) GetAttributePrototype(id string) (*domain.AttributePrototype, error) {
	return nil, nil
}

func (r *TimeTrackingRepository) GetBoardTimeTrackingData(boardID string) (*domain.BoardTimeTrackingData, error) {
	var boardName sql.NullString
	query := `SELECT name FROM agile_boards WHERE id = $1`

	err := r.db.QueryRow(query, boardID).Scan(&boardName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agile board %s not found", boardID)
		}
		return nil, fmt.Errorf("querying agile board: %w", err)
	}

	data := &domain.BoardTimeTrackingData{
		ID:   boardID,
		Type: "BoardTimeTrackingData",
	}

	if boardName.Valid {
		data.Agile = &domain.AgileBoard{
			ID:   boardID,
			Name: boardName.String,
			Type: "AgileBoard",
		}
	}

	return data, nil
}
