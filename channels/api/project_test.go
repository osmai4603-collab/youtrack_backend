package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/store"
)

// mockProjectStore يحاكي مخزن مشاريع مع بيانات مشروع DEMO الكامل (Request #21).
type mockProjectStore struct {
	project *model.Project
}

func (m *mockProjectStore) GetByID(ctx context.Context, id string) (*model.Project, error) {
	return m.project, nil
}
func (m *mockProjectStore) GetByShortName(ctx context.Context, shortName string) (*model.Project, error) {
	return m.project, nil
}
func (m *mockProjectStore) All(ctx context.Context) ([]*model.Project, error) {
	return []*model.Project{m.project}, nil
}
func (m *mockProjectStore) AllByUser(ctx context.Context, userID string) ([]*model.Project, error) {
	return []*model.Project{m.project}, nil
}
func (m *mockProjectStore) CanAccessProject(ctx context.Context, userID string, projectRef string) (bool, error) {
	return true, nil
}

// GetDetailed يعيد مشروع DEMO كامل مع احترام شجرة الحقول، ويعيد pgx.ErrNoRows
// للمعرّف "missing" لمحاكاة حالة 404.
func (m *mockProjectStore) GetDetailed(ctx context.Context, id string, tree *fields.FieldTree) (*model.Project, error) {
	if id == "missing" {
		return nil, pgx.ErrNoRows
	}

	clone := cloneProject(m.project)

	// Leader
	if tree != nil && !tree.Has("leader") {
		clone.Leader = nil
	}

	// Organization
	if tree != nil && !tree.Has("organization") {
		clone.Organization = nil
	}

	// Team
	if tree != nil && tree.Has("team") {
		teamTree := tree.Child("team")
		if clone.Team != nil {
			if teamTree != nil && !teamTree.Has("users") {
				clone.Team.Users = nil
			}
			if teamTree != nil && !teamTree.Has("teamForProject") {
				clone.Team.TeamForProject = nil
			}
		}
	} else {
		clone.Team = nil
	}

	// Widgets
	if tree != nil && !tree.Has("widgets") {
		clone.Widgets = nil
	}

	// Plugins
	if tree != nil && tree.Has("plugins") {
		pluginsTree := tree.Child("plugins")
		if clone.Plugins != nil && pluginsTree != nil {
			if !pluginsTree.Has("vcsIntegrationSettings") {
				clone.Plugins.VcsIntegrationSettings = nil
			}
			if !pluginsTree.Has("timeTrackingSettings") {
				clone.Plugins.TimeTrackingSettings = nil
			}
			if !pluginsTree.Has("grazie") {
				clone.Plugins.Grazie = nil
			}
			if !pluginsTree.Has("helpDeskSettings") {
				clone.Plugins.HelpDeskSettings = nil
			}
		}
	} else {
		clone.Plugins = nil
	}

	// Visibility groups
	if tree != nil && !tree.Has("defaultVisibilityGroup") {
		clone.DefaultVisibilityGroup = nil
	}
	if tree != nil && !tree.Has("relevantVisibilityGroups") {
		clone.RelevantVisibilityGroups = nil
	}

	clone.Normalize()
	return clone, nil
}

