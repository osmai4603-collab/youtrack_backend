package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
	"youtrack_backend/internal/store"
)

type mockAdminStore struct {
	cache        []*model.CachedPermission
	cacheEntries []*model.PermissionCacheEntry
	orgs         []*model.Organization
	services     *model.ServicesPage
}

func (m *mockAdminStore) Roles(ctx context.Context) ([]*model.Role, error) {
	return []*model.Role{}, nil
}
func (m *mockAdminStore) Permissions(ctx context.Context) ([]*model.Permission, error) {
	return []*model.Permission{}, nil
}
func (m *mockAdminStore) GlobalSettings(ctx context.Context) (*model.GlobalSettings, error) {
	return &model.GlobalSettings{}, nil
}
func (m *mockAdminStore) AdminGlobalSettings(ctx context.Context, tree *fields.FieldTree) (*model.AdminGlobalSettings, error) {
	g := model.DefaultAdminGlobalSettings()
	if tree != nil {
		if !tree.Has("restSettings") {
			g.RestSettings = nil
		}
		if !tree.Has("imageTextRecognitionSettings") {
			g.ImageTextRecognitionSettings = nil
		}
		if !tree.Has("systemSettings") {
			g.SystemSettings = nil
		}
		if !tree.Has("notificationSettings") {
			g.NotificationSettings = nil
		}
	}
	return g, nil
}
func (m *mockAdminStore) BannersConfig(ctx context.Context, tree *fields.FieldTree) (*model.BannersConfig, error) {
	b := model.DefaultBannersConfig()
	if tree != nil && tree.Child("banners") != nil {
		bt := tree.Child("banners")
		if !bt.Has("globalBanner") {
			b.GlobalBanner = ""
		}
		if !bt.Has("globalBannerEnabled") {
			b.GlobalBannerEnabled = false
		}
		if !bt.Has("systemEventsBanners") {
			b.SystemEventsBanners = nil
		}
	}
	return b, nil
}
func (m *mockAdminStore) Widgets(ctx context.Context) ([]*model.DashboardWidget, error) {
	return []*model.DashboardWidget{}, nil
}
func (m *mockAdminStore) CachedPermissions(ctx context.Context, userID string) ([]*model.CachedPermission, error) {
	if userID != "2-1" {
		return []*model.CachedPermission{}, nil
	}
	return m.cache, nil
}

// PermissionsCache يحاكي جلب الصلاحيات المخبأة مع احترام شجرة الحقول
// من خلال تصفية مصفوفة cacheEntries حسب الحقول المطلوبة.
func (m *mockAdminStore) PermissionsCache(ctx context.Context, userID string, tree *fields.FieldTree) ([]*model.PermissionCacheEntry, error) {
	if userID != "2-1" {
		return []*model.PermissionCacheEntry{}, nil
	}
	if m.cacheEntries == nil {
		return []*model.PermissionCacheEntry{}, nil
	}
	return m.cacheEntries, nil
}

// ProjectDashboard يحاكي جلب لوحة ودجات المشروع مع احترام شجرة الحقول،
// ويعيد pgx.ErrNoRows للمشروع "missing" لمحاكاة حالة 404.
func (m *mockAdminStore) ProjectDashboard(ctx context.Context, projectKey string, tree *fields.FieldTree) (*model.ProjectDashboard, error) {
	if projectKey == "missing" {
		return nil, pgx.ErrNoRows
	}

	dashboard := &model.ProjectDashboard{
		Widgets: []*model.ProjectDashboardWidget{
			{
				ID:       "w-1",
				Key:      "issues",
				X:        0,
				Y:        0,
				Width:    4,
				Height:   3,
				Widget:   &model.DashboardWidget{ID: "dash-1", Type: "WidgetView"},
				Settings: "{\"query\":\"#Unresolved\"}",
				Type:     "ProjectDashboardWidget",
			},
		},
		Type: "ProjectDashboard",
	}

	var widgetTree *fields.FieldTree
	if tree != nil && !tree.IsEmpty() {
		widgetTree = tree.Child("widgets")
		if widgetTree == nil {
			dashboard.Widgets = nil
			return dashboard, nil
		}
	}

	if widgetTree != nil && !widgetTree.IsEmpty() {
		for _, w := range dashboard.Widgets {
			if !widgetTree.Has("id") {
				w.ID = ""
			}
			if !widgetTree.Has("key") {
				w.Key = ""
			}
			if !widgetTree.Has("x") {
				w.X = 0
			}
			if !widgetTree.Has("y") {
				w.Y = 0
			}
			if !widgetTree.Has("width") {
				w.Width = 0
			}
			if !widgetTree.Has("height") {
				w.Height = 0
			}
			if !widgetTree.Has("settings") {
				w.Settings = ""
			}
			if !widgetTree.Has("widget") {
				w.Widget = nil
			}
		}
	}
	return dashboard, nil
}

