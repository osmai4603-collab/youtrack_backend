package api

import (
	"bytes"
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

// mockSubscriptionStore يحاكي مخزن الاشتراكات في قوائم المشاكل.
type mockSubscriptionStore struct {
	sub *model.IssueListSubscriptionBean
}

func (m *mockSubscriptionStore) SubscribeIssueList(ctx context.Context, userID string, req *model.IssueListSubscriptionRequest, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	return m.sub, nil
}

func (m *mockSubscriptionStore) GetIssueListSubscriptionByTicket(ctx context.Context, userID string, ticket string, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	return m.sub, nil
}

// mockSubscriptionFullStore يحقق واجهة store.Store للاختبار.
type mockSubscriptionFullStore struct {
	subscriptions *mockSubscriptionStore
}

func (m *mockSubscriptionFullStore) Users() store.UserStore                    { return nil }
func (m *mockSubscriptionFullStore) Projects() store.ProjectStore              { return nil }
func (m *mockSubscriptionFullStore) Issues() store.IssueStore                  { return nil }
func (m *mockSubscriptionFullStore) Admin() store.AdminStore                   { return nil }
func (m *mockSubscriptionFullStore) Inbox() store.InboxStore                   { return nil }
func (m *mockSubscriptionFullStore) SavedQueries() store.SavedQueryStore       { return nil }
func (m *mockSubscriptionFullStore) Search() store.SearchStore                 { return nil }
func (m *mockSubscriptionFullStore) Subscriptions() store.SubscriptionStore    { return m.subscriptions }
func (m *mockSubscriptionFullStore) SecuritySearch() store.SecuritySearchStore { return nil }

// sampleSubscription يبني كائن اشتراك مشابهًا لـ request18.txt.
func sampleSubscription() *model.IssueListSubscriptionBean {
	return &model.IssueListSubscriptionBean{
		Ticket:      "g9qk4fe77jefpavefi1qacdhbefa3l",
		Query:       "issue id: DEMO-4",
		Subscribe:   true,
		ContextType: "Project",
		ContextID:   "0-0",
		Context: &model.SubscriptionContext{
			Type: "Project",
			ID:   "0-0",
		},
		Issues: []*model.IssueListSubscriptionItem{
			{ID: "3-3", Matches: true, Type: "IssueListSubscriptionItem"},
		},
		FolderID: "",
		Type:     "IssueListSubscriptionBean",
	}
}

func newSubscriptionHandler() *SubscriptionHandler {
	// sub := sampleSubscription()
	// full := &mockSubscriptionFullStore{subscriptions: &mockSubscriptionStore{sub: sub}}
	return NewSubscriptionHandler(app.New())
}

// TestIssueListSubscriptionRequest18 يتحقق من استجابة request18.txt:
// POST /api/issueListSubscription?fields=ticket
// {"ticket":"g9qk4fe77jefpavefi1qacdhbefa3l","$type":"IssueListSubscriptionBean"}
func TestIssueListSubscriptionRequest18(t *testing.T) {
	h := newSubscriptionHandler()

	reqBody := []byte(`{"query":"issue id: DEMO-4","issues":[{"id":"3-3","matches":true}],"subscribe":true,"context":{"$type":"Project","id":"0-0"}}`)
	req := httptest.NewRequest("POST", "/api/issueListSubscription?fields=ticket", bytes.NewBuffer(reqBody))
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	h.SubscribeIssueList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	wantKeys := []string{"$type", "ticket"}
	if got := keysOf(res); !equalKeys(got, wantKeys) {
		t.Errorf("root keys mismatch: got %v want %v", got, wantKeys)
	}
	if res["ticket"] != "g9qk4fe77jefpavefi1qacdhbefa3l" {
		t.Errorf("expected ticket, got %v", res["ticket"])
	}
	if res["$type"] != "IssueListSubscriptionBean" {
		t.Errorf("expected $type IssueListSubscriptionBean, got %v", res["$type"])
	}
	// الحقول غير المطلوبة يجب ألا تظهر
	for _, forbidden := range []string{"query", "subscribe", "context", "issues", "id", "folderId"} {
		if _, exists := res[forbidden]; exists {
			t.Errorf("field %q should NOT appear with fields=ticket", forbidden)
		}
	}
}

// TestIssueListSubscriptionFullFields يتحقق من الجلب الانتقائي باستخدام حقول متعددة.
func TestIssueListSubscriptionFullFields(t *testing.T) {
	h := newSubscriptionHandler()

	reqBody := []byte(`{"query":"issue id: DEMO-4","issues":[{"id":"3-3","matches":true}]}`)
	req := httptest.NewRequest("POST", "/api/issueListSubscription?fields=ticket,query,subscribe,context,issues(id,matches)", bytes.NewBuffer(reqBody))
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	h.SubscribeIssueList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	wantRoot := []string{"$type", "ticket", "query", "subscribe", "context", "issues"}
	if got := keysOf(res); !equalKeys(got, wantRoot) {
		t.Errorf("root keys mismatch: got %v want %v", got, wantRoot)
	}

	contextObj, ok := res["context"].(map[string]any)
	if !ok {
		t.Fatal("expected context object")
	}
	if contextObj["$type"] != "Project" || contextObj["id"] != "0-0" {
		t.Errorf("unexpected context: %v", contextObj)
	}

	issues, ok := res["issues"].([]any)
	if !ok || len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %v", res["issues"])
	}
	item := issues[0].(map[string]any)
	if item["id"] != "3-3" || item["matches"] != true {
		t.Errorf("unexpected issue item: %v", item)
	}
	if item["$type"] != "IssueListSubscriptionItem" {
		t.Errorf("expected issue item $type, got %v", item["$type"])
	}
}

// TestIssueListSubscriptionWithoutBody يتحقق من الطلب بدون جسم (Request Body).
func TestIssueListSubscriptionWithoutBody(t *testing.T) {
	h := newSubscriptionHandler()

	req := httptest.NewRequest("POST", "/api/issueListSubscription?fields=ticket", nil)
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	h.SubscribeIssueList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if res["ticket"] != "g9qk4fe77jefpavefi1qacdhbefa3l" {
		t.Errorf("expected ticket, got %v", res["ticket"])
	}
	if _, exists := res["issues"]; exists {
		t.Errorf("issues should be absent with fields=ticket")
	}
}

// TestIssueListSubscriptionGet يتحقق من دعم GET للتوافقية.
func TestIssueListSubscriptionGet(t *testing.T) {
	h := newSubscriptionHandler()

	req := httptest.NewRequest("GET", "/api/issueListSubscription?fields=ticket", nil)
	req = req.WithContext(withUserID(req.Context(), "313-0"))
	rec := httptest.NewRecorder()

	h.SubscribeIssueList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if res["ticket"] != "g9qk4fe77jefpavefi1qacdhbefa3l" {
		t.Errorf("expected ticket, got %v", res["ticket"])
	}
	if res["$type"] != "IssueListSubscriptionBean" {
		t.Errorf("expected $type IssueListSubscriptionBean, got %v", res["$type"])
	}
}

// TestIssueListSubscriptionUnauthorized يتحقق من 401 عبر الراوتر بدون JWT.
func TestIssueListSubscriptionUnauthorized(t *testing.T) {
	// full := &mockSubscriptionFullStore{subscriptions: &mockSubscriptionStore{sub: sampleSubscription()}}
	router := NewRouter(app.New(), "test-secret")

	req := httptest.NewRequest("POST", "/api/issueListSubscription", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}
}
