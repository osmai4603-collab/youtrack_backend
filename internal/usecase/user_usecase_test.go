package usecase

import (
	"errors"
	"testing"
	"youtrack_backend/internal/domain"
)

type mockUserRepository struct {
	user *domain.User
	err  error
}

func (m *mockUserRepository) GetByID(id string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) GetByLogin(login string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) GetCurrentUser() (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) GetCurrentUserDetail(id string, _ *domain.UserFieldSelect) (*domain.CurrentUser, error) {
	return nil, m.err
}

func (m *mockUserRepository) Create(user *domain.User) error {
	return m.err
}

func (m *mockUserRepository) Update(user *domain.User) error {
	return m.err
}

func TestUserUseCase_GetCurrentUser_DBSuccess(t *testing.T) {
	expectedUser := &domain.User{
		ID:    "11-556284",
		Login: "custom_user",
	}
	repo := &mockUserRepository{user: expectedUser}
	uc := NewUserUseCase(repo)

	user, err := uc.GetCurrentUser("11-556284")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != expectedUser.ID {
		t.Errorf("expected user ID %s, got %s", expectedUser.ID, user.ID)
	}
	if user.Login != expectedUser.Login {
		t.Errorf("expected user Login %s, got %s", expectedUser.Login, user.Login)
	}
}

func TestUserUseCase_GetCurrentUser_Fallback(t *testing.T) {
	repo := &mockUserRepository{err: errors.New("db error")}
	uc := NewUserUseCase(repo)

	user, err := uc.GetCurrentUser("11-556284")
	if err != nil {
		t.Fatalf("expected fallback to work, got error %v", err)
	}

	if user.ID != "11-556284" {
		t.Errorf("expected fallback user ID to match input, got %s", user.ID)
	}
}
