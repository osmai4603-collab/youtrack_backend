package api

import (
	"net/http"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// ConfigResponse يمثّل مخطط استجابة إعدادات النظام.
type ConfigResponse struct {
	Config *model.Config `json:"config"`
}

// ConfigHandler يعالج طلبات الإعدادات.
type ConfigHandler struct {
	app *app.YouTrackApp
}

func NewConfigHandler(a *app.YouTrackApp) *ConfigHandler {
	return &ConfigHandler{app: a}
}

func (api *API) InitConfig() {
	handler := NewConfigHandler(api.newApp())

	api.BaseRoutes.APIRoot.Handle("/config", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Get(w, r)
	})).Methods("GET")
}

// Get يعيد إعدادات النظام العامة مباشرة في جذر الـ JSON (مطابق لـ request3.txt و request60.txt)،
// مع دعم الكائن banners المنفصل عبر المخطط الجديد المستقل (مطابق لـ request11.txt).
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	cfg := h.app.GetConfig(c.AppContext)
	fieldTree := c.FieldsTree
	result := configToMap(cfg, fieldTree)

	// دعم request11.txt: جلب كائن banners مع الجلب الانتقائي حسب شجرة الحقول
	// (banners(globalBanner,globalBannerEnabled,systemEventsBanners)).
	if fieldTree == nil || fieldTree.IsEmpty() || fieldTree.Has("banners") {
		banners := h.app.GetBannersConfig(c.AppContext, fieldTree)
		result["banners"] = bannersToMap(banners, fieldTree.Child("banners"))
	}

	writeJSON(w, http.StatusOK, result)
}

// bannersToMap يحوّل إعدادات اللافتات إلى خريطة تحترم معامل fields
// وتضمّن $type "BannersConfig" كما في استجابة YouTrack الأصلية (request11.txt).
func bannersToMap(b *model.BannersConfig, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if b == nil {
		b = model.DefaultBannersConfig()
	}
	if tree == nil || tree.IsEmpty() {
		result["globalBanner"] = b.GlobalBanner
		result["globalBannerEnabled"] = b.GlobalBannerEnabled
		result["systemEventsBanners"] = systemEventsBannersToMap(b.SystemEventsBanners)
	} else {
		if tree.Has("globalBanner") {
			result["globalBanner"] = b.GlobalBanner
		}
		if tree.Has("globalBannerEnabled") {
			result["globalBannerEnabled"] = b.GlobalBannerEnabled
		}
		if tree.Has("systemEventsBanners") {
			result["systemEventsBanners"] = systemEventsBannersToMap(b.SystemEventsBanners)
		}
	}
	result["$type"] = "BannersConfig"
	return result
}

// systemEventsBannersToMap يحوّل قائمة لافتات أحداث النظام إلى مصفوفة خرائط.
func systemEventsBannersToMap(events []*model.SystemEventsBanner) []map[string]any {
	if events == nil {
		return []map[string]any{}
	}
	result := make([]map[string]any, 0, len(events))
	for _, e := range events {
		result = append(result, systemEventsBannerToMap(e))
	}
	return result
}

// systemEventsBannerToMap يحوّل لافتة حدث نظام واحدة، وتضمّن $type "SystemEventsBanner".
func systemEventsBannerToMap(e *model.SystemEventsBanner) map[string]any {
	result := make(map[string]any)
	if e == nil {
		e = &model.SystemEventsBanner{}
	}
	result["id"] = e.ID
	if e.Text != "" {
		result["text"] = e.Text
	}
	result["$type"] = "SystemEventsBanner"
	return result
}

// configToMap يحوّل إعدادات النظام إلى خريطة مسطّحة في الجذر، مع احترام معامل fields.
// يُضاف دائمًا حقل $type بصيغة "FrontendConfig" كما في استجابة YouTrack الأصلية.
func configToMap(cfg *model.Config, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if tree == nil || tree.IsEmpty() {
		// لا يوجد filter: نُعيد كل الحقول.
		result["version"] = cfg.Version
		result["build"] = cfg.Build
		result["releaseDate"] = cfg.ReleaseDate
		result["defaultPage"] = cfg.DefaultPage
		result["contextPath"] = cfg.ContextPath
		result["readOnly"] = cfg.ReadOnly
		result["statisticsEnabled"] = cfg.StatisticsEnabled
		result["helpdeskEnabled"] = cfg.HelpdeskEnabled
		result["maxUploadFileSize"] = cfg.MaxUploadFileSize
		result["maxExportItems"] = cfg.MaxExportItems
		result["ssePingTimeoutMs"] = cfg.SSEPingTimeoutMs
		result["logoUrl"] = cfg.LogoURL
		result["wideLogoUrl"] = cfg.WideLogoURL
		result["squareLogoUrl"] = cfg.SquareLogoURL
	} else {
		if tree.Has("version") {
			result["version"] = cfg.Version
		}
		if tree.Has("build") {
			result["build"] = cfg.Build
		}
		if tree.Has("releaseDate") {
			result["releaseDate"] = cfg.ReleaseDate
		}
		if tree.Has("defaultPage") {
			result["defaultPage"] = cfg.DefaultPage
		}
		if tree.Has("contextPath") {
			result["contextPath"] = cfg.ContextPath
		}
		if tree.Has("readOnly") {
			result["readOnly"] = cfg.ReadOnly
		}
		if tree.Has("statisticsEnabled") {
			result["statisticsEnabled"] = cfg.StatisticsEnabled
		}
		if tree.Has("helpdeskEnabled") {
			result["helpdeskEnabled"] = cfg.HelpdeskEnabled
		}
		if tree.Has("maxUploadFileSize") {
			result["maxUploadFileSize"] = cfg.MaxUploadFileSize
		}
		if tree.Has("maxExportItems") {
			result["maxExportItems"] = cfg.MaxExportItems
		}
		if tree.Has("ssePingTimeoutMs") {
			result["ssePingTimeoutMs"] = cfg.SSEPingTimeoutMs
		}
		if tree.Has("logoUrl") {
			result["logoUrl"] = cfg.LogoURL
		}
		if tree.Has("wideLogoUrl") {
			result["wideLogoUrl"] = cfg.WideLogoURL
		}
		if tree.Has("squareLogoUrl") {
			result["squareLogoUrl"] = cfg.SquareLogoURL
		}
	}
	result["$type"] = "FrontendConfig"
	return result
}