// Organizations يحاكي جلب المنظمات مع احترام شجرة الحقول والحدود.
func (m *mockAdminStore) Organizations(ctx context.Context, tree *fields.FieldTree, top int, skip int, sorting string) ([]*model.Organization, error) {
	orgs := m.orgs
	if orgs == nil {
		orgs = []*model.Organization{}
	}
	// تطبيق الشجرة على المشاريع الفرعية (Mock: إعادة فارغة إلا عند طلب projects)
	for _, o := range orgs {
		if tree == nil || tree.IsEmpty() || tree.Has("projects") {
			o.Projects = []*model.Project{
				{
					ID:          "22-59",
					Name:        "DEMO",
					ShortName:   "DEMO",
					Pinned:      true,
					IconURL:     "",
					Template:    false,
					Archived:    false,
					Restricted:  false,
					HasArticles: true,
					ProjectType: &model.ProjectType{ID: "DEFAULT", Type: "ProjectType"},
					Team:        &model.ProjectTeamDetailed{ID: "team-1", Type: "ProjectTeam"},
					Type:        "Project",
				},
			}
		} else {
			o.Projects = nil
		}
		o.Normalize()
	}
	// تطبيق $skip ثم $top
	if skip > 0 && skip < len(orgs) {
		orgs = orgs[skip:]
	} else if skip > 0 {
		orgs = nil
	}
	if top > 0 && top < len(orgs) {
		orgs = orgs[:top]
	}
	return orgs, nil
}

// Services يحاكي جلب صفحة خدمات Hub مع احترام شجرة الحقول والحدود.
func (m *mockAdminStore) Services(ctx context.Context, tree *fields.FieldTree, top int, skip int) (*model.ServicesPage, error) {
	if m.services == nil {
		return &model.ServicesPage{Type: "ServicesPage", Services: []*model.HubService{}}, nil
	}
	page := &model.ServicesPage{
		Type:     "ServicesPage",
		Skip:     skip,
		Top:      top,
		Total:    len(m.services.Services),
		Services: m.services.Services,
	}
	// تطبيق $skip ثم $top (مثل المنظمات)
	items := page.Services
	if skip > 0 && skip < len(items) {
		items = items[skip:]
	} else if skip > 0 {
		items = nil
	}
	if top > 0 && top < len(items) {
		items = items[:top]
	}
	page.Services = items
	return page, nil
}

type mockAdminStoreHolder struct {
	admin *mockAdminStore
}

func (m *mockAdminStoreHolder) Users() store.UserStore                 { return nil }
func (m *mockAdminStoreHolder) Projects() store.ProjectStore           { return nil }
func (m *mockAdminStoreHolder) Issues() store.IssueStore               { return nil }
func (m *mockAdminStoreHolder) Admin() store.AdminStore                { return m.admin }
func (m *mockAdminStoreHolder) Inbox() store.InboxStore                { return nil }
func (m *mockAdminStoreHolder) SavedQueries() store.SavedQueryStore    { return nil }
func (m *mockAdminStoreHolder) Search() store.SearchStore              { return nil }
func (m *mockAdminStoreHolder) Subscriptions() store.SubscriptionStore { return nil }

func TestPermissionsCacheEndpoint(t *testing.T) {
	global := false
	globalTrue := true
	cache := []*model.PermissionCacheEntry{
		{
			ID:     "jetbrains.jetpass.project-read-basic",
			Global: &global,
			Projects: []*model.PermissionCacheProject{
				{
					ID:          "22-59",
					ProjectType: &model.ProjectType{ID: "DEFAULT", Type: "ProjectType"},
					Type:        "Project",
				},
			},
			Organizations: []*model.PermissionCacheOrganization{},
			Type:          "CachedPermission",
		},
		{
			ID:            "jetbrains.jetpass.profile-updateSelf",
			Global:        &globalTrue,
			Projects:      nil,
			Organizations: nil,
			Type:          "CachedPermission",
		},
	}
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{cacheEntries: cache}}
	a := app.New(holder)
	h := NewAdminHandler(a)

	req := httptest.NewRequest("GET", "/api/permissions/cache?fields=id,global,projects(id,projectType(id)),organizations(id)", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.PermissionsCache(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 cached permissions, got %d", len(res))
	}

	p0 := res[0]
	if p0["id"] != "jetbrains.jetpass.project-read-basic" {
		t.Errorf("unexpected id %v", p0["id"])
	}
	if p0["global"] != false {
		t.Errorf("expected global false, got %v", p0["global"])
	}
	if p0["$type"] != "CachedPermission" {
		t.Errorf("expected $type CachedPermission, got %v", p0["$type"])
	}
	projects, _ := p0["projects"].([]any)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	pr, _ := projects[0].(map[string]any)
	if pr["id"] != "22-59" {
		t.Errorf("expected project id 22-59, got %v", pr["id"])
	}
	if pr["$type"] != "Project" {
		t.Errorf("expected project $type Project, got %v", pr["$type"])
	}
	if pr["projectType"].(map[string]any)["id"] != "DEFAULT" {
		t.Errorf("expected projectType DEFAULT")
	}
	if _, ok := p0["organizations"].([]any); !ok {
		t.Errorf("expected organizations array for non-global permission")
	}

	p1 := res[1]
	if p1["id"] != "jetbrains.jetpass.profile-updateSelf" {
		t.Errorf("unexpected id %v", p1["id"])
	}
	if p1["global"] != true {
		t.Errorf("expected global true, got %v", p1["global"])
	}
	if p1["projects"] != nil {
		t.Errorf("expected projects null for global permission")
	}
	if p1["organizations"] != nil {
		t.Errorf("expected organizations null for global permission")
	}
}

