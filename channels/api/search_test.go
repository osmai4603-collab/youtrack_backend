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
	a := app.New()
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
