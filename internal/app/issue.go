package app

import (
	"context"
	"errors"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
	"youtrack_backend/internal/store"
)

// ListIssues يعيد قائمة القضايا، مع دعم بحث نصي بسيط.
func (a *App) ListIssues(ctx context.Context, query string) ([]*model.Issue, error) {
	limit := 50
	issues, err := a.store.Issues().All(ctx, query, limit)
	if err != nil {
		return nil, model.Internal("failed to load issues: %v", err)
	}
	for _, i := range issues {
		a.enrichIssue(ctx, i)
	}
	return issues, nil
}

// GetIssue يعيد قضية واحدة مع تفاصيلها.
func (a *App) GetIssue(ctx context.Context, id string) (*model.Issue, error) {
	var i *model.Issue
	var err error
	i, err = a.store.Issues().GetByID(ctx, id)
	if err != nil {
		i, err = a.store.Issues().GetByReadableID(ctx, id)
	}
	if err != nil {
		return nil, model.NotFound("issue %q not found", id)
	}
	a.enrichIssue(ctx, i)
	return i, nil
}

// CreateIssueRequest يحمل بيانات إنشاء قضية.
type CreateIssueRequest struct {
	Summary     string
	Description string
	ProjectID   string
	ReporterID  string
}

// CreateIssue ينشئ قضية جديدة.
func (a *App) CreateIssue(ctx context.Context, req CreateIssueRequest) (*model.Issue, error) {
	if req.Summary == "" {
		return nil, model.BadRequest("summary is required")
	}
	project, err := a.store.Projects().GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, model.BadRequest("project %q not found", req.ProjectID)
	}

	i := &model.Issue{
		ID:              newIssueID(),
		IDReadable:      project.ShortName + "-1",
		NumberInProject: 1,
		Summary:         req.Summary,
		Description:     req.Description,
		ProjectID:       req.ProjectID,
		ReporterID:      req.ReporterID,
		CreatorID:       req.ReporterID,
		Created:         nowMillis(),
		Updated:         nowMillis(),
	}
	if err := a.store.Issues().Create(ctx, i); err != nil {
		return nil, model.Internal("failed to create issue")
	}
	a.enrichIssue(ctx, i)
	return i, nil
}

// GetIssueComments يعيد تعليقات قضية.
func (a *App) GetIssueComments(ctx context.Context, issueID string) ([]*model.IssueComment, error) {
	comments, err := a.store.Issues().Comments(ctx, issueID)
	if err != nil {
		return nil, model.Internal("failed to load comments")
	}
	for _, c := range comments {
		c.Normalize()
	}
	return comments, nil
}

// GetSortedIssues returns a list of sorted issues for a folder or search query.
func (a *App) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int, tree *fields.FieldTree) (*model.SortedIssuesResponse, error) {
	items, err := a.store.Issues().GetSortedIssues(ctx, folderID, query, top, skip)
	if err != nil {
		return nil, model.Internal("failed to fetch sorted issues: %v", err)
	}

	resp := &model.SortedIssuesResponse{
		Tree: items,
	}
	resp.Normalize()
	return resp, nil
}

// GetIssuesGetter يعيد قائمة المشاكل مع حقولها المخصصة (مطابق لـ request28.txt).
func (a *App) GetIssuesGetter(ctx context.Context, refs []string, query string, top int, skip int, tree *fields.FieldTree) ([]*model.IssueGetterIssue, error) {
	issues, err := a.store.Issues().GetIssuesGetter(ctx, refs, query, top, skip, tree)
	if err != nil {
		return nil, model.Internal("failed to fetch issues getter: %v", err)
	}
	return issues, nil
}

// GetIssueCount يعيد عدد المشاكل ضمن مجلد معيّن (مطابق لـ request17.txt).
func (a *App) GetIssueCount(ctx context.Context, folderID string, query string, unresolvedOnly bool) (*model.IssueCountResponse, error) {
	resp, err := a.store.Issues().GetIssueCount(ctx, folderID, query, unresolvedOnly)
	if errors.Is(err, store.ErrFolderNotFound) {
		return nil, model.NotFound("folder %q not found", folderID)
	}
	if err != nil {
		return nil, model.Internal("failed to fetch issue count: %v", err)
	}
	resp.Normalize()
	return resp, nil
}

// enrichIssue يملأ مراجع المشروع والمُبلِّغ والتعليقات.
func (a *App) enrichIssue(ctx context.Context, issue *model.Issue) {
	issue.Normalize()
	if p, err := a.store.Projects().GetByID(ctx, issue.ProjectID); err == nil {
		p.Normalize()
		issue.Project = p
	}
	if issue.ReporterID != "" {
		if u, err := a.store.Users().GetByID(ctx, issue.ReporterID); err == nil {
			u.Normalize()
			issue.Reporter = u
		}
	}
	if comments, err := a.store.Issues().Comments(ctx, issue.ID); err == nil {
		issue.Comments = comments
	}
	if tags, err := a.store.Issues().Tags(ctx, issue.ID); err == nil {
		issue.Tags = tags
	}
}