// TestPermissionsCacheFieldFiltering يتحقق من المطابقة الدقيقة لحقول معامل fields
// (مطابق لـ request20.txt): id, global, projects(id,projectType(id)), organizations(id).
func TestPermissionsCacheFieldFiltering(t *testing.T) {
	global := false
	cache := []*model.PermissionCacheEntry{
		{
			ID:     "jetbrains.jetpass.project-read-basic",
			Global: &global,
			Projects: []*model.PermissionCacheProject{
				{
					ID:          "22-59",
					ProjectType: &model.ProjectType{ID: "DEFAULT", Type: "ProjectType"},
					Type:        "Project",
				},
			},
			Organizations: []*model.PermissionCacheOrganization{},
			Type:          "CachedPermission",
		},
	}
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{cacheEntries: cache}}
	a := app.New(holder)
	h := NewAdminHandler(a)

	req := httptest.NewRequest("GET", "/api/permissions/cache?fields=id,global,projects(id,projectType(id)),organizations(id)", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.PermissionsCache(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 cached permission, got %d", len(res))
	}

	p := res[0]
	for _, key := range []string{"id", "global", "$type"} {
		if _, present := p[key]; !present {
			t.Errorf("expected %s present (requested field), got %v", key, p)
		}
	}

	projects, ok := p["projects"].([]any)
	if !ok || len(projects) != 1 {
		t.Fatalf("expected 1 project, got %v", p["projects"])
	}
	pr := projects[0].(map[string]any)
	if pr["id"] != "22-59" {
		t.Errorf("expected project id 22-59, got %v", pr["id"])
	}
	if pr["$type"] != "Project" {
		t.Errorf("expected project $type Project, got %v", pr["$type"])
	}
	pt, ok := pr["projectType"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested projectType object, got %v", pr["projectType"])
	}
	if pt["id"] != "DEFAULT" {
		t.Errorf("expected projectType id DEFAULT, got %v", pt["id"])
	}
	if pt["$type"] != "ProjectType" {
		t.Errorf("expected projectType $type ProjectType, got %v", pt["$type"])
	}

	if orgs, ok := p["organizations"].([]any); !ok || len(orgs) != 0 {
		t.Errorf("expected empty organizations array, got %v", p["organizations"])
	}
}

// TestPermissionsCachePartialFields يتحقق أن طلب حقول فرعية فقط لا يُرجع حقولاً
// غير مطلوبة في المشاريع (مثل projectType عند عدم طلبه).
func TestPermissionsCachePartialFields(t *testing.T) {
	global := false
	cache := []*model.PermissionCacheEntry{
		{
			ID:     "jetbrains.jetpass.project-read-basic",
			Global: &global,
			Projects: []*model.PermissionCacheProject{
				{ID: "22-59", ProjectType: &model.ProjectType{ID: "DEFAULT", Type: "ProjectType"}, Type: "Project"},
			},
			Organizations: []*model.PermissionCacheOrganization{},
			Type:          "CachedPermission",
		},
	}
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{cacheEntries: cache}}
	h := NewAdminHandler(app.New(holder))

	req := httptest.NewRequest("GET", "/api/permissions/cache?fields=id,global,projects(id)", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.PermissionsCache(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 cached permission, got %d", len(res))
	}

	p := res[0]
	if _, present := p["projects"]; !present {
		t.Fatalf("expected projects field, got %v", p)
	}
	projects, ok := p["projects"].([]any)
	if !ok || len(projects) != 1 {
		t.Fatalf("expected 1 project, got %v", p["projects"])
	}
	pr := projects[0].(map[string]any)
	if pr["id"] != "22-59" {
		t.Errorf("expected project id, got %v", pr["id"])
	}
	if _, present := pr["projectType"]; present {
		t.Errorf("expected projectType absent (not requested), got %v", pr)
	}
	if pr["$type"] != "Project" {
		t.Errorf("expected project $type Project, got %v", pr["$type"])
	}
}

