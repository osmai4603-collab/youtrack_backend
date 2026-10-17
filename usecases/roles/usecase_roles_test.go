package roles

import (
	"context"
	"errors"
	"reflect"
	"testing"

	permstorage "youtrack_backend/internal/storage/permissions"
)

type roleRepositoryStub struct {
	roles []permstorage.Role
	err   error
}

func (stub *roleRepositoryStub) GetAllRoles(context.Context) ([]permstorage.Role, error) {
	return stub.roles, stub.err
}

func (*roleRepositoryStub) GetRoleByID(context.Context, string) (*permstorage.Role, error) {
	panic("unexpected call")
}

func (*roleRepositoryStub) CreateRole(context.Context, string, *string) (*permstorage.Role, error) {
	panic("unexpected call")
}

func (*roleRepositoryStub) UpdateRole(context.Context, string, *string, *string) (*permstorage.Role, error) {
	panic("unexpected call")
}

func (*roleRepositoryStub) DeleteRole(context.Context, string) error {
	panic("unexpected call")
}

func TestListReturnsAllRoles(t *testing.T) {
	expected := []permstorage.Role{{ID: "role-1", Name: "Administrator"}}
	useCase := New(&roleRepositoryStub{roles: expected})

	actual, err := useCase.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("List() = %#v, want %#v", actual, expected)
	}
}

func TestListReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	useCase := New(&roleRepositoryStub{err: expectedErr})

	if _, err := useCase.List(context.Background()); !errors.Is(err, expectedErr) {
		t.Fatalf("List() error = %v, want %v", err, expectedErr)
	}
}

func TestListRejectsMissingRepository(t *testing.T) {
	if _, err := New(nil).List(context.Background()); err == nil {
		t.Fatal("List() error = nil, want repository configuration error")
	}
}
