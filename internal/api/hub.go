package api

import (
	"net/http"
	"strconv"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// HubHandler يعالج طلبات Hub REST API.
type HubHandler struct {
	app *app.App
}

// NewHubHandler ينشئ معالج طلبات Hub.
func NewHubHandler(a *app.App) *HubHandler {
	return &HubHandler{app: a}
}

// GetServices يعيد قائمة خدمات Hub مع احترام معاملات fields و $top و $skip
// (مطابق لـ request24.txt و hh.json). الافتراضي لـ $top هو 100، و -1 يجلب الكل.
func (h *HubHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))

	top := 100
	if s := r.URL.Query().Get("$top"); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			top = val
		}
	}
	skip := 0
	if s := r.URL.Query().Get("$skip"); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			skip = val
		}
	}

	page, err := h.app.GetServices(r.Context(), fieldTree, top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, servicesPageToMap(page, fieldTree))
}

// servicesPageToMap يحوّل صفحة الخدمات إلى خريطة تحترم معامل fields وتضمّن
// type ("ServicesPage") كما في استجابة Hub الأصلية.
func servicesPageToMap(page *model.ServicesPage, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if page == nil {
		page = &model.ServicesPage{
			Type:     "ServicesPage",
			Services: []*model.HubService{},
		}
	}
	result["type"] = "ServicesPage"
	result["skip"] = page.Skip
	result["top"] = page.Top
	result["total"] = page.Total

	services := make([]map[string]any, 0, len(page.Services))
	for _, s := range page.Services {
		services = append(services, serviceToMap(s, tree))
	}
	result["services"] = services
	return result
}

// serviceToMap يحوّل خدمة إلى خريطة تحترم شجرة الحقول وتضمّن type ("service")
// كما في استجابة Hub الأصلية.
func serviceToMap(s *model.HubService, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.HubService{}
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = nullOr(s.ID)
		result["name"] = nullOr(s.Name)
		result["key"] = nullOr(s.Key)
		result["homeUrl"] = nullOr(s.HomeURL)
		result["applicationName"] = nullOr(s.ApplicationName)
		result["vendor"] = nullOr(s.Vendor)
		result["version"] = nullOr(s.Version)
		if s.Trusted != nil {
			result["trusted"] = *s.Trusted
		}
		result["iconUrl"] = nullOr(s.IconURL)
		result["userUriPattern"] = nullOr(s.UserUriPattern)
		result["groupUriPattern"] = nullOr(s.GroupUriPattern)
		result["audience"] = nullOr(s.Audience)
		if s.Immutable != nil {
			result["immutable"] = *s.Immutable
		}
		if s.ClientCredentialsFlowEnabled != nil {
			result["clientCredentialsFlowEnabled"] = *s.ClientCredentialsFlowEnabled
		}
		if s.AuthCodeFlowEnabled != nil {
			result["authCodeFlowEnabled"] = *s.AuthCodeFlowEnabled
		}
		if s.ImplicitFlowEnabled != nil {
			result["implicitFlowEnabled"] = *s.ImplicitFlowEnabled
		}
	} else {
		if tree.Has("id") {
			result["id"] = nullOr(s.ID)
		}
		if tree.Has("name") {
			result["name"] = nullOr(s.Name)
		}
		if tree.Has("key") {
			result["key"] = nullOr(s.Key)
		}
		if tree.Has("homeUrl") {
			result["homeUrl"] = nullOr(s.HomeURL)
		}
		if tree.Has("applicationName") {
			result["applicationName"] = nullOr(s.ApplicationName)
		}
		if tree.Has("vendor") {
			result["vendor"] = nullOr(s.Vendor)
		}
		if tree.Has("version") {
			result["version"] = nullOr(s.Version)
		}
		if tree.Has("trusted") && s.Trusted != nil {
			result["trusted"] = *s.Trusted
		}
		if tree.Has("iconUrl") {
			result["iconUrl"] = nullOr(s.IconURL)
		}
		if tree.Has("userUriPattern") {
			result["userUriPattern"] = nullOr(s.UserUriPattern)
		}
		if tree.Has("groupUriPattern") {
			result["groupUriPattern"] = nullOr(s.GroupUriPattern)
		}
		if tree.Has("audience") {
			result["audience"] = nullOr(s.Audience)
		}
		if tree.Has("immutable") && s.Immutable != nil {
			result["immutable"] = *s.Immutable
		}
		if tree.Has("clientCredentialsFlowEnabled") && s.ClientCredentialsFlowEnabled != nil {
			result["clientCredentialsFlowEnabled"] = *s.ClientCredentialsFlowEnabled
		}
		if tree.Has("authCodeFlowEnabled") && s.AuthCodeFlowEnabled != nil {
			result["authCodeFlowEnabled"] = *s.AuthCodeFlowEnabled
		}
		if tree.Has("implicitFlowEnabled") && s.ImplicitFlowEnabled != nil {
			result["implicitFlowEnabled"] = *s.ImplicitFlowEnabled
		}
	}
	result["type"] = "service"
	return result
}