// TestPermissionsCacheRequest26 يتحقق من المطابقة الدقيقة للطلب رقم 26:
// fields=global,permission(id,key,name),projects(id)
func TestPermissionsCacheRequest26(t *testing.T) {
	global := false
	cache := []*model.PermissionCacheEntry{
		{
			ID:             "jetbrains.jetpass.project-read",
			Global:         &global,
			PermissionName: "Read Project",
			Projects: []*model.PermissionCacheProject{
				{ID: "0-0", Type: "Project"},
			},
			Type: "CachedPermission",
		},
	}
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{cacheEntries: cache}}
	h := NewAdminHandler(app.New(holder))

	req := httptest.NewRequest("GET", "/api/permissions/cache?fields=global,permission(id,key,name),projects(id)", nil)
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	h.PermissionsCache(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}

	p := res[0]
	if p["global"] != false {
		t.Errorf("expected global false")
	}
	if p["$type"] != "CachedPermission" {
		t.Errorf("expected $type CachedPermission")
	}

	perm, ok := p["permission"].(map[string]any)
	if !ok {
		t.Fatalf("expected permission object, got %v", p["permission"])
	}
	if perm["id"] != "jetbrains.jetpass.project-read" {
		t.Errorf("expected permission id, got %v", perm["id"])
	}
	if perm["key"] != "jetbrains.jetpass.project-read" {
		t.Errorf("expected permission key, got %v", perm["key"])
	}
	if perm["name"] != "Read Project" {
		t.Errorf("expected permission name, got %v", perm["name"])
	}
	if perm["$type"] != "Permission" {
		t.Errorf("expected permission $type Permission")
	}

	projects, ok := p["projects"].([]any)
	if !ok || len(projects) != 1 {
		t.Fatalf("expected 1 project, got %v", p["projects"])
	}
	pr := projects[0].(map[string]any)
	if pr["id"] != "0-0" {
		t.Errorf("expected project id 0-0")
	}
	if _, present := pr["projectType"]; present {
		t.Errorf("projectType should be absent")
	}
}

func TestPermissionsCacheUnauthorized(t *testing.T) {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{}}
	h := NewAdminHandler(app.New(holder))

	req := httptest.NewRequest("GET", "/api/permissions/cache", nil)
	rec := httptest.NewRecorder()

	h.PermissionsCache(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

func newGlobalSettingsHandler() *AdminHandler {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{}}
	return NewAdminHandler(app.New(holder))
}

func TestGlobalSettingsRequest5(t *testing.T) {
	h := newGlobalSettingsHandler()
	req := httptest.NewRequest("GET", "/api/admin/globalSettings?fields=restSettings(allowAllOrigins,allowedOrigins),imageTextRecognitionSettings(enabled),systemSettings(ocrSupported),notificationSettings(emailSettings(isEnabled))", nil)
	rec := httptest.NewRecorder()

	h.GlobalSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res["$type"] != "GlobalSettings" {
		t.Errorf("expected $type GlobalSettings, got %v", res["$type"])
	}

	rest, ok := res["restSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected restSettings object, got %v", res["restSettings"])
	}
	if rest["$type"] != "RestCorsSettings" {
		t.Errorf("expected restSettings $type RestCorsSettings, got %v", rest["$type"])
	}
	if rest["allowAllOrigins"] != false {
		t.Errorf("expected allowAllOrigins false, got %v", rest["allowAllOrigins"])
	}
	if origins, ok := rest["allowedOrigins"].([]any); !ok || len(origins) != 0 {
		t.Errorf("expected allowedOrigins [], got %v", rest["allowedOrigins"])
	}

	itr, ok := res["imageTextRecognitionSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected imageTextRecognitionSettings object, got %v", res["imageTextRecognitionSettings"])
	}
	if itr["$type"] != "ImageTextRecognitionSettings" {
		t.Errorf("expected imageTextRecognitionSettings $type, got %v", itr["$type"])
	}
	if itr["enabled"] != true {
		t.Errorf("expected enabled true, got %v", itr["enabled"])
	}

	sys, ok := res["systemSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected systemSettings object, got %v", res["systemSettings"])
	}
	if sys["$type"] != "SystemSettings" {
		t.Errorf("expected systemSettings $type, got %v", sys["$type"])
	}
	if sys["ocrSupported"] != "true" {
		t.Errorf("expected ocrSupported \"true\", got %v", sys["ocrSupported"])
	}

	notif, ok := res["notificationSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected notificationSettings object, got %v", res["notificationSettings"])
	}
	if notif["$type"] != "NotificationSettings" {
		t.Errorf("expected notificationSettings $type, got %v", notif["$type"])
	}
	email, ok := notif["emailSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected emailSettings object, got %v", notif["emailSettings"])
	}
	if email["$type"] != "EmailSettings" {
		t.Errorf("expected emailSettings $type, got %v", email["$type"])
	}
	if email["isEnabled"] != true {
		t.Errorf("expected isEnabled true, got %v", email["isEnabled"])
	}
	if _, present := email["isDefault"]; present {
		t.Errorf("expected isDefault to be absent for request5 fields, got %v", email["isDefault"])
	}
}

