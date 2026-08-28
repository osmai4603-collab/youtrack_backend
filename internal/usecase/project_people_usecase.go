package usecase

import (
	"youtrack_backend/internal/domain"
)

type ProjectPeopleUseCase struct {
	ppRepo domain.ProjectPeopleRepository
}

func NewProjectPeopleUseCase(ppRepo domain.ProjectPeopleRepository) *ProjectPeopleUseCase {
	return &ProjectPeopleUseCase{
		ppRepo: ppRepo,
	}
}

func (uc *ProjectPeopleUseCase) GetProjectPeople(projectID string, transitiveRolesQuery string) (*domain.ProjectPeople, error) {
	if uc.ppRepo != nil {
		people, err := uc.ppRepo.GetProjectPeople(projectID, transitiveRolesQuery)
		if err == nil && people != nil {
			return people, nil
		}
	}

	return &domain.ProjectPeople{
		Type: "ProjectPeople",
	}, nil
}

func (uc *ProjectPeopleUseCase) GetProjectDashboard(projectID string) (*domain.ProjectDashboard, error) {
	if uc.ppRepo != nil {
		dashboard, err := uc.ppRepo.GetProjectDashboard(projectID)
		if err == nil && dashboard != nil {
			return dashboard, nil
		}
	}

	return &domain.ProjectDashboard{
		Type: "ProjectDashboard",
	}, nil
}
