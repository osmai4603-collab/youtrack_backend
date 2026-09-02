package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
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
	app *app.YouTrackApp
}

func NewAdminHandler(a *app.YouTrackApp) *AdminHandler {
	return &AdminHandler{app: a}
}

func (api *API) InitAdmin() {
	handler := NewAdminHandler(api.newApp())

	api.BaseRoutes.APIRoot.Handle("/roles", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Roles(w, r)
	})).Methods("GET")
	api.BaseRoutes.APIRoot.Handle("/roles/{name:[A-Za-z0-9_.\\-]+}", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.RoleByName(w, r)
	})).Methods("GET")

	api.BaseRoutes.APIRoot.Handle("/permissions", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Permissions(w, r)
	})).Methods("GET")
	api.BaseRoutes.APIRoot.Handle("/permissions/{name:[A-Za-z0-9_.\\-]+}", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.PermissionByName(w, r)
	})).Methods("GET")

	api.BaseRoutes.APIRoot.Handle("/permissions/cache", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.PermissionsCache(w, r)
	})).Methods("GET")

	api.BaseRoutes.Admin.Handle("/globalSettings", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GlobalSettings(w, r)
	})).Methods("GET")

	api.BaseRoutes.Admin.Handle("/widgets/general", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Widgets(w, r)
	})).Methods("GET")

	api.BaseRoutes.Admin.Handle("/organizations", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Organizations(w, r)
	})).Methods("GET")

	api.BaseRoutes.Admin.Handle("/projects/{id:[A-Za-z0-9_\\-\\.]+}/dashboard", api.APISessionRequiredWithPermission(app.PermissionSystemAdmin, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.ProjectDashboard(w, r)
	})).Methods("GET")
}

func (h *AdminHandler) Roles(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	roles, err := h.app.GetRoles(c.AppContext)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, RoleListResponse{Roles: roles, Count: len(roles)})
}

func (h *AdminHandler) Permissions(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	perms, err := h.app.GetPermissions(c.AppContext)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, PermissionListResponse{Permissions: perms, Count: len(perms)})
}

func (h *AdminHandler) RoleByName(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	name := mux.Vars(r)["name"]
	role, err := h.app.GetRoleByName(c.AppContext, name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, role)
}

func (h *AdminHandler) PermissionByName(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	name := mux.Vars(r)["name"]
	perm, err := h.app.GetPermissionByName(c.AppContext, name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, perm)
}

// PermissionsCache يعيد الصلاحيات المخبأة للمستخدم الحالي مع احترام معامل fields.
func (h *AdminHandler) PermissionsCache(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("Admin.PermissionsCache", "user context missing"))
		return
	}
	fieldTree := c.FieldsTree
	cache, err := h.app.GetPermissionsCache(c.AppContext, c.UserID(), fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, permissionsCacheToSlice(cache, fieldTree))
}

func permissionsCacheToSlice(cache []*model.PermissionCacheEntry, tree *fields.FieldTree) []map[string]any {
	result := make([]map[string]any, 0, len(cache))
	for _, c := range cache {
		entry := make(map[string]any)
		if c == nil {
			entry["$type"] = "CachedPermission"
			result = append(result, entry)
			continue
		}
		if tree == nil || tree.IsEmpty() {
			if c.ID != "" {
				entry["id"] = c.ID
			}
			if c.Global != nil {
				entry["global"] = *c.Global
			}
			entry["projects"] = permissionCacheProjectsToSlice(c.Projects, nil)
			entry["organizations"] = permissionCacheOrgsToSlice(c.Organizations, nil)
		} else {
			if tree.Has("id") {
				entry["id"] = c.ID
			}
			if tree.Has("global") {
				entry["global"] = *c.Global
			}
			if tree.Has("projects") {
				entry["projects"] = permissionCacheProjectsToSlice(c.Projects, tree.Child("projects"))
			}
			if tree.Has("organizations") {
				entry["organizations"] = permissionCacheOrgsToSlice(c.Organizations, tree.Child("organizations"))
			}
			if tree.Has("permission") {
				entry["permission"] = permissionToMapR26(c, tree.Child("permission"))
			}
		}
		entry["$type"] = "CachedPermission"
		result = append(result, entry)
	}
	if result == nil {
		result = []map[string]any{}
	}
	return result
}

func permissionToMapR26(c *model.PermissionCacheEntry, tree *fields.FieldTree) map[string]any {
	m := make(map[string]any)
	if tree == nil || tree.IsEmpty() {
		m["id"] = c.ID
		m["key"] = c.ID
		m["name"] = c.PermissionName
	} else {
		if tree.Has("id") {
			m["id"] = c.ID
		}
		if tree.Has("key") {
			m["key"] = c.ID
		}
		if tree.Has("name") {
			m["name"] = c.PermissionName
		}
	}
	m["$type"] = "Permission"
	return m
}

func permissionCacheProjectsToSlice(projects []*model.PermissionCacheProject, tree *fields.FieldTree) any {
	if projects == nil {
		return nil
	}
	wantID := tree == nil || tree.IsEmpty() || tree.Has("id")
	wantProjectType := tree == nil || tree.IsEmpty() || tree.Has("projectType") || tree.Has("projectType.id")

	out := make([]map[string]any, 0, len(projects))
	for _, p := range projects {
		m := make(map[string]any)
		if wantID && p.ID != "" {
			m["id"] = p.ID
		}
		if wantProjectType && p.ProjectType != nil {
			pt := make(map[string]any)
			pt["id"] = p.ProjectType.ID
			pt["$type"] = "ProjectType"
			m["projectType"] = pt
		}
		m["$type"] = "Project"
		out = append(out, m)
	}
	return out
}

