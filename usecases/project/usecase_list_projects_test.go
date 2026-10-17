package projectusecase

import (
	"context"
	"errors"
	"testing"

	"youtrack_backend/internal/domain"
	storageproject "youtrack_backend/internal/storage/project"
)

func seedRepository() *storageproject.InMemoryRepository {
	repo := storageproject.NewInMemoryRepository()
	repo.Upsert(&domain.Project{ID: "p-1", Name: "Alpha", ShortName: "ALPHA"})
	repo.Upsert(&domain.Project{ID: "p-2", Name: "Beta", ShortName: "BETA"})
	repo.Grant("user-1", "p-1")
	return repo
}

// failingReader proves the use case surfaces repository failures instead of
// swallowing them into an empty result.
type failingReader struct {
	err error
}

func (f *failingReader) ListVisibleToUser(context.Context, string, bool) ([]*domain.Project, error) {
	return nil, f.err
}

func (f *failingReader) ListAll(context.Context, bool) ([]*domain.Project, error) {
	return nil, f.err
}

func (f *failingReader) ListPublic(context.Context, bool) ([]*domain.Project, error) {
	return nil, f.err
}

func TestListProjectsUseCase_ListForUserRepositoryNotConfigured(t *testing.T) {
	uc := NewListProjects(nil, nil)

	if _, err := uc.ListForUser(context.Background(), &ListProjectsParams{UserID: "user-1"}); !errors.Is(err, ErrRepositoryNotConfigured) {
		t.Errorf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}