// TestGlobalSettingsRequest19 يتحقق من شكل استجابة request19.txt:
// GET /api/admin/globalSettings?fields=restSettings(allowAllOrigins,allowedOrigins),imageTextRecognitionSettings(enabled),systemSettings(ocrSupported),notificationSettings(emailSettings(isEnabled))
// وهو مطابق لشكل request5.txt.
func TestGlobalSettingsRequest19(t *testing.T) {
	h := newGlobalSettingsHandler()
	req := httptest.NewRequest("GET", "/api/admin/globalSettings?fields=restSettings(allowAllOrigins,allowedOrigins),imageTextRecognitionSettings(enabled),systemSettings(ocrSupported),notificationSettings(emailSettings(isEnabled))", nil)
	rec := httptest.NewRecorder()

	h.GlobalSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res["$type"] != "GlobalSettings" {
		t.Errorf("expected $type GlobalSettings, got %v", res["$type"])
	}

	rest, ok := res["restSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected restSettings object, got %v", res["restSettings"])
	}
	if rest["$type"] != "RestCorsSettings" {
		t.Errorf("expected restSettings $type RestCorsSettings, got %v", rest["$type"])
	}
	if rest["allowAllOrigins"] != false {
		t.Errorf("expected allowAllOrigins false, got %v", rest["allowAllOrigins"])
	}
	if origins, ok := rest["allowedOrigins"].([]any); !ok || len(origins) != 0 {
		t.Errorf("expected allowedOrigins [], got %v", rest["allowedOrigins"])
	}

	itr, ok := res["imageTextRecognitionSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected imageTextRecognitionSettings object, got %v", res["imageTextRecognitionSettings"])
	}
	if itr["$type"] != "ImageTextRecognitionSettings" {
		t.Errorf("expected imageTextRecognitionSettings $type, got %v", itr["$type"])
	}
	if itr["enabled"] != true {
		t.Errorf("expected enabled true, got %v", itr["enabled"])
	}

	sys, ok := res["systemSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected systemSettings object, got %v", res["systemSettings"])
	}
	if sys["$type"] != "SystemSettings" {
		t.Errorf("expected systemSettings $type, got %v", sys["$type"])
	}
	if sys["ocrSupported"] != "true" {
		t.Errorf("expected ocrSupported \"true\", got %v", sys["ocrSupported"])
	}

	notif, ok := res["notificationSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected notificationSettings object, got %v", res["notificationSettings"])
	}
	if notif["$type"] != "NotificationSettings" {
		t.Errorf("expected notificationSettings $type, got %v", notif["$type"])
	}
	email, ok := notif["emailSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected emailSettings object, got %v", notif["emailSettings"])
	}
	if email["$type"] != "EmailSettings" {
		t.Errorf("expected emailSettings $type, got %v", email["$type"])
	}
	if email["isEnabled"] != true {
		t.Errorf("expected isEnabled true, got %v", email["isEnabled"])
	}
	if _, present := email["isDefault"]; present {
		t.Errorf("expected isDefault to be absent for request19 fields, got %v", email["isDefault"])
	}
}

func TestGlobalSettingsRequest46(t *testing.T) {
	h := newGlobalSettingsHandler()
	req := httptest.NewRequest("GET", "/api/admin/globalSettings?fields=notificationSettings(emailSettings(isDefault))", nil)
	rec := httptest.NewRecorder()

	h.GlobalSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res["$type"] != "GlobalSettings" {
		t.Errorf("expected $type GlobalSettings, got %v", res["$type"])
	}
	if _, present := res["restSettings"]; present {
		t.Errorf("expected restSettings to be absent, got %v", res["restSettings"])
	}
	if _, present := res["systemSettings"]; present {
		t.Errorf("expected systemSettings to be absent, got %v", res["systemSettings"])
	}
	if _, present := res["imageTextRecognitionSettings"]; present {
		t.Errorf("expected imageTextRecognitionSettings to be absent, got %v", res["imageTextRecognitionSettings"])
	}

	notif, ok := res["notificationSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected notificationSettings object, got %v", res["notificationSettings"])
	}
	if notif["$type"] != "NotificationSettings" {
		t.Errorf("expected notificationSettings $type, got %v", notif["$type"])
	}
	email, ok := notif["emailSettings"].(map[string]any)
	if !ok {
		t.Fatalf("expected emailSettings object, got %v", notif["emailSettings"])
	}
	if email["$type"] != "EmailSettings" {
		t.Errorf("expected emailSettings $type, got %v", email["$type"])
	}
	if email["isDefault"] != true {
		t.Errorf("expected isDefault true, got %v", email["isDefault"])
	}
	if _, present := email["isEnabled"]; present {
		t.Errorf("expected isEnabled to be absent for request46 fields, got %v", email["isEnabled"])
	}
}

func TestGlobalSettingsFullSchema(t *testing.T) {
	h := newGlobalSettingsHandler()
	req := httptest.NewRequest("GET", "/api/admin/globalSettings", nil)
	rec := httptest.NewRecorder()

	h.GlobalSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res["$type"] != "GlobalSettings" {
		t.Errorf("expected $type GlobalSettings, got %v", res["$type"])
	}
	for _, key := range []string{"restSettings", "imageTextRecognitionSettings", "systemSettings", "notificationSettings"} {
		if _, ok := res[key].(map[string]any); !ok {
			t.Errorf("expected full schema to include %s, got %v", key, res[key])
		}
	}
}

func TestGlobalSettingsUnauthorized(t *testing.T) {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{}}
	router := NewRouter(app.New(holder), "test-secret")

	req := httptest.NewRequest("GET", "/api/admin/globalSettings", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

func newConfigHandler() *ConfigHandler {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{}}
	return NewConfigHandler(app.New(holder))
}

