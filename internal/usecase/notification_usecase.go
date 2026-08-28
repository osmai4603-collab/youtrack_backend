package usecase

import (
	"youtrack_backend/internal/domain"
)

type NotificationUseCase struct {
	notifRepo domain.NotificationRepository
}

func NewNotificationUseCase(notifRepo domain.NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{
		notifRepo: notifRepo,
	}
}

func (uc *NotificationUseCase) GetInboxFolders() ([]*domain.InboxFolder, error) {
	if uc.notifRepo != nil {
		folders, err := uc.notifRepo.GetInboxFolders()
		if err == nil && len(folders) > 0 {
			return folders, nil
		}
	}

	return []*domain.InboxFolder{
		{
			ID:      "direct",
			Enabled: true,
			Type:    "InboxFolder",
		},
		{
			ID:      "subscription",
			Enabled: true,
			Type:    "InboxFolder",
		},
		{
			ID:      "system",
			Enabled: true,
			Type:    "InboxFolder",
		},
		{
			ID:      "whatsnew",
			Enabled: true,
			Type:    "InboxFolder",
		},
		{
			ID:      "version_deploy",
			Enabled: false,
			Type:    "InboxFolder",
		},
	}, nil
}

func (uc *NotificationUseCase) GetThreads(folderID string) ([]*domain.InboxThread, error) {
	if uc.notifRepo != nil {
		threads, err := uc.notifRepo.GetThreads(folderID)
		if err == nil && len(threads) > 0 {
			return threads, nil
		}
	}

	return []*domain.InboxThread{}, nil
}
