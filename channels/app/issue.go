package app

import (
	"context"
	"errors"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
	"youtrack_backend/channels/store"
)

// ListIssues يعيد قائمة القضايا، مع دعم بحث نصي بسيط.
func (a *YouTrackApp) ListIssues(c request.CTX, query string) ([]*model.Issue, *model.AppError) {
	limit := 50
	ctx := c.Context()
	issues, err := a.Store().Issues().All(ctx, query, limit)
	if err != nil {
		return nil, model.NewInternalError("App.ListIssues", "failed to load issues", err)
	}
	for _, i := range issues {
		a.enrichIssue(ctx, i)
	}
	return issues, nil
}

// GetIssue يعيد قضية واحدة مع تفاصيلها.
func (a *YouTrackApp) GetIssue(c request.CTX, id string) (*model.Issue, *model.AppError) {
	ctx := c.Context()
	var i *model.Issue
	var err error
	i, err = a.Store().Issues().GetByID(ctx, id)
	if err != nil {
		i, err = a.Store().Issues().GetByReadableID(ctx, id)
	}
	if err != nil {
		return nil, model.NewNotFoundError("App.GetIssue", "issue not found")
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
func (a *YouTrackApp) CreateIssue(c request.CTX, req CreateIssueRequest) (*model.Issue, *model.AppError) {
	if req.Summary == "" {
		return nil, model.NewBadRequestError("App.CreateIssue", "summary is required")
	}
	ctx := c.Context()
	project, err := a.Store().Projects().GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, model.NewBadRequestError("App.CreateIssue", "project not found")
	}

	reporterID := req.ReporterID
	if reporterID == "" {
		reporterID = c.UserID()
	}

	i := &model.Issue{
		ID:              newIssueID(),
		IDReadable:      project.ShortName + "-1",
		NumberInProject: 1,
		Summary:         req.Summary,
		Description:     req.Description,
		ProjectID:       req.ProjectID,
		ReporterID:      reporterID,
		CreatorID:       reporterID,
		Created:         nowMillis(),
		Updated:         nowMillis(),
	}
	if err := a.Store().Issues().Create(ctx, i); err != nil {
		return nil, model.NewInternalError("App.CreateIssue", "failed to create issue", err)
	}
	a.enrichIssue(ctx, i)
	return i, nil
}

// GetIssueComments يعيد تعليقات قضية.
func (a *YouTrackApp) GetIssueComments(c request.CTX, issueID string) ([]*model.IssueComment, *model.AppError) {
	ctx := c.Context()
	comments, err := a.Store().Issues().Comments(ctx, issueID)
	if err != nil {
		return nil, model.NewInternalError("App.GetIssueComments", "failed to load comments", err)
	}
	for _, comm := range comments {
		comm.Normalize()
	}
	return comments, nil
}

// GetSortedIssues يعيد قائمة القضايا المرتبة لمجلد أو استعلام بحث.
func (a *YouTrackApp) GetSortedIssues(c request.CTX, folderID string, query string, top int, skip int, tree *fields.FieldTree) (*model.SortedIssuesResponse, *model.AppError) {
	ctx := c.Context()
	items, err := a.Store().Issues().GetSortedIssues(ctx, folderID, query, top, skip)
	if err != nil {
		return nil, model.NewInternalError("App.GetSortedIssues", "failed to fetch sorted issues", err)
	}

	resp := &model.SortedIssuesResponse{
		Tree: items,
	}
	resp.Normalize()
	return resp, nil
}

// GetIssuesGetter يعيد قائمة المشاكل مع حقولها المخصصة (مطابق لـ request28.txt).
func (a *YouTrackApp) GetIssuesGetter(c request.CTX, refs []string, query string, top int, skip int, tree *fields.FieldTree) ([]*model.IssueGetterIssue, *model.AppError) {
	ctx := c.Context()
	issues, err := a.Store().Issues().GetIssuesGetter(ctx, refs, query, top, skip, tree)
	if err != nil {
		return nil, model.NewInternalError("App.GetIssuesGetter", "failed to fetch issues getter", err)
	}
	return issues, nil
}

// GetIssueCount يعيد عدد المشاكل ضمن مجلد معيّن (مطابق لـ request17.txt).
func (a *YouTrackApp) GetIssueCount(c request.CTX, folderID string, query string, unresolvedOnly bool) (*model.IssueCountResponse, *model.AppError) {
	ctx := c.Context()
	resp, err := a.Store().Issues().GetIssueCount(ctx, folderID, query, unresolvedOnly)
	if errors.Is(err, store.ErrFolderNotFound) {
		return nil, model.NewNotFoundError("App.GetIssueCount", "folder not found")
	}
	if err != nil {
		return nil, model.NewInternalError("App.GetIssueCount", "failed to fetch issue count", err)
	}
	resp.Normalize()
	return resp, nil
}

// enrichIssue يملأ مراجع المشروع والمُبلِّغ والتعليقات.
func (a *YouTrackApp) enrichIssue(ctx context.Context, issue *model.Issue) {
	issue.Normalize()
	if p, err := a.Store().Projects().GetByID(ctx, issue.ProjectID); err == nil {
		p.Normalize()
		issue.Project = p
	}
	if issue.ReporterID != "" {
		if u, err := a.Store().Users().GetByID(ctx, issue.ReporterID); err == nil {
			u.Normalize()
			issue.Reporter = u
		}
	}
	if comments, err := a.Store().Issues().Comments(ctx, issue.ID); err == nil {
		issue.Comments = comments
	}
	if tags, err := a.Store().Issues().Tags(ctx, issue.ID); err == nil {
		issue.Tags = tags
	}
}
