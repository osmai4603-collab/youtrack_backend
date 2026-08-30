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

type mockIssueStoreForSorted struct {
	items     []*model.IssueTreeItem
	countResp *model.IssueCountResponse
}

func (m *mockIssueStoreForSorted) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForSorted) GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForSorted) All(ctx context.Context, query string, limit int) ([]*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForSorted) Create(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForSorted) Update(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForSorted) Delete(ctx context.Context, id string) error      { return nil }
func (m *mockIssueStoreForSorted) Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error) {
	return nil, nil
}
func (m *mockIssueStoreForSorted) CreateComment(ctx context.Context, c *model.IssueComment) error {
	return nil
}
func (m *mockIssueStoreForSorted) Tags(ctx context.Context, issueID string) ([]*model.Tag, error) {
	return nil, nil
}
func (m *mockIssueStoreForSorted) Links(ctx context.Context, issueID string) ([]*model.IssueLink, error) {
	return nil, nil
}
func (m *mockIssueStoreForSorted) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error) {
	return m.items, nil
}

func (m *mockIssueStoreForSorted) GetIssueCount(ctx context.Context, folderID string, query string, unresolvedOnly bool) (*model.IssueCountResponse, error) {
	return m.countResp, nil
}
func (m *mockIssueStoreForSorted) GetIssuesGetter(ctx context.Context, refs []string, query string, top int, skip int, tree *fields.FieldTree) ([]*model.IssueGetterIssue, error) {
	return nil, nil
}

type mockIssueStoreForCount struct {
	countResp *model.IssueCountResponse
	err       error
}

func (m *mockIssueStoreForCount) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) All(ctx context.Context, query string, limit int) ([]*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) Create(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForCount) Update(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForCount) Delete(ctx context.Context, id string) error      { return nil }
func (m *mockIssueStoreForCount) Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) CreateComment(ctx context.Context, c *model.IssueComment) error {
	return nil
}
func (m *mockIssueStoreForCount) Tags(ctx context.Context, issueID string) ([]*model.Tag, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) Links(ctx context.Context, issueID string) ([]*model.IssueLink, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error) {
	return nil, nil
}
func (m *mockIssueStoreForCount) GetIssueCount(ctx context.Context, folderID string, query string, unresolvedOnly bool) (*model.IssueCountResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.countResp, nil
}
func (m *mockIssueStoreForCount) GetIssuesGetter(ctx context.Context, refs []string, query string, top int, skip int, tree *fields.FieldTree) ([]*model.IssueGetterIssue, error) {
	return nil, nil
}

type mockIssueStoreForGetter struct {
	issues []*model.IssueGetterIssue
}

func (m *mockIssueStoreForGetter) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) All(ctx context.Context, query string, limit int) ([]*model.Issue, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) Create(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForGetter) Update(ctx context.Context, i *model.Issue) error { return nil }
func (m *mockIssueStoreForGetter) Delete(ctx context.Context, id string) error      { return nil }
func (m *mockIssueStoreForGetter) Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) CreateComment(ctx context.Context, c *model.IssueComment) error {
	return nil
}
func (m *mockIssueStoreForGetter) Tags(ctx context.Context, issueID string) ([]*model.Tag, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) Links(ctx context.Context, issueID string) ([]*model.IssueLink, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) GetIssueCount(ctx context.Context, folderID string, query string, unresolvedOnly bool) (*model.IssueCountResponse, error) {
	return nil, nil
}
func (m *mockIssueStoreForGetter) GetIssuesGetter(ctx context.Context, refs []string, query string, top int, skip int, tree *fields.FieldTree) ([]*model.IssueGetterIssue, error) {
	return m.issues, nil
}

type mockFullStore struct {
	issues store.IssueStore
}

func (m *mockFullStore) Users() store.UserStore                 { return nil }
func (m *mockFullStore) Projects() store.ProjectStore           { return nil }
func (m *mockFullStore) Issues() store.IssueStore               { return m.issues }
func (m *mockFullStore) Admin() store.AdminStore                { return nil }
func (m *mockFullStore) Inbox() store.InboxStore                { return nil }
func (m *mockFullStore) SavedQueries() store.SavedQueryStore    { return nil }
func (m *mockFullStore) Search() store.SearchStore              { return nil }
func (m *mockFullStore) Subscriptions() store.SubscriptionStore { return nil }

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

