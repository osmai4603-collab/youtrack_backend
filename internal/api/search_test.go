package api

import (
	"bytes"
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

type mockSearchStore struct {
	assist *model.SearchAssistResponse
}

func (m *mockSearchStore) GetAssist(ctx context.Context, query string, caret int, tree *fields.FieldTree) (*model.SearchAssistResponse, error) {
	resp := *m.assist
	if tree != nil {
		if !tree.Has("styleRanges") {
			resp.StyleRanges = nil
		}
		if !tree.Has("suggestions") {
			resp.Suggestions = nil
		}
		if !tree.Has("queryFeatures") {
			resp.QueryFeatures = nil
		}
	}
	return &resp, nil
}

type mockGlobalStore struct {
	store.Store
	search *mockSearchStore
}

func (m *mockGlobalStore) Search() store.SearchStore {
	return m.search
}

func TestSearchAssist(t *testing.T) {
	assist := &model.SearchAssistResponse{
		Query: "project: DEMO",
		Caret: 13,
		StyleRanges: []*model.SearchStyleRange{
			{Length: 7, Start: 0, Style: "field", Type: "SearchStyleRange"},
		},
		Suggestions: []*model.SearchSuggestion{
			{Option: "DEMO", Type: "SearchSuggestion"},
		},
		Type: "SearchAssistResponse",
	}

	mock := &mockGlobalStore{
		search: &mockSearchStore{assist: assist},
	}
	a := app.New(mock)
	handler := NewSearchHandler(a)

	t.Run("Full fields", func(t *testing.T) {
		reqBody, _ := json.Marshal(SearchAssistRequest{Query: "project: DEMO", Caret: 13})
		req := httptest.NewRequest("POST", "/api/search/assist", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		handler.GetAssist(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}

		var res model.SearchAssistResponse
		json.Unmarshal(rr.Body.Bytes(), &res)

		if len(res.StyleRanges) == 0 {
			t.Error("expected styleRanges to be present")
		}
		if len(res.Suggestions) == 0 {
			t.Error("expected suggestions to be present")
		}
	})

	t.Run("Selective fields", func(t *testing.T) {
		reqBody, _ := json.Marshal(SearchAssistRequest{Query: "project: DEMO", Caret: 13})
		req := httptest.NewRequest("POST", "/api/search/assist?fields=query,suggestions(option)", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		handler.GetAssist(rr, req)

		var res map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &res)

		if _, ok := res["styleRanges"]; ok {
			t.Error("expected styleRanges to be absent")
		}
		if _, ok := res["suggestions"]; !ok {
			t.Error("expected suggestions to be present")
		}
	})
}
