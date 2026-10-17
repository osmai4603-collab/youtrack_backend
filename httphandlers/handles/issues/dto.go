package issueshandler

import (
	"time"
	"youtrack_backend/internal/domain"
)

const issueType = "jetbrains.jetpass.tracker.Issue"

type ProjectIssuesResponse struct {
	ID               string              `json:"id"`
	ProjectID        string              `json:"project_id"`
	IssueKey         string              `json:"issue_key"`
	IssueSequence    int                 `json:"issue_sequence"`
	Summary          string              `json:"summary"`
	Description      string              `json:"description"`
	State            string              `json:"state"`
	Priority         string              `json:"priority"`
	IssueType        string              `json:"issue_type"`
	AssigneeID       *string             `json:"assignee_id"`
	ReporterID       string              `json:"reporter_id"`
	SubsystemID      *string             `json:"subsystem_id"`
	FixVersions      string              `json:"fix_versions"`
	ParentID         *string             `json:"parent_id"`
	BuildID          *string             `json:"build_id"`
	AssigneeAvatar   *string             `json:"assignee_avatar_url"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
	DueDate          *string             `json:"due_date"`
	Estimation       *int                `json:"estimation"`
	SpentTime        *int                `json:"spent_time"`
	Votes            int                 `json:"votes"`
	WatchersCount    int                 `json:"watchers_count"`
	AttachmentsCount int                 `json:"attachments_count"`
	CommentsCount    int                 `json:"comments_count"`
	IsStarred        bool                `json:"is_starred"`
	Tags             []any               `json:"tags"`
	Sprints          []any               `json:"sprints"`
	IssueLinks       []any               `json:"issue_links"`
	Visibility       map[string][]string `json:"visibility"`
	Type             string              `json:"$type"`
}

// newIssueParams maps a domain issue onto the wire shape.
func newIssueParams(issue *domain.Issue) ProjectIssuesResponse {
	if issue == nil {
		return ProjectIssuesResponse{Type: issueType, Tags: []any{}, Sprints: []any{}, IssueLinks: []any{}, Visibility: emptyVisibility()}
	}

	response := ProjectIssuesResponse{
		ID:               issue.ID,
		ProjectID:        issue.ProjectID,
		IssueKey:         issue.IssueKey,
		IssueSequence:    issue.IssueNumber,
		Summary:          issue.Summary,
		Description:      issue.Description,
		State:            issue.State,
		Priority:         issue.Priority,
		IssueType:        issue.IssueType,
		AssigneeID:       issue.AssigneeID,
		ReporterID:       issue.ReporterID,
		SubsystemID:      issue.SubsystemID,
		FixVersions:      issue.FixVersions,
		ParentID:         issue.ParentID,
		CreatedAt:        FormatMillis(issue.CreatedAt),
		UpdatedAt:        FormatMillis(issue.UpdatedAt),
		DueDate:          FormatOptionalMillis(issue.DueDate),
		Estimation:       issue.Estimation,
		SpentTime:        issue.SpentTime,
		Votes:            issue.Votes,
		WatchersCount:    issue.Watchers,
		AttachmentsCount: issue.Attachments,
		CommentsCount:    issue.Comments,
		IsStarred:        issue.IsStarred,
		Tags:             []any{},
		Sprints:          []any{},
		IssueLinks:       []any{},
		Visibility:       issue.Visibility,
		Type:             issueType,
	}
	if response.Visibility == nil {
		response.Visibility = emptyVisibility()
	}

	return response
}

func FormatMillis(millis int64) string {
	if millis <= 0 {
		return ""
	}
	return time.UnixMilli(millis).UTC().Format(time.RFC3339)
}

func FormatOptionalMillis(millis *int64) *string {
	if millis == nil || *millis <= 0 {
		return nil
	}
	formatted := FormatMillis(*millis)
	if formatted == "" {
		return nil
	}
	return &formatted
}

func emptyVisibility() map[string][]string {
	return map[string][]string{"users": {}, "groups": {}}
}