// GetProjectTeamAndLeader يعيد بيانات القائد والفريق (الطلب #27) مطابقة لـ request27.txt.
func (m *mockProjectStore) GetProjectTeamAndLeader(ctx context.Context, id string, leaderTree, teamTree *fields.FieldTree) (*model.ProjectTeamAndLeader, error) {
	if id == "missing" {
		return nil, pgx.ErrNoRows
	}
	gu := "guest"
	guAvatar := "/hub/api/rest/avatar/5dd5893c-e15b-4857-ba25-c37e1f959908?s=48"
	adminAvatar := "/hub/api/rest/avatar/0d1219ff-dae2-4722-aa17-00381eb6be68?s=48"
	osmAvatar := "/hub/api/rest/avatar/921211c1-9f61-4784-92ad-34250abd94c0?s=48"
	invAvatar := "/hub/api/rest/avatar/0df202be-b74d-4a94-b233-2ffea5b4bbe1?s=48"
	adminEmail := "osmflutterdeveloper@gmail.com"
	osmEmail := "osmai4603@gmail.com"
	invEmail := "osminventory@gmail.com"

	res := &model.ProjectTeamAndLeader{
		Leader: &model.User27{
			ID:        "2-1",
			Login:     "admin",
			Name:      "admin",
			AvatarURL: &adminAvatar,
			Email:     &adminEmail,
			Type:      "User",
		},
		Team: &model.Team27{
			Name: "Demo project Team",
			Type: "ProjectTeam",
			Users: []*model.User27{
				{ID: "2-0", Login: "guest", Name: "guest", AvatarURL: &guAvatar, Email: nil, Type: "User"},
				{ID: "2-1", Login: "admin", Name: "admin", AvatarURL: &adminAvatar, Email: &adminEmail, Type: "User"},
				{ID: "2-3", Login: "osm", Name: "osm", AvatarURL: &osmAvatar, Email: &osmEmail, Type: "User"},
				{ID: "2-4", Login: "osminventory", Name: "osminventory", AvatarURL: &invAvatar, Email: &invEmail, Type: "User"},
			},
		},
		Type: "Project",
	}
	_ = gu
	return res, nil
}

// cloneProject ينسخ المشروع بعمق لتجنب تأثير الاختبارات على بعضها البعض.
func cloneProject(p *model.Project) *model.Project {
	if p == nil {
		return nil
	}
	c := *p
	if p.Team != nil {
		t := *p.Team
		c.Team = &t
	}
	if p.Organization != nil {
		o := *p.Organization
		c.Organization = &o
	}
	if p.Plugins != nil {
		pl := *p.Plugins
		c.Plugins = &pl
	}
	if p.Leader != nil {
		l := *p.Leader
		c.Leader = &l
	}
	return &c
}

func newMockDemoProject() *model.Project {
	team := &model.ProjectTeamDetailed{
		ID:            "5-0",
		Name:          "Demo project Team",
		AuditTargetID: "5-0",
		Type:          "ProjectTeam",
		Users: []*model.User{
			{
				ID:       "2-4",
				Login:    "osminventory",
				Email:    "osminventory@gmail.com",
				Name:     "osminventory",
				FullName: "osminventory",
				Type:     "User",
				UserType: &model.UserType{ID: "STANDARD_USER", Name: "Standard user", Type: "UserType"},
			},
		},
	}

	return &model.Project{
		ID:            "0-0",
		Name:          "Demo project",
		ShortName:     "DEMO",
		Pinned:        true,
		Query:         "project: {Demo project}",
		CreationTime:  1784851998350,
		AuditTargetID: "0-0",
		Type:          "Project",
		ProjectType:   &model.ProjectType{ID: "DEFAULT", Type: "ProjectType"},
		Leader: &model.User{
			ID:       "2-4",
			Login:    "osminventory",
			Email:    "osminventory@gmail.com",
			Name:     "osminventory",
			FullName: "osminventory",
			Type:     "User",
			UserType: &model.UserType{ID: "STANDARD_USER", Name: "Standard user", Type: "UserType"},
		},
		Team:                   team,
		Widgets:                []*model.DashboardWidget{{ID: "w-1", Key: "issues", Name: "Issues", Type: "DashboardWidget"}},
		DefaultVisibilityGroup: nil,
		Plugins: &model.ProjectPlugins{
			Type: "ProjectPlugins",
			TimeTrackingSettings: &model.TimeTrackingSettings{
				ID: "195-0", Enabled: true, Type: "ProjectTimeTrackingSettings",
			},
			HelpDeskSettings: &model.HelpDeskSettings{
				ID: "209-0", Type: "ProjectHelpDeskSettings",
			},
			VcsIntegrationSettings: &model.VcsIntegrationSettings{
				HasVcsIntegrations: false, Type: "ProjectVcsIntegrationSettings",
			},
			Grazie: &model.GrazieSettings{
				Disabled: false, Type: "ProjectGraziePlugin",
			},
		},
	}
}

