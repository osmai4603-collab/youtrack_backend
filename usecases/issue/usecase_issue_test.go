package issueusecase

import (
	"context"
	"errors"
	"testing"

	"youtrack_backend/internal/domain"
)

// stubIssueReader is a hand-written IssueReader. Using a stub rather than the
// in-memory repository keeps these tests on the use case's own logic — argument
// validation, error wrapping, nil handling — instead of re-testing the store.
type stubIssueReader struct {
	issues    []*domain.Issue
	issue     *domain.Issue
	listErr   error
	projErr   error
	getErr    error

	// captured inputs, so a test can assert what the use case passed down
	gotUserID    string
	gotProjectID string
	gotIssueID   string
	gotQuery     *domain.IssueQuery
	listCalls    int
}

func (s *stubIssueReader) ListVisibleToUser(_ context.Context, userID string, query *domain.IssueQuery) ([]*domain.Issue, error) {
	s.listCalls++
	s.gotUserID = userID
	s.gotQuery = query
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.issues, nil
}

func (s *stubIssueReader) ListForProject(_ context.Context, userID, projectID string, query *domain.IssueQuery) ([]*domain.Issue, error) {
	s.listCalls++
	s.gotUserID = userID
	s.gotProjectID = projectID
	s.gotQuery = query
	if s.projErr != nil {
		return nil, s.projErr
	}
	return s.issues, nil
}

func (s *stubIssueReader) GetVisibleToUser(_ context.Context, userID, issueID string) (*domain.Issue, error) {
	s.gotUserID = userID
	s.gotIssueID = issueID
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.issue, nil
}

func TestListIssues_RequiresUserID(t *testing.T) {
	cases := []struct {
		name   string
		params *ListIssuesParams
	}{
		{name: "nil params", params: nil},
		{name: "empty user", params: &ListIssuesParams{}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewListIssues(nil, &stubIssueReader{}).Execute(context.Background(), testCase.params)
			if !errors.Is(err, ErrUserIDRequired) {
				t.Fatalf("expected ErrUserIDRequired, got %v", err)
			}
		})
	}
}

func TestListIssues_PropagatesRepositoryError(t *testing.T) {
	sentinel := errors.New("database is down")
	stub := &stubIssueReader{listErr: sentinel}

	_, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{UserID: "user-1"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the repository error to survive wrapping, got %v", err)
	}
}

