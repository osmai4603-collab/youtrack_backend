package api

import (
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// RoleListResponse يمثّل مخطط استجابة قائمة الأدوار.
type RoleListResponse struct {
	Roles []*model.Role `json:"roles"`
	Count int           `json:"count"`
}

// PermissionListResponse يمثّل مخطط استجابة قائمة الصلاحيات.
type PermissionListResponse struct {
	Permissions []*model.Permission `json:"permissions"`
	Count       int                 `json:"count"`
}

// AdminHandler يعالج طلبات الإدارة.
type AdminHandler struct {
	app *app.App
}

func NewAdminHandler(a *app.App) *AdminHandler {
	return &AdminHandler{app: a}
}

func (h *AdminHandler) Roles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.app.GetRoles(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, RoleListResponse{Roles: roles, Count: len(roles)})
}

func (h *AdminHandler) Permissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.app.GetPermissions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, PermissionListResponse{Permissions: perms, Count: len(perms)})
}

// PermissionsCache يعيد الصلاحيات المخبأة للمستخدم الحالي (مطابق لـ request7.txt).
func (h *AdminHandler) PermissionsCache(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	cache, err := h.app.GetPermissionsCache(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, cache)
}

// Widgets يعيد قائمة الودجات العامة بصيغة مصفوفة JSON (مطابق لـ request6.txt).
func (h *AdminHandler) Widgets(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.app.GetWidgets(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, widgets)
}

// GlobalSettings يعيد الإعدادات الإدارية العامة مع احترام معامل fields
// (مطابق لـ request5.txt و request19.txt و request46.txt).
func (h *AdminHandler) GlobalSettings(w http.ResponseWriter, r *http.Request) {
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	gs := h.app.GetAdminGlobalSettings(r.Context(), fieldTree)
	writeJSON(w, http.StatusOK, globalSettingsToMap(gs, fieldTree))
}

// globalSettingsToMap يحوّل الإعدادات الإدارية إلى خريطة تحترم معامل fields
// وتضمّن $type على كل مستوى كما في استجابة YouTrack الأصلية.
func globalSettingsToMap(gs *model.AdminGlobalSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if gs == nil {
		gs = model.DefaultAdminGlobalSettings()
	}

	if tree == nil || tree.IsEmpty() {
		// لا يوجد filter: نُعيد كل الإعدادات.
		if gs.RestSettings != nil {
			result["restSettings"] = restCorsSettingsToMap(gs.RestSettings, nil)
		}
		if gs.ImageTextRecognitionSettings != nil {
			result["imageTextRecognitionSettings"] = imageTextRecognitionSettingsToMap(gs.ImageTextRecognitionSettings, nil)
		}
		if gs.SystemSettings != nil {
			result["systemSettings"] = systemSettingsToMap(gs.SystemSettings, nil)
		}
		if gs.NotificationSettings != nil {
			result["notificationSettings"] = notificationSettingsToMap(gs.NotificationSettings, nil)
		}
	} else {
		if child := tree.Child("restSettings"); gs.RestSettings != nil && child != nil {
			result["restSettings"] = restCorsSettingsToMap(gs.RestSettings, child)
		}
		if child := tree.Child("imageTextRecognitionSettings"); gs.ImageTextRecognitionSettings != nil && child != nil {
			result["imageTextRecognitionSettings"] = imageTextRecognitionSettingsToMap(gs.ImageTextRecognitionSettings, child)
		}
		if child := tree.Child("systemSettings"); gs.SystemSettings != nil && child != nil {
			result["systemSettings"] = systemSettingsToMap(gs.SystemSettings, child)
		}
		if child := tree.Child("notificationSettings"); gs.NotificationSettings != nil && child != nil {
			result["notificationSettings"] = notificationSettingsToMap(gs.NotificationSettings, child)
		}
	}
	result["$type"] = "GlobalSettings"
	return result
}

func restCorsSettingsToMap(s *model.RestCorsSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.RestCorsSettings{}
	}
	if tree == nil || tree.IsEmpty() {
		result["allowAllOrigins"] = s.AllowAllOrigins
		result["allowedOrigins"] = s.AllowedOrigins
	} else {
		if tree.Has("allowAllOrigins") {
			result["allowAllOrigins"] = s.AllowAllOrigins
		}
		if tree.Has("allowedOrigins") {
			result["allowedOrigins"] = s.AllowedOrigins
		}
	}
	result["$type"] = "RestCorsSettings"
	return result
}