// TestBannersConfigRequest11 يتحقق من شكل استجابة request11.txt:
// GET /api/config?$top=-1&fields=banners(globalBanner,globalBannerEnabled,systemEventsBanners)
func TestBannersConfigRequest11(t *testing.T) {
	h := newConfigHandler()
	req := httptest.NewRequest("GET", "/api/config?$top=-1&fields=banners(globalBanner,globalBannerEnabled,systemEventsBanners)", nil)
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res["$type"] != "FrontendConfig" {
		t.Errorf("expected $type FrontendConfig, got %v", res["$type"])
	}

	banners, ok := res["banners"].(map[string]any)
	if !ok {
		t.Fatalf("expected banners object, got %v", res["banners"])
	}
	if banners["$type"] != "BannersConfig" {
		t.Errorf("expected banners $type BannersConfig, got %v", banners["$type"])
	}

	// الحقول المطلوبة حصراً من معامل fields
	for _, key := range []string{"globalBanner", "globalBannerEnabled", "systemEventsBanners"} {
		if _, present := banners[key]; !present {
			t.Errorf("expected banners to include %s (requested field), got %v", key, banners)
		}
	}

	if banners["globalBannerEnabled"] != false {
		t.Errorf("expected globalBannerEnabled false, got %v", banners["globalBannerEnabled"])
	}
	if text, _ := banners["globalBanner"].(string); text == "" {
		t.Errorf("expected non-empty globalBanner, got %v", banners["globalBanner"])
	}
	if events, ok := banners["systemEventsBanners"].([]any); !ok || len(events) != 0 {
		t.Errorf("expected systemEventsBanners [], got %v", banners["systemEventsBanners"])
	}

	// يجب أن يكون banner وحده في الاستجابة (بلا حقول config أخرى عند تخصيص fields)
	for _, key := range []string{"version", "build", "releaseDate", "helpdeskEnabled"} {
		if _, present := res[key]; present {
			t.Errorf("expected %s to be absent for request11 fields, got %v", key, res[key])
		}
	}
}

// TestBannersConfigFieldFiltering يتحقق من المطابقة الدقيقة بين حقول معامل fields
// وبين الحقول الظاهرة في الاستجابة (طلب حقل واحد فقط داخل banners).
func TestBannersConfigFieldFiltering(t *testing.T) {
	h := newConfigHandler()
	req := httptest.NewRequest("GET", "/api/config?fields=banners(globalBannerEnabled)", nil)
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	banners, ok := res["banners"].(map[string]any)
	if !ok {
		t.Fatalf("expected banners object, got %v", res["banners"])
	}
	if _, present := banners["globalBannerEnabled"]; !present {
		t.Errorf("expected globalBannerEnabled to be present, got %v", banners)
	}
	for _, key := range []string{"globalBanner", "systemEventsBanners"} {
		if _, present := banners[key]; present {
			t.Errorf("expected %s to be absent (not requested), got %v", key, banners)
		}
	}
	if banners["$type"] != "BannersConfig" {
		t.Errorf("expected banners $type BannersConfig, got %v", banners["$type"])
	}
}

// TestConfigWithoutBannersField يتحقق أن طلب fields بلا banners لا يُرجع كائن banners
// (سلامة request3.txt و request60.txt).
func TestConfigWithoutBannersField(t *testing.T) {
	h := newConfigHandler()
	req := httptest.NewRequest("GET", "/api/config?fields=helpdeskEnabled", nil)
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if _, present := res["banners"]; present {
		t.Errorf("expected banners to be absent for fields=helpdeskEnabled, got %v", res["banners"])
	}
	if _, present := res["helpdeskEnabled"]; !present {
		t.Errorf("expected helpdeskEnabled present, got %v", res["helpdeskEnabled"])
	}
	if res["$type"] != "FrontendConfig" {
		t.Errorf("expected $type FrontendConfig, got %v", res["$type"])
	}
}

func newProjectDashboardHandler() *AdminHandler {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{}}
	return NewAdminHandler(app.New(holder))
}

// TestProjectDashboardRequest13 يتحقق من شكل استجابة request13.txt:
// GET /api/admin/projects/22-59/dashboard?fields=widgets(id,key,x,y,width,height,widget(id),settings)
func TestProjectDashboardRequest13(t *testing.T) {
	h := newProjectDashboardHandler()
	req := httptest.NewRequest("GET", "/api/admin/projects/22-59/dashboard?fields=widgets(id,key,x,y,width,height,widget(id),settings)", nil)
	req.SetPathValue("id", "22-59")
	rec := httptest.NewRecorder()

	h.ProjectDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if res["$type"] != "ProjectDashboard" {
		t.Errorf("expected $type ProjectDashboard, got %v", res["$type"])
	}

	widgets, ok := res["widgets"].([]any)
	if !ok || len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %v", res["widgets"])
	}
	w := widgets[0].(map[string]any)
	for _, key := range []string{"id", "key", "x", "y", "width", "height", "settings"} {
		if _, present := w[key]; !present {
			t.Errorf("expected widget to include %s (requested field), got %v", key, w)
		}
	}
	if w["$type"] != "ProjectDashboardWidget" {
		t.Errorf("expected widget $type ProjectDashboardWidget, got %v", w["$type"])
	}
	if w["id"] != "w-1" || w["key"] != "issues" {
		t.Errorf("unexpected widget values: %v", w)
	}

	inner, ok := w["widget"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested widget object, got %v", w["widget"])
	}
	if inner["id"] != "dash-1" {
		t.Errorf("expected nested widget id dash-1, got %v", inner["id"])
	}
	if inner["$type"] != "WidgetView" {
		t.Errorf("expected nested widget $type WidgetView, got %v", inner["$type"])
	}
}

