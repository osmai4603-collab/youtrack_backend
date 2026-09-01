package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/store"
)

// mockSavedQueryStore يحاكي مخزن الاستعلامات المحفوظة.
type mockSavedQueryStore struct {
	items []*model.SavedQuery
}

func (m *mockSavedQueryStore) SavedQueries(ctx context.Context, tree *fields.FieldTree, top int, skip int) ([]*model.SavedQuery, error) {
	return m.items, nil
}

// mockSavedQueryFullStore يحقق واجهة store.Store للاختبار.
type mockSavedQueryFullStore struct {
	savedQueries *mockSavedQueryStore
}

func (m *mockSavedQueryFullStore) Users() store.UserStore                    { return nil }
func (m *mockSavedQueryFullStore) Projects() store.ProjectStore              { return nil }
func (m *mockSavedQueryFullStore) Issues() store.IssueStore                  { return nil }
func (m *mockSavedQueryFullStore) Admin() store.AdminStore                   { return nil }
func (m *mockSavedQueryFullStore) Inbox() store.InboxStore                   { return nil }
func (m *mockSavedQueryFullStore) SavedQueries() store.SavedQueryStore       { return m.savedQueries }
func (m *mockSavedQueryFullStore) Search() store.SearchStore                 { return nil }
func (m *mockSavedQueryFullStore) Subscriptions() store.SubscriptionStore    { return nil }
func (m *mockSavedQueryFullStore) SecuritySearch() store.SecuritySearchStore { return nil }

// sampleSavedQueries يبني بيانات استعلامات محفوظة مشابهة لـ request15.txt.
func sampleSavedQueries() []*model.SavedQuery {
	return []*model.SavedQuery{
		{
			ID:                "71-24321",
			IssuesURL:         "/search/Assigned%20to%20me-24321",
			Name:              "Assigned to me",
			Query:             "for: me",
			PinnedByDefault:   true,
			Pinned:            true,
			PinnedInHelpdesk:  true,
			IsUpdatable:       false,
			IsDeletable:       false,
			IsShareable:       false,
			SortOrderSortable: false,
			OwnerID:           "313-0",
			Owner: &model.User{
				ID:              "313-0",
				Login:           "search_user_2100689168",
				Name:            "JetBrains YouTrack",
				FullName:        "JetBrains YouTrack",
				AvatarURL:       "/robotAvatar.svg",
				UserTypeID:      "STANDARD_USER",
				UserType:        &model.UserType{ID: "STANDARD_USER", Name: "Standard user", Type: "UserType"},
				IsEmailVerified: false,
				Guest:           false,
				Online:          true,
				Banned:          false,
				CanReadProfile:  true,
				IsLocked:        true,
				Type:            "User",
			},
			VisibleForID: "675-0",
			ReadSharingSettings: &model.SavedQuerySharingSettings{
				PermittedGroups: []*model.SavedQueryGroup{
					{
						UserGroup: model.UserGroup{
							ID:            "675-0",
							Name:          "All Users",
							GroupType:     "ALL_USERS",
							AllUsersGroup: true,
							IsUpdatable:   false,
							IsRemovable:   false,
							Type:          "AllUsersGroup",
						},
						TeamForProjectID: "",
					},
				},
				PermittedUsers: []*model.User{},
				Type:           "WatchFolderSharingSettings",
			},
			UpdateSharingSettings: &model.SavedQuerySharingSettings{
				PermittedGroups: []*model.SavedQueryGroup{},
				PermittedUsers:  []*model.User{},
				Type:            "WatchFolderSharingSettings",
			},
			Type: "SavedQuery",
		},
		{
			ID:                "71-24322",
			IssuesURL:         "/search/Commented%20by%20me-24322",
			Name:              "Commented by me",
			Query:             "commenter: me",
			PinnedByDefault:   true,
			Pinned:            true,
			PinnedInHelpdesk:  true,
			SortOrderSortable: true,
			OwnerID:           "313-0",
			Owner: &model.User{
				ID:   "313-0",
				Type: "User",
			},
			UpdateSharingSettings: &model.SavedQuerySharingSettings{
				PermittedGroups: []*model.SavedQueryGroup{},
				PermittedUsers:  []*model.User{},
				Type:            "WatchFolderSharingSettings",
			},
			Type: "SavedQuery",
		},
	}
}

// keysOf يعيد مفاتيح الخريطة مرتبة.
func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// request15Fields هو نص معامل fields الكامل من request15.txt.
const request15Fields = "id,issuesUrl,name,query,pinnedByDefault,pinned,pinnedInHelpdesk,isUpdatable,isDeletable,isShareable,owner(@permittedUsers),readSharingSettings(@updateSharingSettings),updateSharingSettings(@updateSharingSettings),sortOrder(isSortable);@updateSharingSettings:permittedGroups(id,name,$type(),auditTargetId,description,allUsersGroup,icon,teamForProject(id,name,icon),isUpdatable,isRemovable),permittedUsers(@permittedUsers);@permittedUsers:id,login,email,fullName,avatarUrl,userType(id,name),name,isEmailVerified,guest,online,banned,banBadge,canReadProfile,isLocked"

