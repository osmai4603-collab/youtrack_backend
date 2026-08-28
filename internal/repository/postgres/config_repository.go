package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"youtrack_backend/internal/domain"
)

type ConfigRepository struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) GetFrontendConfig() (*domain.FrontendConfig, error) {
	var (
		allowOrigins              bool
		imageTextRecognition      sql.NullBool
		ocrSupported              sql.NullString
		emailEnabled              sql.NullBool
		emailIsDefault            sql.NullBool
		version                   sql.NullString
		build                     sql.NullString
		releaseDate               sql.NullInt64
		defaultPage               sql.NullString
		contextPath               sql.NullString
		readOnly                  bool
		statisticsEnabled         bool
		helpdeskEnabled           sql.NullBool
		maxUploadFileSize         sql.NullInt64
		maxExportItems            sql.NullInt32
		ssePingTimeoutMs          sql.NullInt64
	)

	err := r.db.QueryRow(`
		SELECT
			allow_all_origins,
			image_text_recognition_enabled,
			ocr_supported,
			email_settings_enabled,
			email_settings_is_default,
			version,
			build,
			release_date,
			default_page,
			context_path,
			read_only,
			statistics_enabled,
			helpdesk_enabled,
			max_upload_file_size,
			max_export_items,
			sse_ping_timeout_ms
		FROM global_settings
		WHERE id = 1
	`).Scan(
		&allowOrigins,
		&imageTextRecognition,
		&ocrSupported,
		&emailEnabled,
		&emailIsDefault,
		&version,
		&build,
		&releaseDate,
		&defaultPage,
		&contextPath,
		&readOnly,
		&statisticsEnabled,
		&helpdeskEnabled,
		&maxUploadFileSize,
		&maxExportItems,
		&ssePingTimeoutMs,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no global settings found: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("failed to query global settings: %w", err)
	}

	config := &domain.FrontendConfig{
		Type:              "jetbrainsundleserver.frontend.configs.FrontendConfig",
		ReadOnly:          readOnly,
		StatisticsEnabled: statisticsEnabled,
		Banners: &domain.BannersConfig{
			Type: "jetbrains.bundle.server.frontend.configs.BannersConfig",
		},
	}

	if defaultPage.Valid {
		config.DefaultPage = defaultPage.String
	} else {
		config.DefaultPage = "/issues"
	}

	if contextPath.Valid {
		config.ContextPath = contextPath.String
	}

	if version.Valid {
		config.Version = version.String
	}

	if build.Valid {
		config.Build = build.String
	}

	if releaseDate.Valid {
		config.ReleaseDate = releaseDate.Int64
	}

	if helpdeskEnabled.Valid {
		config.HelpdeskEnabled = helpdeskEnabled.Bool
	}

	config.System = &domain.SystemFrontendConfig{
		Type: "jetbrains.bundle.server.frontend.configs.SystemFrontendConfig",
	}
	if maxUploadFileSize.Valid {
		config.System.MaxUploadFileSize = int(maxUploadFileSize.Int64)
	}
	if maxExportItems.Valid {
		config.System.MaxExportItems = int(maxExportItems.Int32)
	}
	if ssePingTimeoutMs.Valid {
		config.System.SsePingTimeoutMs = int(ssePingTimeoutMs.Int64)
	}

	config.Ring = &domain.RingFrontendConfig{
		Type: "jetbrains.bundle.server.frontend.configs.RingFrontendConfig",
	}

	config.L10n = &domain.L10NFrontendConfig{
		Type:              "jetbrains.bundle.server.frontend.configs.L10NFrontendConfig",
		PredefinedQueries: make(map[string]string),
	}

	config.Hosted = &domain.HostedFrontendConfig{
		Type: "jetbrains.bundle.server.frontend.configs.HostedFrontendConfig",
	}

	config.Konnector = &domain.KonnectorConfig{
		Type: "jetbrains.bundle.server.frontend.configs.KonnectorConfig",
	}

	config.Shortcuts = &domain.ShortcutScheme{
		ID:   "default",
		Type: "jetbrains.bundle.server.frontend.configs.ShortcutScheme",
	}

	config.FeaturesURL = "/api/features"
	config.MarketplaceBaseURL = "https://plugins.jetbrains.com"

	return config, nil
}

func (r *ConfigRepository) GetPublicSettings() (*domain.PublicSettings, error) {
	return &domain.PublicSettings{
		ID:   "0",
		Type: "jetbrains.bundle.server.frontend.configs.PublicSettings",
	}, nil
}