func permissionCacheOrgsToSlice(orgs []*model.PermissionCacheOrganization, tree *fields.FieldTree) any {
	if orgs == nil {
		return nil
	}
	wantID := tree == nil || tree.IsEmpty() || tree.Has("id")

	out := make([]map[string]any, 0, len(orgs))
	for _, o := range orgs {
		m := make(map[string]any)
		if wantID && o.ID != "" {
			m["id"] = o.ID
		}
		m["$type"] = "Organization"
		out = append(out, m)
	}
	return out
}

func (h *AdminHandler) Widgets(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	widgets, err := h.app.GetWidgets(c.AppContext)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, widgets)
}

func (h *AdminHandler) GlobalSettings(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	fieldTree := c.FieldsTree
	gs := h.app.GetAdminGlobalSettings(c.AppContext, fieldTree)
	writeJSON(w, http.StatusOK, globalSettingsToMap(gs, fieldTree))
}

func globalSettingsToMap(gs *model.AdminGlobalSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if gs == nil {
		gs = model.DefaultAdminGlobalSettings()
	}

	if tree == nil || tree.IsEmpty() {
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

// ProjectDashboard يعيد لوحة ودجات المشروع المحدد مع احترام معامل fields.
func (h *AdminHandler) ProjectDashboard(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	projectKey := mux.Vars(r)["id"]
	if projectKey == "" {
		projectKey = r.PathValue("id")
	}

	fieldTree := c.FieldsTree
	dashboard, err := h.app.GetProjectDashboard(c.AppContext, projectKey, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projectDashboardToMap(dashboard, fieldTree))
}

func projectDashboardToMap(d *model.ProjectDashboard, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if d == nil {
		d = &model.ProjectDashboard{}
	}

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

// Organizations يعيد قائمة المنظمات.
func (h *AdminHandler) Organizations(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	fieldTree := c.FieldsTree
	top, skip := parsePagination(r)
	if s := r.URL.Query().Get("$top"); s == "-1" {
		top = -1
	} else if s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			top = val
		}
	}
	sorting := r.URL.Query().Get("sorting")
	if sorting == "" {
		sorting = r.URL.Query().Get("$sorting")
	}

	orgs, err := h.app.GetOrganizations(c.AppContext, fieldTree, top, skip, sorting)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, organizationsToSlice(orgs, fieldTree))
}

func organizationsToSlice(orgs []*model.Organization, tree *fields.FieldTree) []map[string]any {
	result := make([]map[string]any, 0, len(orgs))
	for _, o := range orgs {
		result = append(result, organizationToMap(o, tree))
	}
	if result == nil {
		result = []map[string]any{}
	}
	return result
}

func organizationToMap(o *model.Organization, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if o == nil {
		o = &model.Organization{}
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = o.ID
	} else {
		if tree.Has("id") {
			result["id"] = o.ID
		}
		if tree.Has("key") {
			result["key"] = o.Key
		}
		if tree.Has("name") {
			result["name"] = o.Name
		}
		if tree.Has("iconUrl") {
			if o.IconURL != nil {
				result["iconUrl"] = *o.IconURL
			} else {
				result["iconUrl"] = nil
			}
		}
		if tree.Has("projectsCount") {
			result["projectsCount"] = o.ProjectsCount
		}
		if tree.Has("auditTargetId") {
			result["auditTargetId"] = o.AuditTargetID
		}
		if tree.Has("description") {
			result["description"] = o.Description
		}
		if tree.Has("projects") {
			result["projects"] = organizationProjectsToSlice(o.Projects, tree.Child("projects"))
		}
	}
	result["$type"] = "Organization"
	return result
}

func organizationProjectsToSlice(projects []*model.Project, tree *fields.FieldTree) any {
	if projects == nil {
		return nil
	}
	out := make([]map[string]any, 0, len(projects))
	for _, p := range projects {
		if p == nil {
			continue
		}
		m := make(map[string]any)
		want := func(name string) bool { return tree == nil || tree.IsEmpty() || tree.Has(name) }

		if want("id") {
			m["id"] = p.ID
		}
		if want("name") {
			m["name"] = p.Name
		}
		if want("shortName") {
			m["shortName"] = p.ShortName
		}
		if want("pinned") {
			m["pinned"] = p.Pinned
		}
		if want("iconUrl") {
			if p.IconURL != "" {
				m["iconUrl"] = p.IconURL
			} else {
				m["iconUrl"] = nil
			}
		}
		if want("template") {
			m["template"] = p.Template
		}
		if want("archived") {
			m["archived"] = p.Archived
		}
		if want("restricted") {
			m["restricted"] = p.Restricted
		}
		if want("hasArticles") {
			m["hasArticles"] = p.HasArticles
		}
		if want("projectType") && p.ProjectType != nil {
			pt := make(map[string]any)
			pt["id"] = p.ProjectType.ID
			pt["$type"] = "ProjectType"
			m["projectType"] = pt
		}
		if want("team") && p.Team != nil {
			team := make(map[string]any)
			team["id"] = p.Team.ID
			team["$type"] = "ProjectTeam"
			m["team"] = team
		}
		m["$type"] = "Project"
		out = append(out, m)
	}
	return out
}
