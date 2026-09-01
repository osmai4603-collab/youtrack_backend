package app

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetConfig يعيد إعدادات النظام العامة.
func (a *YouTrackApp) GetConfig(c request.CTX) *model.Config {
	ctx := c.Context()
	cfg := &model.Config{}
	if gs, err := a.Store().Admin().GlobalSettings(ctx); err == nil && gs != nil {
		cfg.Version = gs.Version
		cfg.Build = gs.Build
		cfg.ReadOnly = gs.ReadOnly
		cfg.HelpdeskEnabled = gs.HelpdeskEnabled
		cfg.MaxUploadFileSize = gs.MaxUploadFileSize
		cfg.MaxExportItems = gs.MaxExportItems
	}
	cfg.SetDefaults()
	return cfg
}

// GetGlobalSettings يعيد الإعدادات المخزنة أو قيماً افتراضية.
func (a *YouTrackApp) GetGlobalSettings(c request.CTX) *model.GlobalSettings {
	ctx := c.Context()
	if gs, err := a.Store().Admin().GlobalSettings(ctx); err == nil && gs != nil {
		return gs
	}
	gs := &model.GlobalSettings{Version: "2025.1", Build: "0", ReadOnly: false}
	return gs
}

// GetAdminGlobalSettings يعيد إعدادات GET /api/admin/globalSettings.
func (a *YouTrackApp) GetAdminGlobalSettings(c request.CTX, tree *fields.FieldTree) *model.AdminGlobalSettings {
	ctx := c.Context()
	if gs, err := a.Store().Admin().AdminGlobalSettings(ctx, tree); err == nil && gs != nil {
		return gs
	}
	return model.DefaultAdminGlobalSettings()
}

// GetBannersConfig يعيد إعدادات اللافتات لنقطة GET /api/config.
func (a *YouTrackApp) GetBannersConfig(c request.CTX, tree *fields.FieldTree) *model.BannersConfig {
	ctx := c.Context()
	if bc, err := a.Store().Admin().BannersConfig(ctx, tree); err == nil && bc != nil {
		return bc
	}
	return model.DefaultBannersConfig()
}
