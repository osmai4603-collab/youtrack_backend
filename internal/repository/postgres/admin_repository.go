package postgres

import (
	"database/sql"
	"fmt"
	"youtrack_backend/internal/domain"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetWorkTimeSettings() (*domain.WorkTimeSettings, error) {
	query := `
		SELECT days_a_week, minutes_a_day, minutes_a_day_presentation, first_day_of_week
		FROM work_time_settings
		LIMIT 1
	`
	settings := &domain.WorkTimeSettings{
		WorkDays: []int{1, 2, 3, 4, 5},
	}

	err := r.db.QueryRow(query).Scan(
		&settings.DaysAWeek,
		&settings.MinutesADay,
		&settings.MinutesADayPresentation,
		&settings.FirstDayOfWeek,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &domain.WorkTimeSettings{
				DaysAWeek:               5,
				MinutesADay:             480,
				MinutesADayPresentation: "8h",
				FirstDayOfWeek:          1,
				WorkDays:                []int{1, 2, 3, 4, 5},
				Type:                    "WorkTimeSettings",
			}, nil
		}
		return nil, fmt.Errorf("failed to get work time settings: %w", err)
	}

	settings.Type = "WorkTimeSettings"
	return settings, nil
}

func (r *AdminRepository) GetGlobalSettings() (*domain.GlobalSettings, error) {
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

func (r *AdminRepository) GetWidgets() ([]*domain.WidgetView, error) {
	expectedHeight := 400
	expectedWidth := 600

	widgets := []*domain.WidgetView{
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
	}

	return widgets, nil
}
