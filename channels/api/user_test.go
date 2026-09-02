package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/store"
)

type mockUserStore struct {
	user *model.User
}

func (m *mockUserStore) GetByID(ctx context.Context, id string) (*model.User, error) {
	return m.user, nil
}
func (m *mockUserStore) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	return m.user, nil
}
func (m *mockUserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.user, nil
}
func (m *mockUserStore) Create(ctx context.Context, u *model.User) error {
	return nil
}
func (m *mockUserStore) All(ctx context.Context) ([]*model.User, error) {
	return []*model.User{m.user}, nil
}
func (m *mockUserStore) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	return &model.UserProfile{UserID: userID}, nil
}
func (m *mockUserStore) CreateProfile(ctx context.Context, p *model.UserProfile) error {
	return nil
}
func (m *mockUserStore) GetMe(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error) {
	u := *m.user

	if tree == nil || tree.Has("featureFlags") {
		u.FeatureFlags = []*model.FeatureFlag{
			{ID: "jetbrains.youtrack.feature.inlineComments", Enabled: true, Type: "FeatureFlag"},
		}
	} else {
		u.FeatureFlags = nil
	}

	if tree == nil || tree.Has("widgets") {
		u.Widgets = []*model.DashboardWidget{}
	} else {
		u.Widgets = nil
	}

	if tree == nil || tree.Has("issueRelatedGroup") {
		u.IssueRelatedGroup = nil
	} else {
		u.IssueRelatedGroup = nil
	}

	if tree == nil || tree.Has("profiles") {
		profiles := model.DefaultUserProfiles()
		if tree != nil && tree.Child("profiles") != nil {
			profTree := tree.Child("profiles")
			if !profTree.Has("ai") {
				profiles.AI = nil
			}
			if !profTree.Has("tips") {
				profiles.Tips = nil
			}
			if !profTree.Has("helpdesk") {
				profiles.Helpdesk = nil
			}
			if !profTree.Has("timetracking") {
				profiles.Timetracking = nil
			}
			if !profTree.Has("general") {
				profiles.General = nil
			}
			if !profTree.Has("appearance") {
				profiles.Appearance = nil
			}
			if !profTree.Has("issuesList") {
				profiles.IssuesList = nil
			}
			if !profTree.Has("articles") {
				profiles.Articles = nil
			}
			if !profTree.Has("notifications") {
				profiles.Notifications = nil
			}
		}
		u.Profiles = profiles
	} else {
		u.Profiles = nil
	}

	u.NormalizeMe()
	return &u, nil
}
func (m *mockUserStore) GetFeatureFlags(ctx context.Context) ([]*model.FeatureFlag, error) {
	return []*model.FeatureFlag{
		{ID: "jetbrains.youtrack.feature.inlineComments", Enabled: true, Type: "FeatureFlag"},
	}, nil
}
func (m *mockUserStore) GetRecentIssues(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentIssue, error) {
	return []*model.RecentIssue{
		{
			ID:     "3-3",
			Pinned: false,
			Date:   1785771300229,
			Type:   "RecentIssue",
			Issue: &model.Issue{
				ID:         "3-3",
				IDReadable: "DEMO-4",
				Summary:    "Create a demo project",
				Type:       "Issue",
			},
		},
	}, nil
}
func (m *mockUserStore) GetRecentArticles(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentArticle, error) {
	return []*model.RecentArticle{
		{
			ID:     "186-0",
			Pinned: false,
			Date:   1785818696478,
			Type:   "RecentArticle",
		},
	}, nil
}
func (m *mockUserStore) GetGrazieProfile(ctx context.Context, userID string) (*model.GrazieUserProfile, error) {
	return &model.GrazieUserProfile{
		ExcludedIssueTypes:            "",
		EnableSpellChecker:            true,
		HasMoreTokens:                 true,
		EnableTextCompletion:          true,
		CycleRestart:                  1788271206218,
		FreeLicense:                   false,
		SpellCheckerEnabledInSystem:   true,
		TextCompletionEnabledInSystem: true,
		Enabled:                       false,
		Type:                          "GrazieUserProfile",
	}, nil
}
func (m *mockUserStore) GetGeneralProfile(ctx context.Context, userID string) (*model.GeneralUserProfile, error) {
	return &model.GeneralUserProfile{
		ID: "generalProfile",
		Timezone: &model.TimeZoneDescriptor{
			ID:   "Europe/Prague",
			Type: "TimeZoneDescriptor",
		},
		DateFormat: &model.DateFormatDescriptor{
			Pattern:     "d MMM yyyy HH:mm",
			DatePattern: "d MMM yyyy",
			Type:        "DateFormatDescriptor",
		},
		Locale: &model.LocaleDescriptor{
			Name:      "English",
			Community: false,
			Locale:    "en_US",
			ID:        "en_US",
			Language:  "en",
			Type:      "LocaleDescriptor",
		},
		SemanticSearchForArticles: false,
		LastCreatedIssue:          nil,
		SearchContext:             nil,
		HelpdeskContext:           nil,
		Type:                      "GeneralUserProfile",
	}, nil
}
func (m *mockUserStore) GetQuestionnaireProfile(ctx context.Context, userID string) (*model.QuestionnaireUserProfile, error) {
	return &model.QuestionnaireUserProfile{
		ShowSurvey:    false,
		ShowPmfSurvey: false,
		Type:          "QuestionnaireUserProfile",
	}, nil
}
func (m *mockUserStore) GetHubMe(ctx context.Context, userID string) (*model.HubUser, error) {
	return &model.HubUser{
		Guest: false,
		ID:    "0d1219ff-dae2-4722-aa17-00381eb6be68",
		Name:  "OSM ABASSI",
		Login: "osmflutterdeveloper",
		Type:  "User",
	}, nil
}
func (m *mockUserStore) InboxFolders(ctx context.Context, userID string) ([]*model.InboxFolder, error) {
	return []*model.InboxFolder{
		{ID: "direct", LastNotified: 0, LastSeen: 0, Enabled: true, Type: "InboxFolder"},
		{ID: "subscription", LastNotified: 0, LastSeen: 0, Enabled: true, Type: "InboxFolder"},
		{ID: "system", LastNotified: 0, LastSeen: 0, Enabled: false, Type: "InboxFolder"},
		{ID: "whatsnew", LastNotified: 0, LastSeen: 0, Enabled: true, Type: "InboxFolder"},
		{ID: "version_deploy", LastNotified: 1779755105649, LastSeen: 0, Enabled: true, Type: "InboxFolder"},
	}, nil
}