// TestProjectDashboardFieldFiltering يتحقق من المطابقة الدقيقة لحقول معامل fields.
func TestProjectDashboardFieldFiltering(t *testing.T) {
	h := newProjectDashboardHandler()
	req := httptest.NewRequest("GET", "/api/admin/projects/0-0/dashboard?fields=widgets(id)", nil)
	req.SetPathValue("id", "0-0")
	rec := httptest.NewRecorder()

	h.ProjectDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if res["$type"] != "ProjectDashboard" {
		t.Errorf("expected $type ProjectDashboard, got %v", res["$type"])
	}
	widgets, ok := res["widgets"].([]any)
	if !ok || len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %v", res["widgets"])
	}
	w := widgets[0].(map[string]any)
	if _, present := w["id"]; !present {
		t.Errorf("expected widget id present, got %v", w)
	}
	for _, key := range []string{"key", "x", "y", "width", "height", "settings", "widget"} {
		if _, present := w[key]; present {
			t.Errorf("expected %s to be absent (not requested), got %v", key, w)
		}
	}
	if w["$type"] != "ProjectDashboardWidget" {
		t.Errorf("expected widget $type ProjectDashboardWidget, got %v", w["$type"])
	}
}

// TestProjectDashboardWithoutWidgetsField يتحقق أن طلب fields بلا widgets لا يُرجع widgets.
func TestProjectDashboardWithoutWidgetsField(t *testing.T) {
	h := newProjectDashboardHandler()
	req := httptest.NewRequest("GET", "/api/admin/projects/0-0/dashboard?fields=foo", nil)
	req.SetPathValue("id", "0-0")
	rec := httptest.NewRecorder()

	h.ProjectDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if _, present := res["widgets"]; present {
		t.Errorf("expected widgets absent for fields=foo, got %v", res["widgets"])
	}
	if res["$type"] != "ProjectDashboard" {
		t.Errorf("expected $type ProjectDashboard, got %v", res["$type"])
	}
}

// TestProjectDashboardFullSchema يتحقق من الاستجابة الكاملة بدون معامل fields.
func TestProjectDashboardFullSchema(t *testing.T) {
	h := newProjectDashboardHandler()
	req := httptest.NewRequest("GET", "/api/admin/projects/0-0/dashboard", nil)
	req.SetPathValue("id", "0-0")
	rec := httptest.NewRecorder()

	h.ProjectDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if res["$type"] != "ProjectDashboard" {
		t.Errorf("expected $type ProjectDashboard, got %v", res["$type"])
	}
	widgets, ok := res["widgets"].([]any)
	if !ok || len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %v", res["widgets"])
	}
	w := widgets[0].(map[string]any)
	if w["$type"] != "ProjectDashboardWidget" {
		t.Errorf("expected widget $type ProjectDashboardWidget, got %v", w["$type"])
	}
	if inner, ok := w["widget"].(map[string]any); ok {
		if inner["$type"] != "WidgetView" {
			t.Errorf("expected nested widget $type WidgetView, got %v", inner["$type"])
		}
	}
}

