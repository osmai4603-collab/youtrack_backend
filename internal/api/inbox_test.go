package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
	"youtrack_backend/internal/store"
)

type mockInboxStore struct {
	threads []*model.InboxThread
	lastTree *fields.FieldTree
}

func (m *mockInboxStore) Threads(ctx context.Context, userID string, top int, skip int, tree *fields.FieldTree) ([]*model.InboxThread, error) {
	m.lastTree = tree
	return m.threads, nil
}

type mockInboxFullStore struct {
	inbox *mockInboxStore
}

func (m *mockInboxFullStore) Users() store.UserStore       { return nil }
func (m *mockInboxFullStore) Projects() store.ProjectStore { return nil }
func (m *mockInboxFullStore) Issues() store.IssueStore     { return nil }
func (m *mockInboxFullStore) Admin() store.AdminStore      { return nil }
func (m *mockInboxFullStore) Inbox() store.InboxStore      { return m.inbox }

func TestGetInboxThreadsDynamicFields(t *testing.T) {
	mockData := []*model.InboxThread{
		{
			ID:    "thread-1",
			Read:  true,
			Type:  "InboxThread",
			Subject: &model.InboxSubject{
				Text: "Sample subject",
				Type: "InboxSubject",
			},
		},
	}

	mockStore := &mockInboxStore{threads: mockData}
	fullStore := &mockInboxFullStore{inbox: mockStore}
	a := app.New(fullStore)
	handler := NewInboxHandler(a)

	t.Run("Fetch specific fields", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/threads?fields=id,read", nil)
		req = req.WithContext(withUserID(req.Context(), "user-1"))
		rec := httptest.NewRecorder()

		handler.GetThreads(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		if mockStore.lastTree == nil {
			t.Fatal("expected field tree to be passed to store")
		}
		if !mockStore.lastTree.Has("id") || !mockStore.lastTree.Has("read") || mockStore.lastTree.Has("text") {
			t.Errorf("incorrect field tree: %+v", mockStore.lastTree)
		}

		var res []map[string]any
		json.Unmarshal(rec.Body.Bytes(), &res)

		if len(res) == 0 {
			t.Fatal("expected result")
		}

		// Note: The mock returns the full object, but the handler/app logic for Inbox
		// currently just returns what the store gives.
		// The user requested that the STORE should fetch only requested fields.
	})

	t.Run("Default fields when no fields param", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/threads", nil)
		req = req.WithContext(withUserID(req.Context(), "user-1"))
		rec := httptest.NewRecorder()

		handler.GetThreads(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		if mockStore.lastTree != nil && !mockStore.lastTree.IsEmpty() {
			t.Errorf("expected nil or empty field tree, got %+v", mockStore.lastTree)
		}
	})
}