func TestIssuesGetter(t *testing.T) {
	mockIssues := []*model.IssueGetterIssue{
		{
			ID:         "1-1",
			IDReadable: "DEMO-1",
			Summary:    "Test issue",
			Fields: []*model.IssueCustomField{
				{
					ID:   "pcf-1",
					Type: "SingleEnumIssueCustomField",
					Value: &model.IssueFieldValue{
						ID:   "val-1",
						Name: "High",
						Type: "EnumBundleElement",
					},
					ProjectCustomField: &model.ProjectCustomField{
						ID: "pcf-1",
						Bundle: &model.FieldBundle{
							ID:   "b-1",
							Type: "EnumBundle",
						},
						Field: &model.CustomFieldMetadata{
							ID:   "f-1",
							Name: "Priority",
							FieldType: &model.FieldType{
								ID:        "enum[1]",
								ValueType: "enum",
							},
						},
					},
				},
			},
		},
	}

	issueMock := &mockIssueStoreForGetter{issues: mockIssues}
	a := app.New(&mockFullStore{issues: issueMock})
	handler := NewIssueHandler(a)

	t.Run("Returns correct structure with fields", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/issuesGetter?fields=id,idReadable,summary,fields(id,value(id,name),projectCustomField(id,field(name)))", nil)
		rec := httptest.NewRecorder()

		handler.Getter(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if len(body) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(body))
		}

		issue := body[0]
		if issue["idReadable"] != "DEMO-1" {
			t.Errorf("expected DEMO-1, got %v", issue["idReadable"])
		}

		fieldsList, ok := issue["fields"].([]any)
		if !ok || len(fieldsList) != 1 {
			t.Fatalf("expected 1 field, got %v", issue["fields"])
		}

		field := fieldsList[0].(map[string]any)
		if field["id"] != "pcf-1" {
			t.Errorf("expected pcf-1, got %v", field["id"])
		}

		val, ok := field["value"].(map[string]any)
		if !ok || val["name"] != "High" {
			t.Errorf("expected High, got %v", field["value"])
		}

		pcf, ok := field["projectCustomField"].(map[string]any)
		if !ok || pcf["id"] != "pcf-1" {
			t.Errorf("expected pcf id pcf-1, got %v", field["projectCustomField"])
		}

		fMetadata := pcf["field"].(map[string]any)
		if fMetadata["name"] != "Priority" {
			t.Errorf("expected Priority, got %v", fMetadata["name"])
		}
	})
}

func TestGetIssueCount(t *testing.T) {
	pinned := true
	baseResp := &model.IssueCountResponse{
		Count: 42,
		Folder: &model.IssueFolder{
			ID:               "22-59",
			Name:             "OC",
			ShortName:        "OC",
			Pinned:           &pinned,
			PinnedInHelpdesk: nil,
			Type:             "Project",
		},
		Type: "IssueCountResponse",
	}

	issueMock := &mockIssueStoreForCount{countResp: baseResp}
	a := app.New(&mockFullStore{issues: issueMock})
	handler := NewIssueHandler(a)

	t.Run("fields folder(id),count returns folder $type injected", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/issuesGetter/count?folderId=22-59&fields=folder(id),count", nil)
		rec := httptest.NewRecorder()

		handler.Count(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if body["$type"] != "IssueCountResponse" {
			t.Errorf("expected IssueCountResponse, got %v", body["$type"])
		}
		if body["count"] != float64(42) {
			t.Errorf("expected count 42, got %v", body["count"])
		}
		folder, ok := body["folder"].(map[string]any)
		if !ok {
			t.Fatalf("expected folder object, got %v", body["folder"])
		}
		if folder["id"] != "22-59" {
			t.Errorf("expected folder id 22-59, got %v", folder["id"])
		}
		if folder["$type"] != "Project" {
			t.Errorf("expected folder $type Project, got %v", folder["$type"])
		}
		if _, exists := folder["name"]; exists {
			t.Errorf("name should be filtered out by fields=folder(id),count, got %v", folder["name"])
		}
	})

	t.Run("fields=count only omits folder", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/issuesGetter/count?fields=count", nil)
		rec := httptest.NewRecorder()

		handler.Count(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if _, exists := body["folder"]; exists {
			t.Errorf("folder should be omitted, got %v", body["folder"])
		}
		if body["count"] != float64(42) {
			t.Errorf("expected count 42, got %v", body["count"])
		}
	})

	t.Run("unresolvedOnly=true and folder query pass through", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/issuesGetter/count?folder=OC&query=bug&unresolvedOnly=true&fields=folder(id),count", nil)
		rec := httptest.NewRecorder()

		handler.Count(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		folder, ok := body["folder"].(map[string]any)
		if !ok {
			t.Fatalf("expected folder object, got %v", body["folder"])
		}
		if folder["id"] != "22-59" {
			t.Errorf("expected folder id 22-59, got %v", folder["id"])
		}
	})

	t.Run("folder not found returns 404", func(t *testing.T) {
		issueMock := &mockIssueStoreForCount{err: store.ErrFolderNotFound}
		a := app.New(&mockFullStore{issues: issueMock})
		handler := NewIssueHandler(a)

		req := httptest.NewRequest("GET", "/api/issuesGetter/count?folderId=missing", nil)
		rec := httptest.NewRecorder()

		handler.Count(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}