// TestProjectDashboardNotFound يتحقق من 404 لمشروع غير موجود.
func TestProjectDashboardNotFound(t *testing.T) {
	h := newProjectDashboardHandler()
	req := httptest.NewRequest("GET", "/api/admin/projects/missing/dashboard", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	h.ProjectDashboard(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestProjectDashboardUnauthorized يتحقق من 401 عند الطلب عبر الراوتر بدون JWT.
func TestProjectDashboardUnauthorized(t *testing.T) {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{}}
	router := NewRouter(app.New(holder), "test-secret")

	req := httptest.NewRequest("GET", "/api/admin/projects/0-0/dashboard", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

func newOrganizationsHandler() *AdminHandler {
	icon := "https://osm.youtrack.cloud/icon.png"
	orgs := []*model.Organization{
		{
			ID:            "1-0",
			Key:           "CFSksvCi5N6T06bFUtA8w",
			Name:          "me",
			IconURL:       &icon,
			ProjectsCount: 1,
			AuditTargetID: "target-1",
			Description:   "Default organization",
			Type:          "Organization",
		},
		{
			ID:            "2-0",
			Key:           "SECOND",
			Name:          "Second Org",
			ProjectsCount: 0,
			Type:          "Organization",
		},
	}
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{orgs: orgs}}
	return NewAdminHandler(app.New(holder))
}

// TestOrganizationsBasic يتحقق من المطابقة الدقيقة لـ request22.txt:
// GET /api/admin/organizations?$top=1 تُرجع [{"id":"1-0","$type":"Organization"}].
func TestOrganizationsBasic(t *testing.T) {
	h := newOrganizationsHandler()
	req := httptest.NewRequest("GET", "/api/admin/organizations?$top=1", nil)
	rec := httptest.NewRecorder()

	h.Organizations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 organization, got %d", len(res))
	}
	if res[0]["id"] != "1-0" {
		t.Errorf("expected id 1-0, got %v", res[0]["id"])
	}
	if len(res[0]) != 2 {
		t.Errorf("expected exactly {id, $type}, got keys %v", res[0])
	}
	if res[0]["$type"] != "Organization" {
		t.Errorf("expected $type Organization, got %v", res[0]["$type"])
	}
}

// TestOrganizationsAdvancedFields يتحقق من الجلب الموسع للحقول مع مشاريع متداخلة
// (مطابق لـ hh.json و hh2.json).
func TestOrganizationsAdvancedFields(t *testing.T) {
	fieldsParam := "id,key,name,iconUrl,projectsCount,auditTargetId,description,projects(id,name,shortName,projectType(id),pinned,iconUrl,template,archived,restricted,hasArticles,team(id))"
	h := newOrganizationsHandler()
	req := httptest.NewRequest("GET", "/api/admin/organizations?fields="+fieldsParam+"&$top=101&sorting=natural", nil)
	rec := httptest.NewRecorder()

	h.Organizations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 organizations, got %d", len(res))
	}
	o := res[0]
	if o["id"] != "1-0" {
		t.Errorf("expected id 1-0, got %v", o["id"])
	}
	if o["key"] != "CFSksvCi5N6T06bFUtA8w" {
		t.Errorf("unexpected key %v", o["key"])
	}
	if o["name"] != "me" {
		t.Errorf("unexpected name %v", o["name"])
	}
	if o["iconUrl"] != "https://osm.youtrack.cloud/icon.png" {
		t.Errorf("unexpected iconUrl %v", o["iconUrl"])
	}
	if o["projectsCount"] != float64(1) {
		t.Errorf("unexpected projectsCount %v", o["projectsCount"])
	}
	if o["auditTargetId"] != "target-1" {
		t.Errorf("unexpected auditTargetId %v", o["auditTargetId"])
	}
	if o["description"] != "Default organization" {
		t.Errorf("unexpected description %v", o["description"])
	}
	if o["$type"] != "Organization" {
		t.Errorf("expected $type Organization, got %v", o["$type"])
	}
	projects, ok := o["projects"].([]any)
	if !ok || len(projects) != 1 {
		t.Fatalf("expected 1 project, got %v", o["projects"])
	}
	p := projects[0].(map[string]any)
	if p["id"] != "22-59" {
		t.Errorf("unexpected project id %v", p["id"])
	}
	if p["shortName"] != "DEMO" {
		t.Errorf("unexpected project shortName %v", p["shortName"])
	}
	if p["$type"] != "Project" {
		t.Errorf("expected project $type Project, got %v", p["$type"])
	}
	if pt, ok := p["projectType"].(map[string]any); ok {
		if pt["id"] != "DEFAULT" || pt["$type"] != "ProjectType" {
			t.Errorf("unexpected projectType %v", p["projectType"])
		}
	} else {
		t.Errorf("expected projectType object")
	}
	if team, ok := p["team"].(map[string]any); ok {
		if team["id"] != "team-1" || team["$type"] != "ProjectTeam" {
			t.Errorf("unexpected team %v", p["team"])
		}
	} else {
		t.Errorf("expected team object")
	}
}

// TestOrganizationsNoFields يتحقق من أن حذف معامل fields يعرض id فقط مع $type.
func TestOrganizationsNoFields(t *testing.T) {
	h := newOrganizationsHandler()
	req := httptest.NewRequest("GET", "/api/admin/organizations?$top=-1", nil)
	rec := httptest.NewRecorder()

	h.Organizations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 organizations, got %d", len(res))
	}
	for _, o := range res {
		if len(o) != 2 {
			t.Errorf("expected only id and $type, got %v", o)
		}
	}
}

// TestOrganizationsSkipTop يتحقق من تأثير $skip و $top على التقسيم.
func TestOrganizationsSkipTop(t *testing.T) {
	h := newOrganizationsHandler()
	req := httptest.NewRequest("GET", "/api/admin/organizations?$skip=1&$top=1", nil)
	rec := httptest.NewRecorder()

	h.Organizations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 organization after skip, got %d", len(res))
	}
	if res[0]["id"] != "2-0" {
		t.Errorf("expected id 2-0 after skip, got %v", res[0]["id"])
	}
}

// TestOrganizationsRoute يتحقق من تسجيل المسار عبر الراوتر مع مصادقة JWT صالحة.
func TestOrganizationsRoute(t *testing.T) {
	holder := &mockAdminStoreHolder{admin: &mockAdminStore{orgs: []*model.Organization{{ID: "1-0", Type: "Organization"}}}}
	router := NewRouter(app.New(holder), "test-secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "2-1"})
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/organizations?$top=1", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 1 || res[0]["id"] != "1-0" {
		t.Errorf("unexpected organizations response %v", res)
	}
}
