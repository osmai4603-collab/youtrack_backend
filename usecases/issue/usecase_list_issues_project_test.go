package issueusecase

import (
	"context"
	"errors"
	"testing"

	"youtrack_backend/internal/domain"
)

// stubIssueReader already records what it was called with. These tests use it to
// pin which reader method a listing goes through, because the route authorizes
// the request in a project scope: if the query and the guard disagree about the
// project, the caller is told about a scope nobody checked.
func TestListIssues_NamedProjectGoesThroughListForProject(t *testing.T) {
	stub := &stubIssueReader{issues: []*domain.Issue{}}

	if _, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{
		UserID: "user-1",
		Query:  &domain.IssueQuery{ProjectID: "proj-1"},
	}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if stub.gotProjectID != "proj-1" {
		t.Fatalf("the listing was filtered on project %q, want proj-1", stub.gotProjectID)
	}
}

func TestListIssues_ResolvesTheProjectAliasBeforeChoosing(t *testing.T) {
	// The client sends ?projectId=; ResolveProject folds it into ProjectID. If
	// the use case branched on the raw field it would miss the filter and list
	// every project, while the guard had authorized only project-1.
	stub := &stubIssueReader{issues: []*domain.Issue{}}

	if _, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{
		UserID: "user-1",
		Query:  &domain.IssueQuery{ProjectIDAlias: "proj-1"},
	}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if stub.gotProjectID != "proj-1" {
		t.Fatalf("the project alias was not resolved; filtered on %q", stub.gotProjectID)
	}
}

func TestListIssues_WithoutAProjectFallsBackToEveryReadableProject(t *testing.T) {
	// The guard refuses a request that names no project, so this branch is only
	// reachable from a caller inside the stack rather than over HTTP. It is kept
	// because the use case is a public entry point, and it must stay safe on its
	// own: ListVisibleToUser filters by the caller's roles.
	stub := &stubIssueReader{issues: []*domain.Issue{}}

	if _, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{
		UserID: "user-1",
	}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if stub.gotProjectID != "" {
		t.Fatalf("expected the cross-project reader, but a project %q was passed", stub.gotProjectID)
	}
}

func TestListIssues_PassesANonNilQueryToTheRepository(t *testing.T) {
	// A nil query used to reach the repository as nil. Resolving the project
	// means this use case now always supplies a value, and the repository must
	// not have to nil-check.
	stub := &stubIssueReader{issues: []*domain.Issue{}}

	if _, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{UserID: "user-1"}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if stub.gotQuery == nil {
		t.Fatal("the repository received a nil query")
	}
}

func TestListIssues_PropagatesTheProjectScopedError(t *testing.T) {
	sentinel := errors.New("project is not visible")
	stub := &stubIssueReader{projErr: sentinel}

	_, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{
		UserID: "user-1",
		Query:  &domain.IssueQuery{ProjectID: "proj-1"},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the repository error to survive wrapping, got %v", err)
	}
}
