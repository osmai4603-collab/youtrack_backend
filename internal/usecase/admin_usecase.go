package usecase

import (
	"youtrack_backend/internal/domain"
)

type AdminUseCase struct {
	adminRepo domain.AdminRepository
}

func NewAdminUseCase(adminRepo domain.AdminRepository) *AdminUseCase {
	return &AdminUseCase{
		adminRepo: adminRepo,
	}
}

func (uc *AdminUseCase) GetWorkTimeSettings() (*domain.WorkTimeSettings, error) {
	if uc.adminRepo != nil {
		settings, err := uc.adminRepo.GetWorkTimeSettings()
		if err == nil && settings != nil {
			return settings, nil
		}
	}

	return &domain.WorkTimeSettings{
		DaysAWeek:               5,
		MinutesADay:             480,
		MinutesADayPresentation: "8h",
		FirstDayOfWeek:          1,
		WorkDays:                []int{1, 2, 3, 4, 5},
		Type:                    "WorkTimeSettings",
	}, nil
}

func (uc *AdminUseCase) GetGlobalSettings() (*domain.GlobalSettings, error) {
	if uc.adminRepo != nil {
		settings, err := uc.adminRepo.GetGlobalSettings()
		if err == nil && settings != nil {
			return settings, nil
		}
	}

	return &domain.GlobalSettings{
		RestSettings: &domain.RestSettings{
			AllowAllOrigins: true,
			AllowedOrigins:  []string{"*"},
			Type:            "RestSettings",
		},
		ImageTextRecognitionSettings: &domain.ImageTextRecognitionSettings{
			Enabled: true,
			Type:    "ImageTextRecognitionSettings",
		},
		SystemSettings: &domain.SystemSettings{
			OcrSupported: true,
			Type:         "SystemSettings",
		},
		NotificationSettings: &domain.NotificationSettings{
			EmailSettings: &domain.EmailSettings{
				IsEnabled: true,
				Type:      "EmailSettings",
			},
			Type: "NotificationSettings",
		},
		Type: "GlobalSettings",
	}, nil
}

func (uc *AdminUseCase) GetWidgets() ([]*domain.WidgetView, error) {
	if uc.adminRepo != nil {
		widgets, err := uc.adminRepo.GetWidgets()
		if err == nil && len(widgets) > 0 {
			return widgets, nil
		}
	}

	expectedHeight := 400
	expectedWidth := 600

	return []*domain.WidgetView{
		{
			ID:             "youtrack-issues-list",
			Key:            "issues-list",
			Name:           "Issues List",
			Description:    "Displays a list of issues matching a search query",
			AppID:          "youtrack-core",
			AppName:        "YouTrack Core",
			AppTitle:       "YouTrack",
			ExtensionPoint: "dashboard-widget",
			IndexPath:      "/widgets/issues-list/index.html",
			Configurable:   true,
			Collapsed:      false,
			ShowHeader:     true,
			Borderless:     false,
			ExpectedHeight: &expectedHeight,
			ExpectedWidth:  &expectedWidth,
			Type:           "WidgetView",
		},
		{
			ID:             "youtrack-quick-notes",
			Key:            "quick-notes",
			Name:           "Quick Notes",
			Description:    "Keep personal or project notes",
			AppID:          "youtrack-core",
			AppName:        "YouTrack Core",
			AppTitle:       "YouTrack",
			ExtensionPoint: "dashboard-widget",
			IndexPath:      "/widgets/quick-notes/index.html",
			Configurable:   false,
			Collapsed:      false,
			ShowHeader:     true,
			Borderless:     false,
			ExpectedHeight: &expectedHeight,
			ExpectedWidth:  &expectedWidth,
			Type:           "WidgetView",
		},
	}, nil
}

func (uc *AdminUseCase) GetRoles() ([]*domain.Role, error) {
	return []*domain.Role{
		{ID: "17-0", Name: "Observer", Type: "Role"},
		{ID: "17-1", Name: "Contributor", Type: "Role"},
		{ID: "17-2", Name: "Project Admin", Type: "Role"},
		{ID: "17-5", Name: "System Admin", Type: "Role"},
	}, nil
}

func (uc *AdminUseCase) GetIssueLinkTypes() ([]*domain.IssueLinkType, error) {
	return []*domain.IssueLinkType{
		{ID: "0", Name: "relates", SourceToTarget: "relates to", TargetToSource: "relates to", Type: "IssueLinkType"},
		{ID: "1", Name: "duplicate", SourceToTarget: "duplicates", TargetToSource: "is duplicated by", Type: "IssueLinkType"},
		{ID: "2", Name: "subtask", SourceToTarget: "is subtask of", TargetToSource: "has subtask", Type: "IssueLinkType"},
		{ID: "3", Name: "type", SourceToTarget: "has type", TargetToSource: "is type of", Type: "IssueLinkType"},
		{ID: "4", Name: "dependency", SourceToTarget: "depends on", TargetToSource: "is depended on by", Type: "IssueLinkType"},
	}, nil
}

func (uc *AdminUseCase) GetProjects() ([]*domain.Project, error) {
	return []*domain.Project{
		{ID: "0-0", Name: "Demo project", ShortName: "DEMO", Type: "Project"},
	}, nil
}