type mockProjectStoreHolder struct {
	mockStoreBase
	projectStore *mockProjectStore
}

func (m *mockProjectStoreHolder) Users() store.UserStore                    { return nil }
func (m *mockProjectStoreHolder) Projects() store.ProjectStore              { return m.projectStore }
func (m *mockProjectStoreHolder) Issues() store.IssueStore                  { return nil }
func (m *mockProjectStoreHolder) Admin() store.AdminStore                   { return nil }
func (m *mockProjectStoreHolder) Inbox() store.InboxStore                   { return nil }
func (m *mockProjectStoreHolder) SavedQueries() store.SavedQueryStore       { return nil }
func (m *mockProjectStoreHolder) Search() store.SearchStore                 { return nil }
func (m *mockProjectStoreHolder) Subscriptions() store.SubscriptionStore    { return nil }
func (m *mockProjectStoreHolder) SecuritySearch() store.SecuritySearchStore { return nil }

// TestProjectGetByID_Request21 يتحقق من مطابقة البنية الكاملة للاستجابة مع request21.
func TestProjectGetByID_Request21(t *testing.T) {
	a := newTestApp(&mockProjectStoreHolder{projectStore: &mockProjectStore{project: newMockDemoProject()}})
	handler := NewProjectHandler(a)

	fieldsParam := "id,name,shortName,pinned,query,creationTime,auditTargetId,leader(id,login),team(id,name,auditTargetId,users(id,login)),plugins(vcsIntegrationSettings(hasVcsIntegrations),timeTrackingSettings(id,enabled),grazie(disabled),helpDeskSettings(id))"
	req := httptest.NewRequest(http.MethodGet, "/api/admin/projects/DEMO?fields="+fieldsParam, nil)
	req.SetPathValue("id", "DEMO")
	req = req.WithContext(withUserID(req.Context(), "2-4"))
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res["$type"] != "Project" {
		t.Errorf("expected $type Project, got %v", res["$type"])
	}
	if res["id"] != "0-0" {
		t.Errorf("expected id 0-0, got %v", res["id"])
	}
	if res["shortName"] != "DEMO" {
		t.Errorf("expected shortName DEMO, got %v", res["shortName"])
	}

	// Team
	team, ok := res["team"].(map[string]any)
	if !ok {
		t.Fatalf("expected team object")
	}
	if team["$type"] != "ProjectTeam" {
		t.Errorf("expected team $type ProjectTeam, got %v", team["$type"])
	}
	if team["id"] != "5-0" {
		t.Errorf("expected team id 5-0, got %v", team["id"])
	}
	if _, ok := team["users"].([]any); !ok {
		t.Errorf("expected team users list")
	}

	// Plugins $types
	plugins, ok := res["plugins"].(map[string]any)
	if !ok {
		t.Fatalf("expected plugins object")
	}
	if plugins["$type"] != "ProjectPlugins" {
		t.Errorf("expected plugins $type ProjectPlugins, got %v", plugins["$type"])
	}
	if tt, ok := plugins["timeTrackingSettings"].(map[string]any); ok {
		if tt["$type"] != "ProjectTimeTrackingSettings" {
			t.Errorf("expected timeTrackingSettings $type ProjectTimeTrackingSettings, got %v", tt["$type"])
		}
		if tt["enabled"] != true {
			t.Errorf("expected timeTrackingSettings enabled true")
		}
	}
	if hd, ok := plugins["helpDeskSettings"].(map[string]any); ok {
		if hd["$type"] != "ProjectHelpDeskSettings" {
			t.Errorf("expected helpDeskSettings $type ProjectHelpDeskSettings, got %v", hd["$type"])
		}
	}
	if vcs, ok := plugins["vcsIntegrationSettings"].(map[string]any); ok {
		if vcs["$type"] != "ProjectVcsIntegrationSettings" {
			t.Errorf("expected vcsIntegrationSettings $type ProjectVcsIntegrationSettings, got %v", vcs["$type"])
		}
	}
	if gz, ok := plugins["grazie"].(map[string]any); ok {
		if gz["$type"] != "ProjectGraziePlugin" {
			t.Errorf("expected grazie $type ProjectGraziePlugin, got %v", gz["$type"])
		}
	}
}

