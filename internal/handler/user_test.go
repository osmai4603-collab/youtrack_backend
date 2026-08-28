package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/middleware"
	"youtrack_backend/internal/usecase"
	"youtrack_backend/pkg/fields"
)

type mockUserRepo struct{}

func (m *mockUserRepo) GetByID(id string) (*domain.User, error)       { return nil, nil }
func (m *mockUserRepo) GetByLogin(login string) (*domain.User, error) { return nil, nil }
func (m *mockUserRepo) GetCurrentUser() (*domain.User, error)         { return nil, nil }
func (m *mockUserRepo) GetCurrentUserDetail(id string, _ *domain.UserFieldSelect) (*domain.CurrentUser, error) {
	return &domain.CurrentUser{
		ID:                id,
		Login:             "current_user",
		Email:             "user@example.com",
		Type:              "Me",
		Widgets:           []*domain.Widget{},
		FeatureFlags:      []*domain.FeatureFlag{},
		Profiles:          &domain.CurrentUserProfiles{Type: "UserProfiles"},
		IssueRelatedGroup: &domain.UserGroup{ID: "g1", Name: "Group", Type: "UserGroup"},
	}, nil
}
func (m *mockUserRepo) Create(user *domain.User) error { return nil }
func (m *mockUserRepo) Update(user *domain.User) error { return nil }

func TestUserHandler_GetCurrentUser(t *testing.T) {
	userUC := usecase.NewUserUseCase(&mockUserRepo{})
	h := NewUserHandler(userUC)

	req, err := http.NewRequest("GET", "/api/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "11-556284"))

	rr := httptest.NewRecorder()
	h.GetCurrentUser(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var user domain.CurrentUser
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if user.ID != "11-556284" {
		t.Errorf("expected user ID '11-556284', got %s", user.ID)
	}
	if user.Type != "Me" {
		t.Errorf("expected $type 'Me', got %s", user.Type)
	}
	if user.Profiles == nil {
		t.Error("expected profiles object")
	}
}

func TestUserHandler_GetCurrentUser_FieldsFilter(t *testing.T) {
	userUC := usecase.NewUserUseCase(&mockUserRepo{})
	h := NewUserHandler(userUC)

	req, err := http.NewRequest("GET", "/api/users/me?fields=id,login,userType(id,name)", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "11-556284"))

	rr := httptest.NewRecorder()
	h.GetCurrentUser(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var m map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// الحقول المطلوبة موجودة
	if _, exists := m["id"]; !exists {
		t.Error("expected 'id' in filtered result")
	}
	if _, exists := m["login"]; !exists {
		t.Error("expected 'login' in filtered result")
	}
	// $type محفوظ دائماً (سلوك YouTrack الموثّق في request1.txt)
	if _, exists := m["$type"]; !exists {
		t.Error("expected '$type' to always be preserved")
	}
	// الحقول غير المطلوبة مستبعدة
	for _, key := range []string{"email", "profiles", "widgets", "featureFlags"} {
		if _, exists := m[key]; exists {
			t.Errorf("unexpected key %q in filtered result", key)
		}
	}
}

func TestUserHandler_GetCurrentUser_TopZero(t *testing.T) {
	userUC := usecase.NewUserUseCase(&mockUserRepo{})
	h := NewUserHandler(userUC)

	req, err := http.NewRequest("GET", "/api/users/me?$top=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "11-556284"))

	rr := httptest.NewRecorder()
	h.GetCurrentUser(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var m map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty result with $top=0, got %d keys", len(m))
	}
}

func TestBuildUserFieldSelect(t *testing.T) {
	cases := []struct {
		name    string
		fields  string
		wantNil bool
		check   func(*domain.UserFieldSelect) bool
	}{
		{
			name:    "no fields -> nil (load all)",
			fields:  "",
			wantNil: true,
		},
		{
			name:   "only scalar roots -> heavy sections false",
			fields: "id,login,email",
			check: func(s *domain.UserFieldSelect) bool {
				return !s.AnyRootSection()
			},
		},
		{
			name:   "profiles bare leaf -> all sub profiles",
			fields: "id,profiles",
			check: func(s *domain.UserFieldSelect) bool {
				return s.Profiles && s.General && s.Articles && s.Tips && s.Appearance &&
					s.IssuesList && s.Helpdesk && s.AI && s.Notifications && s.TimeTracking
			},
		},
		{
			name:   "profiles(general,appearance) -> only those sub sections",
			fields: "id,profiles(general,appearance)",
			check: func(s *domain.UserFieldSelect) bool {
				return s.Profiles && s.General && s.Appearance &&
					!s.Articles && !s.Tips && !s.TimeTracking && !s.IssuesList &&
					!s.Helpdesk && !s.AI && !s.Notifications
			},
		},
		{
			name:   "widgets & featureFlags isolated",
			fields: "widgets,featureFlags",
			check: func(s *domain.UserFieldSelect) bool {
				return s.Widgets && s.FeatureFlags && !s.Profiles && !s.IssueRelatedGroup
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sel := buildUserFieldSelect(fields.Parse(tc.fields))
			if tc.wantNil {
				if sel != nil {
					t.Fatalf("expected nil selection, got %+v", sel)
				}
				return
			}
			if sel == nil {
				t.Fatal("expected non-nil selection")
			}
			if tc.check != nil && !tc.check(sel) {
				t.Errorf("selection check failed: %+v", sel)
			}
		})
	}
}
