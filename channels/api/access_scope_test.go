package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/store"
)

// ملف الاختبارات هذا يغطي تصفية المشاريع والقضايا بحسب المستخدم الحالي
// (بيانات الاختبار: test1 يملك T1A/T1B، test2 يملك T2A/T2B، test3 لا يملك شيئاً).

// scopeProjectStore يخزّن المشاريع لكل مستخدم ويسجّل معرّف المستخدم الذي طُلبت منه القائمة.
type scopeProjectStore struct {
	byUser    map[string][]*model.Project
	lastUser  string
	allowByID map[string]string // project id -> user id يملكه
}

func (m *scopeProjectStore) GetByID(ctx context.Context, id string) (*model.Project, error) {
	return &model.Project{ID: id, Name: id, ShortName: id, Type: "Project"}, nil
}
func (m *scopeProjectStore) GetByShortName(ctx context.Context, shortName string) (*model.Project, error) {
	return &model.Project{ID: shortName, Name: shortName, ShortName: shortName, Type: "Project"}, nil
}
func (m *scopeProjectStore) All(ctx context.Context) ([]*model.Project, error) {
	all := []*model.Project{}
	for _, list := range m.byUser {
		all = append(all, list...)
	}
	return all, nil
}

func (m *scopeProjectStore) AllByUser(ctx context.Context, userID string) ([]*model.Project, error) {
	m.lastUser = userID
	return m.byUser[userID], nil
}

func (m *scopeProjectStore) CanAccessProject(ctx context.Context, userID string, projectRef string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	owner, ok := m.allowByID[projectRef]
	return ok && owner == userID, nil
}

func (m *scopeProjectStore) GetDetailed(ctx context.Context, id string, tree *fields.FieldTree) (*model.Project, error) {
	if _, ok := m.allowByID[id]; !ok {
		return nil, pgx.ErrNoRows
	}
	return &model.Project{ID: id, Name: id, ShortName: id, Type: "Project"}, nil
}

func (m *scopeProjectStore) GetProjectTeamAndLeader(ctx context.Context, id string, leaderTree, teamTree *fields.FieldTree) (*model.ProjectTeamAndLeader, error) {
	return &model.ProjectTeamAndLeader{Type: "Project"}, nil
}

// scopeIssueStore يخزّن القضايا لكل مستخدم بالمعرّف المقروء.
type scopeIssueStore struct {
	byUser    map[string]map[string]*model.Issue
	lastUser  string
	queries   []string
	allowByID map[string]string
}

func (m *scopeIssueStore) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	for _, issues := range m.byUser {
		if i, ok := issues[id]; ok {
			return i, nil
		}
	}
	return nil, pgx.ErrNoRows
}
func (m *scopeIssueStore) GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error) {
	return m.GetByID(ctx, idReadable)
}

func (m *scopeIssueStore) All(ctx context.Context, query string, limit int) ([]*model.Issue, error) {
	all := []*model.Issue{}
	for _, issues := range m.byUser {
		for _, i := range issues {
			all = append(all, i)
		}
	}
	return all, nil
}

func (m *scopeIssueStore) AllByUser(ctx context.Context, userID string, query string, limit int) ([]*model.Issue, error) {
	m.lastUser = userID
	m.queries = append(m.queries, query)
	out := []*model.Issue{}
	for _, i := range m.byUser[userID] {
		out = append(out, i)
	}
	return out, nil
}

func (m *scopeIssueStore) CanAccessIssue(ctx context.Context, userID string, issueRef string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	owner, ok := m.allowByID[issueRef]
	return ok && owner == userID, nil
}

func (m *scopeIssueStore) Create(ctx context.Context, i *model.Issue) error { return nil }
func (m *scopeIssueStore) Update(ctx context.Context, i *model.Issue) error { return nil }
func (m *scopeIssueStore) Delete(ctx context.Context, id string) error      { return nil }
func (m *scopeIssueStore) Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error) {
	return nil, nil
}
func (m *scopeIssueStore) CreateComment(ctx context.Context, c *model.IssueComment) error { return nil }
func (m *scopeIssueStore) Tags(ctx context.Context, issueID string) ([]*model.Tag, error) {
	return nil, nil
}
func (m *scopeIssueStore) Links(ctx context.Context, issueID string) ([]*model.IssueLink, error) {
	return nil, nil
}
func (m *scopeIssueStore) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error) {
	return nil, nil
}
func (m *scopeIssueStore) GetIssueCount(ctx context.Context, folderID string, query string, unresolvedOnly bool) (*model.IssueCountResponse, error) {
	return &model.IssueCountResponse{}, nil
}
func (m *scopeIssueStore) GetIssuesGetter(ctx context.Context, refs []string, query string, top int, skip int, tree *fields.FieldTree) ([]*model.IssueGetterIssue, error) {
	return nil, nil
}

