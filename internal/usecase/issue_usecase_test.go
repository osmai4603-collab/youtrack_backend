package usecase

import (
	"errors"
	"testing"
	"youtrack_backend/internal/domain"
)

type mockIssueRepository struct {
	linkTypes  []*domain.IssueLinkType
	activities *domain.ActivityCursorPage
	err        error
}

func (m *mockIssueRepository) GetByID(id string) (*domain.Issue, error) {
	return nil, m.err
}

func (m *mockIssueRepository) GetLinkTypes() ([]*domain.IssueLinkType, error) {
	return m.linkTypes, m.err
}

func (m *mockIssueRepository) GetActivities(issueID string) (*domain.ActivityCursorPage, error) {
	return m.activities, m.err
}

func (m *mockIssueRepository) Create(issue *domain.Issue) error {
	return m.err
}

func TestIssueUseCase_GetLinkTypes_DBSuccess(t *testing.T) {
	expectedLinks := []*domain.IssueLinkType{
		{ID: "CustomLink", Name: "CustomLink"},
	}
	repo := &mockIssueRepository{linkTypes: expectedLinks}
	uc := NewIssueUseCase(repo)

	links, err := uc.GetLinkTypes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(links) != 1 || links[0].ID != "CustomLink" {
		t.Errorf("expected 1 link type with ID 'CustomLink', got %+v", links)
	}
}

func TestIssueUseCase_GetLinkTypes_Fallback(t *testing.T) {
	repo := &mockIssueRepository{err: errors.New("db error")}
	uc := NewIssueUseCase(repo)

	links, err := uc.GetLinkTypes()
	if err != nil {
		t.Fatalf("expected fallback to work, got %v", err)
	}

	if len(links) != 3 {
		t.Errorf("expected 3 default link types, got %d", len(links))
	}
}
