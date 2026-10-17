package issuerepository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"youtrack_backend/internal/domain"
)

type issueQueryCaptureDB struct {
	queryArgs    []any
	queryRowArgs []any
}

func (db *issueQueryCaptureDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *issueQueryCaptureDB) Query(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
	db.queryArgs = append([]any(nil), args...)
	return nil, errors.New("captured query")
}

func (db *issueQueryCaptureDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	db.queryRowArgs = append([]any(nil), args...)
	return issueQueryCaptureRow{}
}

type issueQueryCaptureRow struct{}

func (issueQueryCaptureRow) Scan(...any) error {
	return errors.New("captured query")
}

func TestListVisibleToUserBindsAllQueryArguments(t *testing.T) {
	db := &issueQueryCaptureDB{}
	repo := NewPostgresRepository(db)

	_, err := repo.ListVisibleToUser(context.Background(), "user-1", &domain.IssueQuery{})
	if err == nil {
		t.Fatal("expected the captured query error")
	}
	if got, want := len(db.queryArgs), 11; got != want {
		t.Fatalf("query received %d arguments, want %d", got, want)
	}

	permissionRefs, ok := db.queryArgs[1].([]string)
	if !ok {
		t.Fatalf("permission argument has type %T, want []string", db.queryArgs[1])
	}
	for _, permission := range permissionRefs {
		if permission == "READ_ISSUE" {
			return
		}
	}
	t.Fatal("permission references do not include READ_ISSUE")
}

func TestGetVisibleToUserBindsAllQueryArguments(t *testing.T) {
	db := &issueQueryCaptureDB{}
	repo := NewPostgresRepository(db)

	_, err := repo.GetVisibleToUser(context.Background(), "user-1", "issue-1")
	if err == nil {
		t.Fatal("expected the captured query error")
	}
	if got, want := len(db.queryRowArgs), 3; got != want {
		t.Fatalf("query received %d arguments, want %d", got, want)
	}
	if got, want := db.queryRowArgs[2], "issue-1"; got != want {
		t.Fatalf("issue ID argument = %v, want %v", got, want)
	}
}
