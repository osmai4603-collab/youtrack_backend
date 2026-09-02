package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/store"
)

type mockSecuritySearchStore struct {
	fields []*model.SecurityFilterField
}

func (m *mockSecuritySearchStore) GetFilterFields(ctx context.Context, entityType string) ([]*model.SecurityFilterField, error) {
	var res []*model.SecurityFilterField
	for _, f := range m.fields {
		if f.EntityType == entityType {
			res = append(res, f)
		}
	}
	return res, nil
}

type mockGlobalStoreSecurity struct {
	store.Store
	securitySearch *mockSecuritySearchStore
}

func (m *mockGlobalStoreSecurity) SecuritySearch() store.SecuritySearchStore {
	return m.securitySearch
}

// sampleSecurityFilterFields يعيد حقول تصفية مشابهة لـ request29.txt.
func sampleSecurityFilterFields() []*model.SecurityFilterField {
	return []*model.SecurityFilterField{
		{ID: "role.scope", Name: "Scope", EntityType: "ProjectPeopleResponse", Type: "SecurityFilterField"},
		{ID: "role.organization", Name: "Organization", EntityType: "ProjectPeopleResponse", Type: "SecurityFilterField"},
	}
}

func TestSecuritySearch_GetFilterFields(t *testing.T) {
	a := newTestApp(&mockGlobalStoreSecurity{securitySearch: &mockSecuritySearchStore{fields: sampleSecurityFilterFields()}})
	handler := NewSecuritySearchHandler(a)

	t.Run("Get ProjectPeopleResponse fields - default", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/securitySearch/filterFields?entityType=ProjectPeopleResponse", nil)
		rr := httptest.NewRecorder()

		handler.GetFilterFields(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}

		var res []map[string]any
		json.Unmarshal(rr.Body.Bytes(), &res)

		if len(res) != 2 {
			t.Errorf("expected 2 fields, got %d", len(res))
		}

		// Check first field content
		if res[0]["id"] != "role.scope" || res[0]["name"] != "Scope" || res[0]["$type"] != "SecurityFilterField" {
			t.Errorf("unexpected field content: %v", res[0])
		}
	})

	t.Run("Get ProjectPeopleResponse fields - selective fields", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/securitySearch/filterFields?entityType=ProjectPeopleResponse&fields=id", nil)
		rr := httptest.NewRecorder()

		handler.GetFilterFields(rr, req)

		var res []map[string]any
		json.Unmarshal(rr.Body.Bytes(), &res)

		if _, ok := res[0]["name"]; ok {
			t.Error("expected name to be absent")
		}
		if res[0]["id"] != "role.scope" {
			t.Errorf("expected id to be role.scope, got %v", res[0]["id"])
		}
		if res[0]["$type"] != "SecurityFilterField" {
			t.Error("$type should always be present")
		}
	})

	t.Run("Missing entityType", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/securitySearch/filterFields", nil)
		rr := httptest.NewRecorder()

		handler.GetFilterFields(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})
}
