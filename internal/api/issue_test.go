package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
	"youtrack_backend/internal/store"
)

type mockIssueStoreForSorted struct {
	items []*model.IssueTreeItem
}

func (m *mockIssueStoreForSorted) GetByID(ctx context.Context, id string) (*model.Issue, error) { return nil, nil }
func (m *mockIssueStoreForSorted) GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error) { return nil, nil }
func (m *mockIssueStoreForSorted) All(ctx context.Context, query string, limit int) ([]*model.Issue, error) { return nil, nil }
func (m *mockIssueStoreForSorted) Create(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForSorted) Update(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForSorted) Delete(ctx context.Context, id string) error { return nil }
func (m *mockIssueStoreForSorted) Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error) { return nil, nil }
func (m *mockIssueStoreForSorted) CreateComment(ctx context.Context, c *model.IssueComment) error { return nil }
func (m *mockIssueStoreForSorted) Tags(ctx context.Context, issueID string) ([]*model.Tag, error) { return nil, nil }
func (m *mockIssueStoreForSorted) Links(ctx context.Context, issueID string) ([]*model.IssueLink, error) { return nil, nil }
func (m *mockIssueStoreForSorted) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error) {
	return m.items, nil
}

type mockFullStore struct {
	issues store.IssueStore
}

func (m *mockFullStore) Users() store.UserStore       { return nil }
func (m *mockFullStore) Projects() store.ProjectStore { return nil }
func (m *mockFullStore) Issues() store.IssueStore     { return m.issues }
func (m *mockFullStore) Admin() store.AdminStore      { return nil }
func (m *mockFullStore) Inbox() store.InboxStore      { return nil }
func (m *mockFullStore) SavedQueries() store.SavedQueryStore { return nil }
func (m *mockFullStore) Search() store.SearchStore    { return nil }

func TestGetSortedIssues(t *testing.T) {
	mockItems := []*model.IssueTreeItem{
		{
			ID:      "issue-1",
			Matches: true,
			Type:    "IssueTreeItem",
			SearchFeatures: &model.IssueSearchFeatures{
				ID:            "issue-1",
				CommentsCount: 5,
				Type:          "IssueSearchFeatures",
			},
		},
	}

	issueMock := &mockIssueStoreForSorted{items: mockItems}
	fullStore := &mockFullStore{issues: issueMock}
	a := app.New(fullStore)
	handler := NewIssueHandler(a)

	t.Run("Returns correct structure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sortedIssues?folderId=FIN&topRoot=10", nil)
		rec := httptest.NewRecorder()

		handler.GetSortedIssues(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp model.SortedIssuesResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if len(resp.Tree) != 1 {
			t.Errorf("expected 1 item, got %d", len(resp.Tree))
		}

		if resp.Tree[0].ID != "issue-1" {
			t.Errorf("expected issue-1, got %s", resp.Tree[0].ID)
		}

		if resp.Type != "SortedIssuesResponse" {
			t.Errorf("expected SortedIssuesResponse, got %s", resp.Type)
		}
	})
}
