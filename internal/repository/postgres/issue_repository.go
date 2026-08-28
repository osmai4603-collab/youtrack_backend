package postgres

import (
	"database/sql"
	"fmt"
	"time"
	"youtrack_backend/internal/domain"
)

type IssueRepository struct {
	db *sql.DB
}

func NewIssueRepository(db *sql.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

func (r *IssueRepository) GetByID(id string) (*domain.Issue, error) {
	query := `
		SELECT 
			i.id, i.id_readable, i.summary, i.description, i.resolved, i.votes, i.created, i.updated,
			p.id, p.name, p.short_name,
			u.id, u.login, u.name, u.full_name, u.avatar_url
		FROM issues i
		LEFT JOIN projects p ON i.project_id = p.id
		LEFT JOIN users u ON i.reporter_id = u.id
		WHERE i.id = $1 OR i.id_readable = $1
	`
	issue := &domain.Issue{}
	var description sql.NullString
	var resolved, created, updated sql.NullInt64
	var pID, pName, pShortName sql.NullString
	var uID, uLogin, uName, uFullName, uAvatarURL sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&issue.ID,
		&issue.IDReadable,
		&issue.Summary,
		&description,
		&resolved,
		&issue.Votes,
		&created,
		&updated,
		&pID,
		&pName,
		&pShortName,
		&uID,
		&uLogin,
		&uName,
		&uFullName,
		&uAvatarURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	if description.Valid {
		issue.Description = description.String
	}
	if resolved.Valid {
		val := resolved.Int64
		issue.Resolved = &val
	}
	if created.Valid {
		issue.Created = created.Int64
	}
	if updated.Valid {
		issue.Updated = updated.Int64
	}

	if pID.Valid {
		issue.Project = &domain.Project{
			ID:        pID.String,
			Name:      pName.String,
			ShortName: pShortName.String,
			Type:      "Project",
		}
	}

	if uID.Valid {
		issue.Reporter = &domain.User{
			ID:        uID.String,
			Login:     uLogin.String,
			Name:      uName.String,
			FullName:  uFullName.String,
			AvatarURL: uAvatarURL.String,
			Type:      "User",
		}
	}

	issue.Type = "Issue"
	return issue, nil
}

func (r *IssueRepository) GetLinkTypes() ([]*domain.IssueLinkType, error) {
	query := `
		SELECT id, name, directed, aggregation, source_to_target, target_to_source
		FROM issue_link_types
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query issue link types: %w", err)
	}
	defer rows.Close()

	var linkTypes []*domain.IssueLinkType
	for rows.Next() {
		lt := &domain.IssueLinkType{}
		var sToT, tToS sql.NullString
		if err := rows.Scan(&lt.ID, &lt.Name, &lt.Directed, &lt.Aggregation, &sToT, &tToS); err != nil {
			return nil, fmt.Errorf("failed to scan issue link type: %w", err)
		}
		if sToT.Valid {
			lt.SourceToTarget = sToT.String
			lt.LocalizedSourceToTarget = &sToT.String
		}
		if tToS.Valid {
			lt.TargetToSource = tToS.String
			lt.LocalizedTargetToSource = &tToS.String
		}
		lt.Type = "IssueLinkType"
		linkTypes = append(linkTypes, lt)
	}

	return linkTypes, nil
}

func (r *IssueRepository) GetActivities(issueID string) (*domain.ActivityCursorPage, error) {
	issue, err := r.GetByID(issueID)
	if err != nil || issue == nil {
		issue = &domain.Issue{
			ID:         issueID,
			IDReadable: issueID,
			Summary:    "Issue " + issueID,
		}
	}

	activityTimestamp := issue.Created
	if activityTimestamp == 0 {
		activityTimestamp = time.Now().UnixMilli()
	}

	author := issue.Reporter
	if author == nil {
		author = &domain.User{
			ID:        "11-556284",
			Login:     "admin",
			FullName:  "System Admin",
			AvatarURL: "/hub/api/rest/avatar/11-556284",
			Type:      "User",
		}
	}

	page := &domain.ActivityCursorPage{
		Cursor:       fmt.Sprintf("AI.%s-:CM.$-:%d", issueID, activityTimestamp),
		BeforeCursor: fmt.Sprintf("AI.%s-:CM.$-:%d", issueID, activityTimestamp),
		AfterCursor:  fmt.Sprintf("AI.%s-:CM.$-:%d", issueID, activityTimestamp),
		HasBefore:    false,
		HasAfter:     false,
		Activities: []*domain.ActivityItem{
			{
				ID:        fmt.Sprintf("%s.0-0", issueID),
				Timestamp: activityTimestamp,
				Author:    author,
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
	}

	return page, nil
}

func (r *IssueRepository) Create(issue *domain.Issue) error {
	query := `
		INSERT INTO issues (id, id_readable, summary, description, project_id, reporter_id, updater_id, created, updated, votes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			summary = EXCLUDED.summary,
			description = EXCLUDED.description,
			updated = EXCLUDED.updated
	`
	var projectID, reporterID, updaterID *string
	if issue.Project != nil {
		projectID = &issue.Project.ID
	}
	if issue.Reporter != nil {
		reporterID = &issue.Reporter.ID
	}
	if issue.Updater != nil {
		updaterID = &issue.Updater.ID
	}

	now := time.Now().UnixMilli()
	if issue.Created == 0 {
		issue.Created = now
	}
	if issue.Updated == 0 {
		issue.Updated = now
	}

	_, err := r.db.Exec(query,
		issue.ID,
		issue.IDReadable,
		issue.Summary,
		issue.Description,
		projectID,
		reporterID,
		updaterID,
		issue.Created,
		issue.Updated,
		issue.Votes,
	)
	if err != nil {
		return fmt.Errorf("failed to insert issue: %w", err)
	}
	return nil
}