type mockFullUserStore struct {
	mockStoreBase
	userStore *mockUserStore
}

func (m *mockFullUserStore) Users() store.UserStore                    { return m.userStore }
func (m *mockFullUserStore) Projects() store.ProjectStore              { return nil }
func (m *mockFullUserStore) Issues() store.IssueStore                  { return nil }
func (m *mockFullUserStore) Admin() store.AdminStore                   { return nil }
func (m *mockFullUserStore) Inbox() store.InboxStore                   { return nil }
func (m *mockFullUserStore) SavedQueries() store.SavedQueryStore       { return nil }
func (m *mockFullUserStore) Search() store.SearchStore                 { return nil }
func (m *mockFullUserStore) Subscriptions() store.SubscriptionStore    { return nil }
func (m *mockFullUserStore) SecuritySearch() store.SecuritySearchStore { return nil }

func setupTestApp() (*app.YouTrackApp, *model.User) {
	user := &model.User{
		ID:              "11-2095841",
		Login:           "osmflutterdeveloper",
		Email:           "osmflutterdeveloper@gmail.com",
		FullName:        "OSM ABASSI",
		Name:            "OSM ABASSI",
		AvatarURL:       "https://hub.jetbrains.com/avatar.png",
		IsEmailVerified: true,
		Guest:           false,
		Online:          true,
		Banned:          false,
		CanReadProfile:  true,
		IsLocked:        false,
		UserType: &model.UserType{
			ID:   "STANDARD_USER",
			Name: "Standard user",
			Type: "UserType",
		},
	}
	user.NormalizeMe()
	a := newTestApp(&mockFullUserStore{userStore: &mockUserStore{user: user}})
	return a, user
}

