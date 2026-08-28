package usecase

import (
	"youtrack_backend/internal/domain"
)

type AgileUseCase struct {
	agileRepo domain.AgileRepository
}

func NewAgileUseCase(agileRepo domain.AgileRepository) *AgileUseCase {
	return &AgileUseCase{
		agileRepo: agileRepo,
	}
}

func (uc *AgileUseCase) GetUserProfile() (*domain.AgileUserProfile, error) {
	if uc.agileRepo != nil {
		profile, err := uc.agileRepo.GetUserProfile()
		if err == nil && profile != nil {
			return profile, nil
		}
	}

	return &domain.AgileUserProfile{
		CardDetailLevel: 2,
		Type:            "AgileUserProfile",
	}, nil
}

func (uc *AgileUseCase) GetBoardExtensions(boardID string) (*domain.Extensions, error) {
	if uc.agileRepo != nil {
		extensions, err := uc.agileRepo.GetBoardExtensions(boardID)
		if err == nil && extensions != nil {
			return extensions, nil
		}
	}

	return nil, nil
}
