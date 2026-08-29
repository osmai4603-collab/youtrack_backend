package app

import (
	"context"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetConfig يعيد إعدادات النظام العامة. إن لم توجد إعدادات مخزنة، يطبّق القيم الافتراضية.
func (a *App) GetConfig(ctx context.Context) *model.Config {
	cfg := &model.Config{}
	if gs, err := a.store.Admin().GlobalSettings(ctx); err == nil && gs != nil {
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

// GetGlobalSettings يعيد الإعدادات المخزنة (إن وُجدت) أو قيمًا افتراضية.
func (a *App) GetGlobalSettings(ctx context.Context) *model.GlobalSettings {
	if gs, err := a.store.Admin().GlobalSettings(ctx); err == nil && gs != nil {
		return gs
	}
	gs := &model.GlobalSettings{Version: "2025.1", Build: "0", ReadOnly: false}
	return gs
}

// GetAdminGlobalSettings يعيد إعدادات GET /api/admin/globalSettings عبر المخطط الجديد
// المستقل، مع الجلب الانتقائي حسب شجرة الحقول، أو القيم الافتراضية إن لم يوجد صف محفوظ.
func (a *App) GetAdminGlobalSettings(ctx context.Context, tree *fields.FieldTree) *model.AdminGlobalSettings {
	if gs, err := a.store.Admin().AdminGlobalSettings(ctx, tree); err == nil && gs != nil {
		return gs
	}
	return model.DefaultAdminGlobalSettings()
}

// GetBannersConfig يعيد إعدادات اللافتات لنقطة GET /api/config عبر المخطط الجديد
// المستقل BannersConfig، مع الجلب الانتقائي حسب شجرة الحقول، أو القيم الافتراضية
// إن لم يوجد صف محفوظ (مطابق لـ request11.txt).
func (a *App) GetBannersConfig(ctx context.Context, tree *fields.FieldTree) *model.BannersConfig {
	if bc, err := a.store.Admin().BannersConfig(ctx, tree); err == nil && bc != nil {
		return bc
	}
	return model.DefaultBannersConfig()
}
