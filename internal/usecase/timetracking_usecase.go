package usecase

import "youtrack_backend/internal/domain"

type TimeTrackingUseCase struct {
	ttRepo domain.TimeTrackingRepository
}

func NewTimeTrackingUseCase(ttRepo domain.TimeTrackingRepository) *TimeTrackingUseCase {
	return &TimeTrackingUseCase{ttRepo: ttRepo}
}

func (uc *TimeTrackingUseCase) GetAttributePrototypes() ([]*domain.AttributePrototype, error) {
	if uc.ttRepo != nil {
		prototypes, err := uc.ttRepo.GetAttributePrototypes()
		if err == nil && prototypes != nil {
			return prototypes, nil
		}
	}
	return []*domain.AttributePrototype{}, nil
}

func (uc *TimeTrackingUseCase) GetAttributePrototype(id string) (*domain.AttributePrototype, error) {
	if uc.ttRepo != nil {
		prototype, err := uc.ttRepo.GetAttributePrototype(id)
		if err == nil && prototype != nil {
			return prototype, nil
		}
	}
	return nil, nil
}

func (uc *TimeTrackingUseCase) GetBoardTimeTrackingData(boardID string) (*domain.BoardTimeTrackingData, error) {
	if uc.ttRepo != nil {
		data, err := uc.ttRepo.GetBoardTimeTrackingData(boardID)
		if err == nil && data != nil {
			return data, nil
		}
	}
	return nil, nil
}