func TestSavedQueriesFullRequestFields(t *testing.T) {
	a := app.New()
	handler := NewSavedQueriesHandler(a)

	req := httptest.NewRequest("GET", "/api/savedQueries?fields="+request15Fields, nil)
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 items, got %d", len(res))
	}

	// المستوى الأول: الحقول المطلوبة بالضبط + $type
	first := res[0]
	wantRoot := []string{
		"$type", "id", "issuesUrl", "name", "query", "pinnedByDefault", "pinned",
		"pinnedInHelpdesk", "isUpdatable", "isDeletable", "isShareable",
		"owner", "readSharingSettings", "updateSharingSettings", "sortOrder",
	}
	if got := keysOf(first); !equalKeys(got, wantRoot) {
		t.Errorf("root keys mismatch:\n got  %v\n want %v", got, wantRoot)
	}
	if first["$type"] != "SavedQuery" {
		t.Errorf("expected $type=SavedQuery, got %v", first["$type"])
	}

	// الحقل sortOrder
	sortOrder, ok := first["sortOrder"].(map[string]any)
	if !ok {
		t.Fatal("sortOrder is not an object")
	}
	wantSort := []string{"$type", "isSortable"}
	if got := keysOf(sortOrder); !equalKeys(got, wantSort) {
		t.Errorf("sortOrder keys mismatch: got %v want %v", got, wantSort)
	}
	if sortOrder["$type"] != "SavedQueryIssueOrder" {
		t.Errorf("expected sortOrder $type=SavedQueryIssueOrder, got %v", sortOrder["$type"])
	}

	// كائن owner يطابق قالب @permittedUsers بالضبط
	owner, ok := first["owner"].(map[string]any)
	if !ok {
		t.Fatal("owner is not an object")
	}
	wantOwner := []string{
		"$type", "id", "login", "email", "fullName", "avatarUrl", "userType",
		"name", "isEmailVerified", "guest", "online", "banned", "banBadge",
		"canReadProfile", "isLocked",
	}
	if got := keysOf(owner); !equalKeys(got, wantOwner) {
		t.Errorf("owner keys mismatch: got %v\n want %v", got, wantOwner)
	}
	if owner["$type"] != "User" {
		t.Errorf("expected owner $type=User, got %v", owner["$type"])
	}
	userType, ok := owner["userType"].(map[string]any)
	if !ok {
		t.Fatal("owner.userType is not an object")
	}
	wantUserType := []string{"$type", "id", "name"}
	if got := keysOf(userType); !equalKeys(got, wantUserType) {
		t.Errorf("userType keys mismatch: got %v want %v", got, wantUserType)
	}

	// إعدادات المشاركة: permittedGroups + permittedUsers + $type بالضبط
	readSettings, ok := first["readSharingSettings"].(map[string]any)
	if !ok {
		t.Fatal("readSharingSettings is not an object")
	}
	wantSettings := []string{"$type", "permittedGroups", "permittedUsers"}
	if got := keysOf(readSettings); !equalKeys(got, wantSettings) {
		t.Errorf("readSharingSettings keys mismatch: got %v want %v", got, wantSettings)
	}
	if readSettings["$type"] != "WatchFolderSharingSettings" {
		t.Errorf("expected readSharingSettings $type=WatchFolderSharingSettings, got %v", readSettings["$type"])
	}
	if users, ok := readSettings["permittedUsers"].([]any); !ok || len(users) != 0 {
		t.Errorf("permittedUsers should be empty array, got %v", readSettings["permittedUsers"])
	}

	// مجموعة مسموحة (All Users): الحقول الـ 9 + teamForProject + $type
	groups, ok := readSettings["permittedGroups"].([]any)
	if !ok || len(groups) != 1 {
		t.Fatalf("expected 1 permittedGroup, got %v", readSettings["permittedGroups"])
	}
	group := groups[0].(map[string]any)
	wantGroup := []string{
		"$type", "id", "name", "auditTargetId", "description", "allUsersGroup",
		"icon", "teamForProject", "isUpdatable", "isRemovable",
	}
	if got := keysOf(group); !equalKeys(got, wantGroup) {
		t.Errorf("group keys mismatch: got %v\n want %v", got, wantGroup)
	}
	if group["$type"] != "AllUsersGroup" {
		t.Errorf("expected group $type=AllUsersGroup, got %v", group["$type"])
	}
}

func TestSavedQueriesSubsetFields(t *testing.T) {
	a := app.New()
	handler := NewSavedQueriesHandler(a)

	req := httptest.NewRequest("GET", "/api/savedQueries?fields=id,name", nil)
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	handler.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(res) == 0 {
		t.Fatal("expected result")
	}

	// لا يجب أن تظهر الحقول غير المطلوبة
	want := []string{"$type", "id", "name"}
	if got := keysOf(res[0]); !equalKeys(got, want) {
		t.Errorf("subset keys mismatch: got %v want %v", got, want)
	}

	for _, forbidden := range []string{"issuesUrl", "query", "pinned", "owner", "readSharingSettings", "updateSharingSettings", "sortOrder", "folderId"} {
		if _, exists := res[0][forbidden]; exists {
			t.Errorf("field %q should NOT appear in subset response", forbidden)
		}
	}
}

func TestSavedQueriesDefaultFields(t *testing.T) {
	a := app.New()
	handler := NewSavedQueriesHandler(a)

	req := httptest.NewRequest("GET", "/api/savedQueries", nil)
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	handler.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(res) == 0 {
		t.Fatal("expected result")
	}

	// بدون fields نُعيد كل الحقول الافتراضية
	want := []string{
		"$type", "id", "issuesUrl", "name", "query", "pinnedByDefault", "pinned",
		"pinnedInHelpdesk", "isUpdatable", "isDeletable", "isShareable",
		"owner", "readSharingSettings", "updateSharingSettings", "sortOrder",
	}
	if got := keysOf(res[0]); !equalKeys(got, want) {
		t.Errorf("default keys mismatch: got %v\n want %v", got, want)
	}
}

// equalKeys يقارن مفتاحين بعد الفرز.
func equalKeys(a, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
