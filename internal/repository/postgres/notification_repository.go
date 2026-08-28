package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"youtrack_backend/internal/domain"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) GetInboxFolders() ([]*domain.InboxFolder, error) {
	query := `
		SELECT id, last_seen, last_notified, enabled
		FROM inbox_folders
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbox folders: %w", err)
	}
	defer rows.Close()

	var folders []*domain.InboxFolder
	for rows.Next() {
		folder := &domain.InboxFolder{}
		var lastSeen, lastNotified sql.NullInt64
		var enabled sql.NullBool

		err := rows.Scan(
			&folder.ID,
			&lastSeen,
			&lastNotified,
			&enabled,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to scan inbox folder: %w", err)
		}

		if lastSeen.Valid {
			folder.LastSeen = lastSeen.Int64
		}
		if lastNotified.Valid {
			folder.LastNotified = lastNotified.Int64
		}
		if enabled.Valid {
			folder.Enabled = enabled.Bool
		}

		folder.Type = "InboxFolder"
		folders = append(folders, folder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate inbox folders: %w", err)
	}

	return folders, nil
}

func (r *NotificationRepository) GetThreads(folderID string) ([]*domain.InboxThread, error) {
	// TODO: implement when inbox_threads table is created
	return []*domain.InboxThread{}, nil
}