// TestProjectGetByID_FieldFiltering يتحقق من حجب الحقول والكائنات غير المطلوبة.
func TestProjectGetByID_FieldFiltering(t *testing.T) {
	a := newTestApp(&mockProjectStoreHolder{projectStore: &mockProjectStore{project: newMockDemoProject()}})
	handler := NewProjectHandler(a)

	// نطلب فقط id و name و widgets
	req := httptest.NewRequest(http.MethodGet, "/api/projects/0-0?fields=id,name,widgets(id,key)", nil)
	req.SetPathValue("id", "0-0")
	req = req.WithContext(withUserID(req.Context(), "2-4"))
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &res)

	if res["id"] != "0-0" {
		t.Errorf("expected id 0-0")
	}
	if res["name"] != "Demo project" {
		t.Errorf("expected name Demo project")
	}
	// يجب ألا تكون هذه الحقول موجودة لأن لم تُطلب
	if res["team"] != nil {
		t.Errorf("team should be hidden, got %v", res["team"])
	}
	if res["plugins"] != nil {
		t.Errorf("plugins should be hidden, got %v", res["plugins"])
	}
	if res["leader"] != nil {
		t.Errorf("leader should be hidden, got %v", res["leader"])
	}
	if res["organization"] != nil {
		t.Errorf("organization should be hidden, got %v", res["organization"])
	}
	if res["widgets"] == nil {
		t.Errorf("widgets should be present")
	}
}

// TestProjectGetByID_PluginSubFieldFiltering يتحقق من حجب فرع غير مطلوب داخل plugins.
func TestProjectGetByID_PluginSubFieldFiltering(t *testing.T) {
	a := newTestApp(&mockProjectStoreHolder{projectStore: &mockProjectStore{project: newMockDemoProject()}})
	handler := NewProjectHandler(a)

	// نطلب plugins مع tylko timeTrackingSettings
	req := httptest.NewRequest(http.MethodGet, "/api/projects/0-0?fields=id,plugins(timeTrackingSettings(id,enabled))", nil)
	req.SetPathValue("id", "0-0")
	req = req.WithContext(withUserID(req.Context(), "2-4"))
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &res)

	plugins, ok := res["plugins"].(map[string]any)
	if !ok {
		t.Fatalf("expected plugins object")
	}
	if _, ok := plugins["timeTrackingSettings"].(map[string]any); !ok {
		t.Errorf("expected timeTrackingSettings to be present")
	}
	if plugins["vcsIntegrationSettings"] != nil {
		t.Errorf("vcsIntegrationSettings should be hidden")
	}
	if plugins["grazie"] != nil {
		t.Errorf("grazie should be hidden")
	}
	if plugins["helpDeskSettings"] != nil {
		t.Errorf("helpDeskSettings should be hidden")
	}
}

