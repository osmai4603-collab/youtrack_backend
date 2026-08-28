package usecase

import (
	"youtrack_backend/internal/domain"
)

type VCSUseCase struct {
	vcsRepo domain.VCSRepository
}

func NewVCSUseCase(vcsRepo domain.VCSRepository) *VCSUseCase {
	return &VCSUseCase{
		vcsRepo: vcsRepo,
	}
}

func (uc *VCSUseCase) GetVCSServers() ([]*domain.VCSServer, error) {
	if uc.vcsRepo != nil {
		servers, err := uc.vcsRepo.GetVCSServers()
		if err == nil && len(servers) > 0 {
			return servers, nil
		}
	}

	return []*domain.VCSServer{
		{
			ID:           "github.com",
			IsPredefined: true,
			AppName:      "GitHub",
			Type:         "GitHub",
		},
		{
			ID:           "gitlab.com",
			IsPredefined: true,
			AppName:      "GitLab",
			Type:         "GitLab",
		},
		{
			ID:           "bitbucket.org",
			IsPredefined: true,
			AppName:      "Bitbucket",
			Type:         "Bitbucket",
		},
		{
			ID:           "azure.com",
			IsPredefined: true,
			AppName:      "Azure DevOps",
			Type:         "AzureDevOps",
		},
		{
			ID:           "generic",
			IsPredefined: true,
			AppName:      "Generic Git Server",
			Type:         "Generic",
		},
	}, nil
}

type AppUseCase struct {
	appRepo domain.AppRepository
}

func NewAppUseCase(appRepo domain.AppRepository) *AppUseCase {
	return &AppUseCase{
		appRepo: appRepo,
	}
}

func (uc *AppUseCase) GetServicesPage() (*domain.ServicesPage, error) {
	if uc.appRepo != nil {
		page, err := uc.appRepo.GetServicesPage()
		if err == nil && page != nil {
			return page, nil
		}
	}

	return &domain.ServicesPage{
		ID:   "0-0-0-0-0",
		Type: "ServicesPage",
	}, nil
}

type HubUseCase struct {
	hubRepo domain.HubRepository
}

func NewHubUseCase(hubRepo domain.HubRepository) *HubUseCase {
	return &HubUseCase{
		hubRepo: hubRepo,
	}
}

func (uc *HubUseCase) GetHubUser() (*domain.HubUser, error) {
	if uc.hubRepo != nil {
		user, err := uc.hubRepo.GetHubUser()
		if err == nil && user != nil {
			return user, nil
		}
	}

	return &domain.HubUser{
		ID:    "1-1",
		Name:  "Administrator",
		Login: "admin",
		Guest: false,
		Profile: &domain.HubProfile{
			Email: &domain.HubEmail{
				Email:    "admin@example.com",
				Verified: true,
				Type:     "HubEmail",
			},
		},
		Type: "HubUser",
	}, nil
}