func TestGetMeEndpoint(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	req := httptest.NewRequest("GET", "/api/users/me?fields=id,login,email,fullName,profiles(general(timezone(id)),appearance(compactMode))", nil)
	req = req.WithContext(withUserID(req.Context(), "11-2095841"))
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	// 1. Root entity type should be "Me"
	if res["$type"] != "Me" {
		t.Errorf("expected $type to be 'Me', got %v", res["$type"])
	}

	// 2. ID and login should be at root
	if res["id"] != "11-2095841" {
		t.Errorf("expected id '11-2095841', got %v", res["id"])
	}
	if res["login"] != "osmflutterdeveloper" {
		t.Errorf("expected login 'osmflutterdeveloper', got %v", res["login"])
	}

	// 3. Profiles object should exist with requested sub-profiles
	profiles, ok := res["profiles"].(map[string]any)
	if !ok {
		t.Fatalf("expected profiles object in response")
	}

	if profiles["general"] == nil {
		t.Errorf("expected general profile to be present")
	}
	if profiles["appearance"] == nil {
		t.Errorf("expected appearance profile to be present")
	}
}

func TestGetMeSelectiveFields(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	// 1. Request with only basic fields: no profiles, no featureFlags, no widgets
	{
		req := httptest.NewRequest("GET", "/api/users/me?fields=id,login,email", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()

		handler.GetMe(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)

		if res["id"] != "11-2095841" {
			t.Errorf("expected id 11-2095841")
		}
		if res["login"] != "osmflutterdeveloper" {
			t.Errorf("expected login osmflutterdeveloper")
		}
		if res["profiles"] != nil {
			t.Errorf("profiles should not be present when not requested in fields")
		}
		if res["featureFlags"] != nil {
			t.Errorf("featureFlags should not be present when not requested in fields")
		}
		if res["widgets"] != nil {
			t.Errorf("widgets should not be present when not requested in fields")
		}
	}

	// 2. Request with specific sub-profile: only profiles(appearance)
	{
		req := httptest.NewRequest("GET", "/api/users/me?fields=id,profiles(appearance(compactMode))", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()

		handler.GetMe(rec, req)

		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)

		profiles, ok := res["profiles"].(map[string]any)
		if !ok {
			t.Fatalf("expected profiles object in response")
		}
		if profiles["appearance"] == nil {
			t.Errorf("expected appearance in profiles")
		}
		if profiles["ai"] != nil {
			t.Errorf("ai profile should not be present when not requested")
		}
		if profiles["tips"] != nil {
			t.Errorf("tips profile should not be present when not requested")
		}
		if profiles["notifications"] != nil {
			t.Errorf("notifications profile should not be present when not requested")
		}
	}
}

func TestGetSubProfiles(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	// 1. Test Grazie Profile
	{
		req := httptest.NewRequest("GET", "/api/users/me/profiles/grazie", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetGrazieProfile(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for grazie, got %d", rec.Code)
		}
		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["$type"] != "GrazieUserProfile" {
			t.Errorf("expected $type 'GrazieUserProfile', got %v", res["$type"])
		}
		if res["enableSpellChecker"] != true {
			t.Errorf("expected enableSpellChecker true")
		}
	}

	// 2. Test General Profile
	{
		req := httptest.NewRequest("GET", "/api/users/me/profiles/general", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetGeneralProfile(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for general profile, got %d", rec.Code)
		}
		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["$type"] != "GeneralUserProfile" {
			t.Errorf("expected $type 'GeneralUserProfile', got %v", res["$type"])
		}
	}

	// 3. Test Questionnaire Profile
	{
		req := httptest.NewRequest("GET", "/api/users/me/profiles/questionnaire", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetQuestionnaireProfile(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for questionnaire, got %d", rec.Code)
		}
		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["$type"] != "QuestionnaireUserProfile" {
			t.Errorf("expected $type 'QuestionnaireUserProfile', got %v", res["$type"])
		}
	}

	// 4. Test Recent Issues
	{
		req := httptest.NewRequest("GET", "/api/users/me/recent/issues", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetRecentIssues(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for recent issues, got %d", rec.Code)
		}
		var res []any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if len(res) == 0 {
			t.Errorf("expected non-empty recent issues array")
		}
	}

	// 5. Test Hub Me
	{
		req := httptest.NewRequest("GET", "/hub/api/rest/users/me", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetHubMe(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for hub me, got %d", rec.Code)
		}
		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["login"] != "osmflutterdeveloper" {
			t.Errorf("expected login 'osmflutterdeveloper', got %v", res["login"])
		}
	}
}

func TestGetGeneralProfileSelectiveFields(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	// 1. Request only id and dateFieldFormat(pattern,datePattern)
	{
		req := httptest.NewRequest("GET", "/api/users/me/profiles/general?fields=id,dateFieldFormat(pattern,datePattern)", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetGeneralProfile(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json response: %v", err)
		}

		if res["$type"] != "GeneralUserProfile" {
			t.Errorf("expected $type 'GeneralUserProfile', got %v", res["$type"])
		}
		if res["id"] != "generalProfile" {
			t.Errorf("expected id 'generalProfile', got %v", res["id"])
		}
		// timezone, locale, semanticSearchForArticles should NOT be present
		if _, present := res["timezone"]; present {
			t.Errorf("expected timezone to be absent with fields=id,dateFieldFormat")
		}
		if _, present := res["locale"]; present {
			t.Errorf("expected locale to be absent with fields=id,dateFieldFormat")
		}
		if _, present := res["semanticSearchForArticles"]; present {
			t.Errorf("expected semanticSearchForArticles to be absent with fields=id,dateFieldFormat")
		}

		// dateFieldFormat should contain only pattern and datePattern
		dff, ok := res["dateFieldFormat"].(map[string]any)
		if !ok {
			t.Fatalf("expected dateFieldFormat to be an object")
		}
		if dff["pattern"] != "d MMM yyyy HH:mm" {
			t.Errorf("expected pattern 'd MMM yyyy HH:mm', got %v", dff["pattern"])
		}
		if dff["datePattern"] != "d MMM yyyy" {
			t.Errorf("expected datePattern 'd MMM yyyy', got %v", dff["datePattern"])
		}
	}

	// 2. Request only dateFieldFormat without nested fields -> full dateFieldFormat object
	{
		req := httptest.NewRequest("GET", "/api/users/me/profiles/general?fields=dateFieldFormat", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetGeneralProfile(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)

		dff, ok := res["dateFieldFormat"].(map[string]any)
		if !ok {
			t.Fatalf("expected dateFieldFormat to be an object")
		}
		if dff["pattern"] != "d MMM yyyy HH:mm" {
			t.Errorf("expected pattern 'd MMM yyyy HH:mm', got %v", dff["pattern"])
		}
		if dff["datePattern"] != "d MMM yyyy" {
			t.Errorf("expected datePattern 'd MMM yyyy', got %v", dff["datePattern"])
		}
	}

	// 3. No fields parameter -> full general profile
	{
		req := httptest.NewRequest("GET", "/api/users/me/profiles/general", nil)
		req = req.WithContext(withUserID(req.Context(), "11-2095841"))
		rec := httptest.NewRecorder()
		handler.GetGeneralProfile(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &res)

		if res["id"] != "generalProfile" {
			t.Errorf("expected id 'generalProfile', got %v", res["id"])
		}
		if _, present := res["timezone"]; !present {
			t.Errorf("expected timezone to be present with no fields parameter")
		}
		if _, present := res["locale"]; !present {
			t.Errorf("expected locale to be present with no fields parameter")
		}
		if _, present := res["semanticSearchForArticles"]; !present {
			t.Errorf("expected semanticSearchForArticles to be present with no fields parameter")
		}
	}
}

func TestUnauthorizedRequests(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	// No context user ID
	req := httptest.NewRequest("GET", "/api/users/me", nil)
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", rec.Code)
	}
}

func TestListAndGetByID(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	// Test List
	{
		req := httptest.NewRequest("GET", "/api/users", nil)
		rec := httptest.NewRecorder()
		handler.List(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for List, got %d", rec.Code)
		}
		var users []map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &users)
		if len(users) == 0 {
			t.Errorf("expected non-empty user list")
		}
	}

	// Test GetByID
	{
		req := httptest.NewRequest("GET", "/api/users/11-2095841", nil)
		req.SetPathValue("id", "11-2095841")
		rec := httptest.NewRecorder()
		handler.GetByID(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for GetByID, got %d", rec.Code)
		}
		var user map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &user)
		if user["login"] != "osmflutterdeveloper" {
			t.Errorf("expected login 'osmflutterdeveloper'")
		}
	}
}

func TestGetInboxFolders(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	req := httptest.NewRequest("GET", "/api/inbox/folders?fields=id,lastNotified,lastSeen,enabled", nil)
	req = req.WithContext(withUserID(req.Context(), "11-2095841"))
	rec := httptest.NewRecorder()

	handler.GetInboxFolders(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 5 {
		t.Fatalf("expected 5 inbox folders, got %d", len(res))
	}

	f0 := res[0]
	if f0["id"] != "direct" {
		t.Errorf("expected id 'direct', got %v", f0["id"])
	}
	if f0["$type"] != "InboxFolder" {
		t.Errorf("expected $type InboxFolder, got %v", f0["$type"])
	}
	if f0["enabled"] != true {
		t.Errorf("expected enabled true, got %v", f0["enabled"])
	}

	last := res[4]
	if last["lastNotified"] != float64(1779755105649) {
		t.Errorf("expected lastNotified 1779755105649, got %v", last["lastNotified"])
	}
}

func TestGetInboxFoldersSelectiveFields(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	req := httptest.NewRequest("GET", "/api/inbox/folders?fields=id,enabled", nil)
	req = req.WithContext(withUserID(req.Context(), "11-2095841"))
	rec := httptest.NewRecorder()

	handler.GetInboxFolders(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 5 {
		t.Fatalf("expected 5 inbox folders, got %d", len(res))
	}

	f0 := res[0]
	if f0["id"] != "direct" {
		t.Errorf("expected id 'direct', got %v", f0["id"])
	}
	if _, ok := f0["enabled"]; !ok {
		t.Errorf("expected enabled field to be present")
	}
	if f0["$type"] != "InboxFolder" {
		t.Errorf("expected $type InboxFolder, got %v", f0["$type"])
	}
	if _, present := f0["lastNotified"]; present {
		t.Errorf("expected lastNotified to be absent with fields=id,enabled")
	}
	if _, present := f0["lastSeen"]; present {
		t.Errorf("expected lastSeen to be absent with fields=id,enabled")
	}
}

func TestGetInboxFoldersNoFields(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	req := httptest.NewRequest("GET", "/api/inbox/folders", nil)
	req = req.WithContext(withUserID(req.Context(), "11-2095841"))
	rec := httptest.NewRecorder()

	handler.GetInboxFolders(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(res) != 5 {
		t.Fatalf("expected 5 inbox folders, got %d", len(res))
	}
	f0 := res[0]
	for _, key := range []string{"id", "lastNotified", "lastSeen", "enabled", "$type"} {
		if _, present := f0[key]; !present {
			t.Errorf("expected full schema to include %s, got %v", key, f0)
		}
	}
}

func TestGetInboxFoldersUnauthorized(t *testing.T) {
	a, _ := setupTestApp()
	handler := NewUserHandler(a)

	req := httptest.NewRequest("GET", "/api/inbox/folders", nil)
	rec := httptest.NewRecorder()

	handler.GetInboxFolders(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}
}