func TestListIssues_NilResultBecomesEmptySlice(t *testing.T) {
	stub := &stubIssueReader{issues: nil}

	result, err := NewListIssues(nil, stub).Execute(context.Background(), &ListIssuesParams{UserID: "user-1"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	// nil would marshal to `null`; the handler and then the client would have to
	// special-case a list that is legitimately empty.
	if result.Issues == nil {
		t.Fatal("expected an empty slice, got nil")
	}
}

func TestListIssues_RejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stub := &stubIssueReader{}
	if _, err := NewListIssues(nil, stub).Execute(ctx, &ListIssuesParams{UserID: "user-1"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if stub.listCalls != 0 {
		t.Error("the repository was called despite a cancelled context")
	}
}

func TestListIssues_WithoutRepository(t *testing.T) {
	// A use case built with no repository has to fail loudly rather than
	// silently answering with an empty list.
	if _, err := NewListIssues(nil, nil).Execute(context.Background(), &ListIssuesParams{UserID: "user-1"}); !errors.Is(err, ErrRepositoryNotConfigured) {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestListProjectIssues_PathParameterOverridesQueryProject(t *testing.T) {
	stub := &stubIssueReader{issues: []*domain.Issue{}}

	_, err := NewListProjectIssues(nil, stub).Execute(context.Background(), &ListProjectIssuesParams{
		UserID:    "user-1",
		ProjectID: "p-1",
		Query:     &domain.IssueQuery{ProjectID: "p-2", ProjectIDAlias: "p-3"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Honouring the query instead would let a caller request p-1's endpoint and
	// receive another project's issues, so the path has to win outright.
	if stub.gotQuery.ProjectID != "p-1" {
		t.Errorf("ProjectID = %q, want p-1", stub.gotQuery.ProjectID)
	}
	if stub.gotQuery.ProjectIDAlias != "" {
		t.Errorf("ProjectIDAlias should be cleared, got %q", stub.gotQuery.ProjectIDAlias)
	}
	if stub.gotProjectID != "p-1" {
		t.Errorf("repository got project %q, want p-1", stub.gotProjectID)
	}
}

func TestListProjectIssues_RequiresProjectID(t *testing.T) {
	stub := &stubIssueReader{}

	_, err := NewListProjectIssues(nil, stub).Execute(context.Background(), &ListProjectIssuesParams{UserID: "user-1"})
	if !errors.Is(err, ErrProjectIDRequired) {
		t.Fatalf("expected ErrProjectIDRequired, got %v", err)
	}
}

func TestListProjectIssues_MapsNotVisibleToItsOwnSentinel(t *testing.T) {
	stub := &stubIssueReader{projErr: domain.ErrProjectNotVisible}

	_, err := NewListProjectIssues(nil, stub).Execute(context.Background(), &ListProjectIssuesParams{
		UserID:    "user-1",
		ProjectID: "p-1",
	})
	if !errors.Is(err, ErrProjectNotVisible) {
		t.Fatalf("expected ErrProjectNotVisible, got %v", err)
	}
	// The handler keys on the use-case sentinel, so the domain one has to reach
	// it intact rather than arriving as an opaque wrapped string.
	if !errors.Is(err, domain.ErrProjectNotVisible) {
		t.Error("the domain sentinel should still be matchable through the wrap")
	}
}

func TestListProjectIssues_ReturnsProjectID(t *testing.T) {
	stub := &stubIssueReader{issues: []*domain.Issue{}}

	result, err := NewListProjectIssues(nil, stub).Execute(context.Background(), &ListProjectIssuesParams{
		UserID:    "user-1",
		ProjectID: "p-1",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.ProjectID != "p-1" {
		t.Errorf("ProjectID = %q, want p-1", result.ProjectID)
	}
}

func TestGetIssue_RequiresIDs(t *testing.T) {
	stub := &stubIssueReader{}

	if _, err := NewGetIssue(nil, stub).Execute(context.Background(), nil); !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}
	if _, err := NewGetIssue(nil, stub).Execute(context.Background(), &GetIssueParams{UserID: "user-1"}); !errors.Is(err, ErrIssueIDRequired) {
		t.Fatalf("expected ErrIssueIDRequired, got %v", err)
	}
}

func TestGetIssue_ReturnsIssue(t *testing.T) {
	stub := &stubIssueReader{issue: &domain.Issue{ID: "i-1"}}

	result, err := NewGetIssue(nil, stub).Execute(context.Background(), &GetIssueParams{UserID: "user-1", IssueID: "i-1"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Issue == nil || result.Issue.ID != "i-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetIssue_MapsNotVisibleToItsOwnSentinel(t *testing.T) {
	stub := &stubIssueReader{getErr: domain.ErrIssueNotVisible}

	_, err := NewGetIssue(nil, stub).Execute(context.Background(), &GetIssueParams{UserID: "user-1", IssueID: "i-1"})
	if !errors.Is(err, ErrIssueNotVisible) {
		t.Fatalf("expected ErrIssueNotVisible, got %v", err)
	}
}

// TestGetIssue_NilIssueWithoutError covers a repository that returns (nil, nil).
// Without this the handler would answer 200 with a null body, and the client's
// json['id']! would throw on it.
func TestGetIssue_NilIssueWithoutError(t *testing.T) {
	stub := &stubIssueReader{issue: nil}

	_, err := NewGetIssue(nil, stub).Execute(context.Background(), &GetIssueParams{UserID: "user-1", IssueID: "i-1"})
	if !errors.Is(err, ErrIssueNotVisible) {
		t.Fatalf("expected ErrIssueNotVisible, got %v", err)
	}
}

func TestGetIssue_WithoutRepository(t *testing.T) {
	if _, err := NewGetIssue(nil, nil).Execute(context.Background(), &GetIssueParams{UserID: "user-1", IssueID: "i-1"}); !errors.Is(err, ErrRepositoryNotConfigured) {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}
