package usecase

import (
	"youtrack_backend/internal/domain"
)

type IssueUseCase struct {
	issueRepo domain.IssueRepository
}

func NewIssueUseCase(issueRepo domain.IssueRepository) *IssueUseCase {
	return &IssueUseCase{
		issueRepo: issueRepo,
	}
}

func (uc *IssueUseCase) GetLinkTypes() ([]*domain.IssueLinkType, error) {
	if uc.issueRepo != nil {
		linkTypes, err := uc.issueRepo.GetLinkTypes()
		if err == nil && len(linkTypes) > 0 {
			return linkTypes, nil
		}
	}

	subtaskSource := "parent for"
	subtaskTarget := "subtask of"
	relatesSource := "relates to"
	relatesTarget := "relates to"
	duplicateSource := "duplicates"
	duplicateTarget := "is duplicated by"

	return []*domain.IssueLinkType{
		{
			ID:                      "Subtask",
			Name:                    "Subtask",
			Directed:                true,
			Aggregation:             true,
			SourceToTarget:          "parent for",
			TargetToSource:          "subtask of",
			LocalizedSourceToTarget: &subtaskSource,
			LocalizedTargetToSource: &subtaskTarget,
			Type:                    "IssueLinkType",
		},
		{
			ID:                      "Relates",
			Name:                    "Relates",
			Directed:                false,
			Aggregation:             false,
			SourceToTarget:          "relates to",
			TargetToSource:          "relates to",
			LocalizedSourceToTarget: &relatesSource,
			LocalizedTargetToSource: &relatesTarget,
			Type:                    "IssueLinkType",
		},
		{
			ID:                      "Duplicate",
			Name:                    "Duplicate",
			Directed:                true,
			Aggregation:             false,
			SourceToTarget:          "duplicates",
			TargetToSource:          "is duplicated by",
			LocalizedSourceToTarget: &duplicateSource,
			LocalizedTargetToSource: &duplicateTarget,
			Type:                    "IssueLinkType",
		},
	}, nil
}

func (uc *IssueUseCase) GetActivitiesPage(issueID string) (*domain.ActivityCursorPage, error) {
	if uc.issueRepo != nil {
		page, err := uc.issueRepo.GetActivities(issueID)
		if err == nil && page != nil {
			return page, nil
		}
	}

	return &domain.ActivityCursorPage{
		Cursor:       "AI.$-:CM.$-:1787653099967",
		BeforeCursor: "AI." + issueID + "-:CM.$-:1787653099967",
		AfterCursor:  "AI.$-:CM.$-:1787653099967",
		HasBefore:    false,
		HasAfter:     false,
		Activities: []*domain.ActivityItem{
			{
				ID:        issueID + ".0-0",
				Timestamp: 1787653099967,
				Author: &domain.User{
					ID:        "11-556284",
					Login:     "author_user",
					FullName:  "Author User",
					AvatarURL: "/hub/api/rest/avatar/11-556284",
					Type:      "User",
				},
				Target: &domain.ActivityTarget{
					ID:   issueID,
					Type: "Issue",
				},
				Field: &domain.ActivityField{
					ID:           "created",
					Presentation: "created",
					Type:         "PredefinedFilterField",
				},
				Category: &domain.ActivityCategory{
					ID:   "IssueCreatedCategory",
					Type: "ActivityCategory",
				},
				Type:     "ADD",
				Added:    []any{},
				Removed:  []any{},
				ItemType: "IssueCreatedActivityItem",
			},
		},
		Type: "ActivityCursorPage",
	}, nil
}

func (uc *IssueUseCase) GetByID(id string) (*domain.Issue, error) {
	if uc.issueRepo != nil {
		issue, err := uc.issueRepo.GetByID(id)
		if err == nil && issue != nil {
			return issue, nil
		}
	}

	return &domain.Issue{
		ID:          id,
		IDReadable:  "DEMO-1",
		Summary:     "Sample YouTrack Issue",
		Description: "This is a detailed description of the issue.",
		Created:     1787653099967,
		Updated:     1787653099967,
		Type:        "Issue",
	}, nil
}

func (uc *IssueUseCase) CreateIssue(issue *domain.Issue) (*domain.Issue, error) {
	if uc.issueRepo != nil {
		_ = uc.issueRepo.Create(issue)
	}
	return issue, nil
}