type scopeStoreHolder struct {
	mockStoreBase
	projects *scopeProjectStore
	issues   *scopeIssueStore
}

func (m *scopeStoreHolder) Projects() store.ProjectStore { return m.projects }
func (m *scopeStoreHolder) Issues() store.IssueStore     { return m.issues }

func newScopeStore() *scopeStoreHolder {
	return &scopeStoreHolder{
		projects: &scopeProjectStore{
			byUser: map[string][]*model.Project{
				"test1": {
					{ID: "proj-t1-a", Name: "Test1 Project Alpha", ShortName: "T1A", Type: "Project"},
					{ID: "proj-t1-b", Name: "Test1 Project Beta", ShortName: "T1B", Type: "Project"},
				},
				"test2": {
					{ID: "proj-t2-a", Name: "Test2 Project Alpha", ShortName: "T2A", Type: "Project"},
					{ID: "proj-t2-b", Name: "Test2 Project Beta", ShortName: "T2B", Type: "Project"},
				},
			},
			allowByID: map[string]string{
				"proj-t1-a": "test1",
				"proj-t1-b": "test1",
				"proj-t2-a": "test2",
				"proj-t2-b": "test2",
			},
		},
		issues: &scopeIssueStore{
			byUser: map[string]map[string]*model.Issue{
				"test1": {
					"T1A-1": {ID: "iss-t1a-1", IDReadable: "T1A-1", Summary: "Setup CI pipeline", ProjectID: "proj-t1-a", Type: "Issue"},
					"T1B-1": {ID: "iss-t1b-1", IDReadable: "T1B-1", Summary: "Design landing page", ProjectID: "proj-t1-b", Type: "Issue"},
				},
				"test2": {
					"T2A-1": {ID: "iss-t2a-1", IDReadable: "T2A-1", Summary: "Database migration", ProjectID: "proj-t2-a", Type: "Issue"},
				},
			},
			allowByID: map[string]string{
				"iss-t1a-1": "test1",
				"T1A-1":     "test1",
				"iss-t1b-1": "test1",
				"T1B-1":     "test1",
				"iss-t2a-1": "test2",
				"T2A-1":     "test2",
			},
		},
	}
}

func scopedRequest(t *testing.T, method, target, userID string, roles []string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	if userID != "" {
		req = req.WithContext(withUserID(req.Context(), userID))
	}
	if len(roles) > 0 {
		req = req.WithContext(withUserRoles(req.Context(), roles))
	}
	return req
}

// TestProjectsListIsScopedToCurrentUser يتحقق أن كل مستخدم يرى مشاريعه فقط.
func TestProjectsListIsScopedToCurrentUser(t *testing.T) {
	for _, tc := range []struct {
		userID string
		want   []string
	}{
		{"test1", []string{"T1A", "T1B"}},
		{"test2", []string{"T2A", "T2B"}},
		{"test3", []string{}},
		{"", []string{}},
	} {
		st := newScopeStore()
		a := newTestApp(st)
		handler := NewProjectHandler(a)

		rec := httptest.NewRecorder()
		handler.List(rec, scopedRequest(t, http.MethodGet, "/api/projects", tc.userID, nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("user %q: expected 200, got %d: %s", tc.userID, rec.Code, rec.Body.String())
		}
		if st.projects.lastUser != tc.userID {
			t.Errorf("expected store to receive userID %q, got %q", tc.userID, st.projects.lastUser)
		}

		var res ProjectListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("user %q: parse response: %v", tc.userID, err)
		}
		if len(res.Projects) != len(tc.want) {
			t.Fatalf("user %q: expected %d projects, got %d", tc.userID, len(tc.want), len(res.Projects))
		}
		for i, shortName := range tc.want {
			if res.Projects[i].ShortName != shortName {
				t.Errorf("user %q: expected project %d = %s, got %s", tc.userID, i, shortName, res.Projects[i].ShortName)
			}
		}
	}
}