func imageTextRecognitionSettingsToMap(s *model.ImageTextRecognitionSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.ImageTextRecognitionSettings{}
	}
	if tree == nil || tree.IsEmpty() {
		result["enabled"] = s.Enabled
	} else if tree.Has("enabled") {
		result["enabled"] = s.Enabled
	}
	result["$type"] = "ImageTextRecognitionSettings"
	return result
}

func systemSettingsToMap(s *model.SystemSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.SystemSettings{}
	}
	if tree == nil || tree.IsEmpty() {
		result["ocrSupported"] = s.OcrSupported
	} else if tree.Has("ocrSupported") {
		result["ocrSupported"] = s.OcrSupported
	}
	result["$type"] = "SystemSettings"
	return result
}

func emailSettingsToMap(s *model.EmailSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.EmailSettings{}
	}
	if tree == nil || tree.IsEmpty() {
		result["isEnabled"] = s.IsEnabled
		result["isDefault"] = s.IsDefault
	} else {
		if tree.Has("isEnabled") {
			result["isEnabled"] = s.IsEnabled
		}
		if tree.Has("isDefault") {
			result["isDefault"] = s.IsDefault
		}
	}
	result["$type"] = "EmailSettings"
	return result
}

func notificationSettingsToMap(s *model.NotificationSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.NotificationSettings{}
	}
	if s.EmailSettings != nil {
		if tree == nil || tree.IsEmpty() || tree.Has("emailSettings") {
			var emailTree *fields.FieldTree
			if tree != nil {
				emailTree = tree.Child("emailSettings")
			}
			result["emailSettings"] = emailSettingsToMap(s.EmailSettings, emailTree)
		}
	}
	result["$type"] = "NotificationSettings"
	return result
}

// ProjectDashboard يعيد لوحة ودجات المشروع المحدد مع احترام معامل fields
// (مطابق لـ request13.txt و request23.txt).
func (h *AdminHandler) ProjectDashboard(w http.ResponseWriter, r *http.Request) {
	projectKey := r.PathValue("id")
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	dashboard, err := h.app.GetProjectDashboard(r.Context(), projectKey, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projectDashboardToMap(dashboard, fieldTree))
}

// projectDashboardToMap يحوّل لوحة ودجات المشروع إلى خريطة تحترم معامل fields
// وتضمّن $type على كل مستوى كما في استجابة YouTrack الأصلية.
func projectDashboardToMap(d *model.ProjectDashboard, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if d == nil {
		d = &model.ProjectDashboard{}
	}

	// عندما لا تُطلب widgets في معامل fields نعرض "$type" فقط.
	widgetsMissing := tree != nil && !tree.IsEmpty() && !tree.Has("widgets")
	if !widgetsMissing {
		var widgetTree *fields.FieldTree
		if tree != nil && !tree.IsEmpty() {
			widgetTree = tree.Child("widgets")
		}
		widgets := make([]map[string]any, 0, len(d.Widgets))
		for _, w := range d.Widgets {
			widgets = append(widgets, projectDashboardWidgetToMap(w, widgetTree))
		}
		result["widgets"] = widgets
	}
	result["$type"] = "ProjectDashboard"
	return result
}

// projectDashboardWidgetToMap يحوّل ودجت لوحة مشروع إلى خريطة تحترم شجرة
// الحقول الفرعية وتضمّن $type.
func projectDashboardWidgetToMap(w *model.ProjectDashboardWidget, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if w == nil {
		w = &model.ProjectDashboardWidget{}
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = w.ID
		result["key"] = w.Key
		result["x"] = w.X
		result["y"] = w.Y
		result["width"] = w.Width
		result["height"] = w.Height
		result["settings"] = w.Settings
		if w.Widget != nil {
			result["widget"] = projectDashboardWidgetViewToMap(w.Widget)
		}
	} else {
		if tree.Has("id") {
			result["id"] = w.ID
		}
		if tree.Has("key") {
			result["key"] = w.Key
		}
		if tree.Has("x") {
			result["x"] = w.X
		}
		if tree.Has("y") {
			result["y"] = w.Y
		}
		if tree.Has("width") {
			result["width"] = w.Width
		}
		if tree.Has("height") {
			result["height"] = w.Height
		}
		if tree.Has("settings") {
			result["settings"] = w.Settings
		}
		if tree.Has("widget") && w.Widget != nil {
			result["widget"] = projectDashboardWidgetViewToMap(w.Widget)
		}
	}
	result["$type"] = "ProjectDashboardWidget"
	return result
}

func projectDashboardWidgetViewToMap(w *model.DashboardWidget) map[string]any {
	result := make(map[string]any)
	if w == nil {
		w = &model.DashboardWidget{}
	}
	if w.ID != "" {
		result["id"] = w.ID
	}
	result["$type"] = "WidgetView"
	return result
}
