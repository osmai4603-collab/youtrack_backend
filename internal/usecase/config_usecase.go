package usecase

import (
	"youtrack_backend/internal/domain"
)

type ConfigUseCase struct {
	configRepo domain.ConfigRepository
}

func NewConfigUseCase(configRepo domain.ConfigRepository) *ConfigUseCase {
	return &ConfigUseCase{configRepo: configRepo}
}

func (uc *ConfigUseCase) GetFrontendConfig() (*domain.FrontendConfig, error) {
	if uc.configRepo != nil {
		cfg, err := uc.configRepo.GetFrontendConfig()
		if err == nil && cfg != nil {
			return cfg, nil
		}
	}

	return &domain.FrontendConfig{
		Ring: &domain.RingFrontendConfig{
			Enabled:        true,
			HasEmbeddedHub: true,
			URL:            "/hub",
			Type:           "RingFrontendConfig",
		},
		L10n: &domain.L10NFrontendConfig{
			IsRTL:   false,
			Locale:  "en-US",
			Language: "en",
			PredefinedQueries: map[string]string{
				"Unresolved": "Unresolved",
				"Open":       "Open",
				"State":      "State",
			},
			Type: "L10NFrontendConfig",
		},
		System: &domain.SystemFrontendConfig{
			MaxUploadFileSize: 10485760,
			MaxExportItems:    500,
			SsePingTimeoutMs:  300000,
			Type:              "SystemFrontendConfig",
		},
		Shortcuts: &domain.ShortcutScheme{
			ID:   "DEFAULT",
			Type: "ShortcutScheme",
		},
		Hosted: &domain.HostedFrontendConfig{
			Hosted:          true,
			Domain:          "osm",
			AvailabilityZone: "eu",
			Type:            "HostedFrontendConfig",
		},
		Konnector: &domain.KonnectorConfig{
			TelegramBotURL: "https://t.me/jetbrains_youtrack_bot",
			URL:            "https://konnector.services.jetbrains.com",
			Type:           "KonnectorConfig",
		},
		StatisticsEnabled:    true,
		ContextPath:          "",
		DefaultPage:          "/dashboard",
		Version:              "2026.2",
		FeaturesURL:          "https://resources.jetbrains.com/youtrack/features",
		MarketplaceBaseURL:   "https://plugins.jetbrains.com",
		RedirectToWelcomeForm: false,
		HelpdeskEnabled:      false,
		ReadOnly:             false,
		Banners: &domain.BannersConfig{
			GlobalBannerEnabled: false,
			GlobalBanner:        "",
			SystemEventsBanners: []string{},
			Type:                "BannersConfig",
		},
		Type: "FrontendConfig",
	}, nil
}

func (uc *ConfigUseCase) GetPermissionsCache() []*domain.CachedPermission {
	defaultProjects := []*domain.PermissionProject{
		{
			ID: "22-195",
			ProjectType: &domain.ProjectType{
				ID:   "DEFAULT",
				Type: "ProjectType",
			},
			Type: "Project",
		},
		{
			ID: "22-1522",
			ProjectType: &domain.ProjectType{
				ID:   "HELPDESK",
				Type: "ProjectType",
			},
			Type: "Project",
		},
	}

	return []*domain.CachedPermission{
		{
			ID:            "jetbrains.jetpass.project-read-basic",
			Global:        false,
			Projects:      defaultProjects,
			Organizations: []*domain.Organization{},
			Type:          "CachedPermission",
		},
		{
			ID:            "jetbrains.jetpass.profile-updateSelf",
			Global:        true,
			Projects:      nil,
			Organizations: nil,
			Type:          "CachedPermission",
		},
		{
			ID:            "jetbrains.jetpass.user-read-basic",
			Global:        true,
			Projects:      nil,
			Organizations: nil,
			Type:          "CachedPermission",
		},
		{
			ID:            "JetBrains.YouTrack.READ_ISSUE",
			Global:        false,
			Projects:      defaultProjects,
			Organizations: []*domain.Organization{},
			Type:          "CachedPermission",
		},
		{
			ID:            "JetBrains.YouTrack.CREATE_ISSUE",
			Global:        false,
			Projects:      defaultProjects,
			Organizations: []*domain.Organization{},
			Type:          "CachedPermission",
		},
		{
			ID:            "JetBrains.YouTrack.CREATE_COMMENT",
			Global:        false,
			Projects:      defaultProjects,
			Organizations: []*domain.Organization{},
			Type:          "CachedPermission",
		},
		{
			ID:            "JetBrains.YouTrack.READ_COMMENT",
			Global:        false,
			Projects:      defaultProjects,
			Organizations: []*domain.Organization{},
			Type:          "CachedPermission",
		},
	}
}
