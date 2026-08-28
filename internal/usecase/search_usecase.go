package usecase

import (
	"youtrack_backend/internal/domain"
)

type SearchUseCase struct {
	searchRepo domain.SearchRepository
}

func NewSearchUseCase(searchRepo domain.SearchRepository) *SearchUseCase {
	return &SearchUseCase{
		searchRepo: searchRepo,
	}
}

func (uc *SearchUseCase) SearchAssist(query string) (*domain.SearchAssist, error) {
	if uc.searchRepo != nil {
		result, err := uc.searchRepo.SearchAssist(query)
		if err == nil && result != nil {
			return result, nil
		}
	}

	return &domain.SearchAssist{
		Caret: len(query),
		Query: query,
		QueryFeatures: &domain.QueryFeatures{
			Length:        len(query),
			QueryLength:   len(query),
			NumberOfWords: 1,
		},
	}, nil
}