// TestProjectGetByID_NotFound يتحقق من إرجاع 404 عند عدم وجود المشروع.
func TestProjectGetByID_NotFound(t *testing.T) {
	a := newTestApp(&mockProjectStoreHolder{projectStore: &mockProjectStore{project: newMockDemoProject()}})
	handler := NewProjectHandler(a)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/projects/missing?fields=id,name", nil)
	req.SetPathValue("id", "missing")
	req = req.WithContext(withUserID(req.Context(), "2-4"))
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestProjectGetByID_Unauthorized يتحقق من حماية المسار عبر JWT عبر الـ router.
func TestProjectGetByID_Unauthorized(t *testing.T) {
	a := newTestApp(&mockProjectStoreHolder{projectStore: &mockProjectStore{project: newMockDemoProject()}})
	router := NewRouter(a, "test-secret")

	// بدون Authorization header
	req := httptest.NewRequest(http.MethodGet, "/api/admin/projects/DEMO?fields=id,name", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestProjectGetByID_Request27 يتحقق من مطابقة الاستجابة حرفياً مع request27.txt
// (لا حقول جذر إضافية، leader، team مع users، email = null للحساب الضيف).
func TestProjectGetByID_Request27(t *testing.T) {
	a := newTestApp(&mockProjectStoreHolder{projectStore: &mockProjectStore{project: newMockDemoProject()}})
	handler := NewProjectHandler(a)

	fieldsParam := "team(name,users(id,login,name,avatarUrl,email)),leader(id)"
	req := httptest.NewRequest(http.MethodGet, "/api/admin/projects/0-0?fields="+fieldsParam, nil)
	req.SetPathValue("id", "0-0")
	req = req.WithContext(withUserID(req.Context(), "2-1"))
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// 1. لا حقول جذر إضافية (id, name, shortName غير مطلوبة في الطلب 27)
	for _, k := range []string{"id", "name", "shortName"} {
		if _, exists := res[k]; exists {
			t.Errorf("root field %q should NOT be present, got %v", k, res[k])
		}
	}
	if res["$type"] != "Project" {
		t.Errorf("expected root $type Project, got %v", res["$type"])
	}

	// 2. leader
	leader, ok := res["leader"].(map[string]any)
	if !ok {
		t.Fatalf("expected leader object")
	}
	if leader["id"] != "2-1" {
		t.Errorf("expected leader.id 2-1, got %v", leader["id"])
	}
	if leader["$type"] != "User" {
		t.Errorf("expected leader $type User, got %v", leader["$type"])
	}
	if _, exists := leader["login"]; exists {
		t.Errorf("leader should only have id and $type, got login: %v", leader["login"])
	}

	// 3. team
	team, ok := res["team"].(map[string]any)
	if !ok {
		t.Fatalf("expected team object")
	}
	if team["$type"] != "ProjectTeam" {
		t.Errorf("expected team $type ProjectTeam, got %v", team["$type"])
	}
	if team["name"] != "Demo project Team" {
		t.Errorf("expected team.name 'Demo project Team', got %v", team["name"])
	}
	users, ok := team["users"].([]any)
	if !ok {
		t.Fatalf("expected team.users list")
	}
	if len(users) != 4 {
		t.Fatalf("expected 4 team users, got %d", len(users))
	}

	// 4. أول مستخدم (guest 2-0) لديه email = null ولا حقول زائدة
	gu := users[0].(map[string]any)
	if gu["id"] != "2-0" {
		t.Errorf("expected first user id 2-0, got %v", gu["id"])
	}
	if gu["login"] != "guest" {
		t.Errorf("expected login guest, got %v", gu["login"])
	}
	if v, exists := gu["email"]; !exists || v != nil {
		t.Errorf("expected guest email to be null, got %v", v)
	}
	if gu["$type"] != "User" {
		t.Errorf("expected user $type User, got %v", gu["$type"])
	}
	if _, exists := gu["banned"]; exists {
		t.Errorf("user should not contain non-requested field 'banned'")
	}

	// 5. مستخدم عادي 2-1 له email غير null
	second := users[1].(map[string]any)
	if second["email"] != "osmflutterdeveloper@gmail.com" {
		t.Errorf("expected admin email, got %v", second["email"])
	}
	if _, exists := second["avatarUrl"]; !exists {
		t.Errorf("expected avatarUrl present")
	}
}