// TestProjectGetByIDRejectsOtherUserProject يتحقق من 404 عند طلب مشروع لا يملكه المستخدم.
func TestProjectGetByIDRejectsOtherUserProject(t *testing.T) {
	cases := []struct {
		name     string
		userID   string
		project  string
		wantCode int
	}{
		{"owner reads own project", "test1", "proj-t1-a", http.StatusOK},
		{"other user is rejected", "test2", "proj-t1-a", http.StatusNotFound},
		{"user without projects is rejected", "test3", "proj-t1-a", http.StatusNotFound},
		{"anonymous is rejected", "", "proj-t1-a", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := newScopeStore()
			a := newTestApp(st)
			handler := NewProjectHandler(a)

			req := scopedRequest(t, http.MethodGet, "/api/projects/"+tc.project, tc.userID, nil)
			req.SetPathValue("id", tc.project)
			rec := httptest.NewRecorder()

			handler.GetByID(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("expected %d, got %d: %s", tc.wantCode, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestIssuesListIsScopedToCurrentUser يتحقق أن كل مستخدم يرى قضايا مشاريعه فقط.
func TestIssuesListIsScopedToCurrentUser(t *testing.T) {
	for _, tc := range []struct {
		userID string
		want   []string
	}{
		{"test1", []string{"T1A-1", "T1B-1"}},
		{"test2", []string{"T2A-1"}},
		{"test3", []string{}},
		{"", []string{}},
	} {
		st := newScopeStore()
		a := newTestApp(st)
		handler := NewIssueHandler(a)

		rec := httptest.NewRecorder()
		handler.List(rec, scopedRequest(t, http.MethodGet, "/api/issues", tc.userID, nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("user %q: expected 200, got %d: %s", tc.userID, rec.Code, rec.Body.String())
		}
		if st.issues.lastUser != tc.userID {
			t.Errorf("expected store to receive userID %q, got %q", tc.userID, st.issues.lastUser)
		}

		var res IssueListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("user %q: parse response: %v", tc.userID, err)
		}
		if len(res.Issues) != len(tc.want) {
			t.Fatalf("user %q: expected %d issues, got %d", tc.userID, len(tc.want), len(res.Issues))
		}
		got := make([]string, 0, len(res.Issues))
		for _, i := range res.Issues {
			got = append(got, i.IDReadable)
		}
		want := append([]string(nil), tc.want...)
		sort.Strings(got)
		sort.Strings(want)
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("user %q: expected issues %v, got %v", tc.userID, want, got)
				break
			}
		}
	}
}

// TestIssueGetByIDRejectsOtherUserIssue يتحقق من 404 عند طلب قضية في مشروع لا يملكه المستخدم.
func TestIssueGetByIDRejectsOtherUserIssue(t *testing.T) {
	cases := []struct {
		name     string
		userID   string
		issue    string
		wantCode int
	}{
		{"owner reads own issue", "test1", "T1A-1", http.StatusOK},
		{"other user is rejected", "test2", "T1A-1", http.StatusNotFound},
		{"user without projects is rejected", "test3", "T1A-1", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := newScopeStore()
			a := newTestApp(st)
			handler := NewIssueHandler(a)

			req := scopedRequest(t, http.MethodGet, "/api/issues/"+tc.issue, tc.userID, nil)
			req.SetPathValue("id", tc.issue)
			rec := httptest.NewRecorder()

			handler.GetByID(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("expected %d, got %d: %s", tc.wantCode, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestAdminSessionSeesAllProjectsAndIssues يتحقق أن الجلسة الإدارية لا تتأثر بالتصفية.
func TestAdminSessionSeesAllProjectsAndIssues(t *testing.T) {
	st := newScopeStore()
	a := newTestApp(st)

	rec := httptest.NewRecorder()
	NewProjectHandler(a).List(rec, scopedRequest(t, http.MethodGet, "/api/projects", "test3", []string{"system_admin"}))
	var projects ProjectListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &projects); err != nil {
		t.Fatalf("parse projects response: %v", err)
	}
	if len(projects.Projects) != 4 {
		t.Errorf("admin should see 4 projects, got %d", len(projects.Projects))
	}
	if st.projects.lastUser != "" {
		t.Errorf("admin must not be filtered by user, but store got userID %q", st.projects.lastUser)
	}

	rec = httptest.NewRecorder()
	NewIssueHandler(a).List(rec, scopedRequest(t, http.MethodGet, "/api/issues", "test3", []string{"system_admin"}))
	var issues IssueListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatalf("parse issues response: %v", err)
	}
	if len(issues.Issues) != 3 {
		t.Errorf("admin should see 3 issues, got %d", len(issues.Issues))
	}
	if st.issues.lastUser != "" {
		t.Errorf("admin must not be filtered by user, but store got userID %q", st.issues.lastUser)
	}
}

// TestIssueQueryIsForwardedToScopedStore يتحقق أن نص البحث يصل للمخزن بعد تطبيق النطاق.
func TestIssueQueryIsForwardedToScopedStore(t *testing.T) {
	st := newScopeStore()
	a := newTestApp(st)

	req := httptest.NewRequest(http.MethodGet, "/api/issues?query=database", nil)
	req = req.WithContext(withUserID(req.Context(), "test1"))
	rec := httptest.NewRecorder()

	NewIssueHandler(a).List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(st.issues.queries) != 1 || st.issues.queries[0] != "database" {
		t.Errorf("expected query \"database\" forwarded to store, got %v", st.issues.queries)
	}
}
