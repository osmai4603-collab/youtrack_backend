package sqlstore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/model"
)

// IssueStore تطبيق عمليات القضايا على PostgreSQL.
type IssueStore struct {
	db *pgxpool.Pool
}

func scanIssue(row interface{ Scan(...any) error }) (*model.Issue, error) {
	i := &model.Issue{}
	err := row.Scan(
		&i.ID, &i.IDReadable, &i.NumberInProject, &i.Summary, &i.Description,
		&i.ProjectID, &i.ReporterID, &i.CreatorID, &i.UpdaterID,
		&i.Created, &i.Updated, &i.Resolved, &i.Votes, &i.IsDraft,
	)
	if err != nil {
		return nil, err
	}
	return i, nil
}

const issueSelect = `SELECT id, id_readable, number_in_project, summary, description,
	project_id, reporter_id, creator_id, updater_id, created, updated, resolved, votes, is_draft
	FROM issues`

func (s *IssueStore) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	row := s.db.QueryRow(ctx, issueSelect+` WHERE id = $1`, id)
	return scanIssue(row)
}

func (s *IssueStore) GetByReadableID(ctx context.Context, idReadable string) (*model.Issue, error) {
	row := s.db.QueryRow(ctx, issueSelect+` WHERE id_readable = $1`, idReadable)
	return scanIssue(row)
}

func (s *IssueStore) All(ctx context.Context, query string, limit int) ([]*model.Issue, error) {
	sql := issueSelect
	args := []any{}
	argIdx := 0
	if query != "" {
		argIdx++
		sql += ` WHERE summary ILIKE $1 OR id_readable ILIKE $1`
		args = append(args, "%"+query+"%")
	}
	sql += ` ORDER BY updated DESC`
	if limit > 0 {
		argIdx++
		sql += ` LIMIT ` + fmt.Sprintf("$%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := []*model.Issue{}
	for rows.Next() {
		i, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, i)
	}
	return issues, rows.Err()
}

func (s *IssueStore) Create(ctx context.Context, i *model.Issue) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO issues (id, id_readable, number_in_project, summary, description, project_id, reporter_id, creator_id, updater_id, created, updated, votes, is_draft)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO NOTHING`,
		i.ID, i.IDReadable, i.NumberInProject, i.Summary, i.Description,
		i.ProjectID, i.ReporterID, i.CreatorID, i.UpdaterID,
		i.Created, i.Updated, i.Votes, i.IsDraft)
	return err
}

func (s *IssueStore) Update(ctx context.Context, i *model.Issue) error {
	_, err := s.db.Exec(ctx, `
		UPDATE issues SET summary=$1, description=$2, project_id=$3, updater_id=$4, updated=$5, resolved=$6, votes=$7
		WHERE id=$8`,
		i.Summary, i.Description, i.ProjectID, i.UpdaterID, i.Updated, i.Resolved, i.Votes, i.ID)
	return err
}

func (s *IssueStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM issues WHERE id=$1`, id)
	return err
}

func (s *IssueStore) Comments(ctx context.Context, issueID string) ([]*model.IssueComment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, issue_id, author_id, text, created, updated, is_deleted
		FROM comments WHERE issue_id=$1 ORDER BY created`, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*model.IssueComment{}
	for rows.Next() {
		c := &model.IssueComment{}
		err := rows.Scan(&c.ID, &c.IssueID, &c.AuthorID, &c.Text, &c.Created, &c.Updated, &c.IsDeleted)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (s *IssueStore) CreateComment(ctx context.Context, c *model.IssueComment) error {
	if c.ID == "" {
		c.ID = newID()
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO comments (id, issue_id, author_id, text, created, updated, is_deleted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.IssueID, c.AuthorID, c.Text, c.Created, c.Updated, c.IsDeleted)
	return err
}

func (s *IssueStore) Tags(ctx context.Context, issueID string) ([]*model.Tag, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.name, t.color_id, t.is_deletable, t.is_updatable, t.is_usable
		FROM tags t JOIN issue_tags it ON it.tag_id = t.id
		WHERE it.issue_id = $1`, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []*model.Tag{}
	for rows.Next() {
		t := &model.Tag{}
		if err := rows.Scan(&t.ID, &t.Name, &t.ColorID, &t.IsDeletable, &t.IsUpdatable, &t.IsUsable); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *IssueStore) Links(ctx context.Context, issueID string) ([]*model.IssueLink, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, source_issue_id, target_issue_id, link_type, direction
		FROM issue_links WHERE source_issue_id=$1 OR target_issue_id=$1`, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []*model.IssueLink{}
	for rows.Next() {
		l := &model.IssueLink{}
		if err := rows.Scan(&l.ID, &l.SourceIssueID, &l.TargetIssueID, &l.LinkType, &l.Direction); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *IssueStore) GetSortedIssues(ctx context.Context, folderID string, query string, top int, skip int) ([]*model.IssueTreeItem, error) {
	sql := `
		SELECT i.id,
		       (SELECT COUNT(*) FROM comments c WHERE c.issue_id = i.id) as comments_count,
		       (SELECT COUNT(*) FROM issue_tags it WHERE it.issue_id = i.id) as tags_count,
		       i.votes,
		       i.created,
		       i.updated,
		       i.resolved,
		       LENGTH(COALESCE(i.description, '')) as content_length
		FROM issues i
	`
	args := []any{}
	argIdx := 0

	whereClause := ""
	if folderID != "" {
		argIdx++
		// Assuming folderID could be project_id or short_name
		whereClause += fmt.Sprintf(" WHERE (i.project_id = $%d OR i.project_id IN (SELECT id FROM projects WHERE short_name = $%d))", argIdx, argIdx)
		args = append(args, folderID)
	}

	if query != "" {
		argIdx++
		if whereClause == "" {
			whereClause = " WHERE "
		} else {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("(i.summary ILIKE $%d OR i.description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+query+"%")
	}

	sql += whereClause + " ORDER BY i.updated DESC"

	if top > 0 {
		argIdx++
		sql += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, top)
	}
	if skip > 0 {
		argIdx++
		sql += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, skip)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.IssueTreeItem{}
	for rows.Next() {
		var id string
		var commentsCount, tagsCount, votes, contentLength int
		var created, updated, resolved *int64
		err := rows.Scan(&id, &commentsCount, &tagsCount, &votes, &created, &updated, &resolved, &contentLength)
		if err != nil {
			return nil, err
		}

		item := &model.IssueTreeItem{
			ID:      id,
			Matches: true,
			Ordered: true,
			SearchFeatures: &model.IssueSearchFeatures{
				ID:                  id,
				CommentsCount:       commentsCount,
				TagsCount:           tagsCount,
				VotesCount:          votes,
				CreatedDateMs:       zeroIfNil(created),
				TimeSinceUpdatedMs:  timeSince(updated),
				TimeSinceCreatedMs:  timeSince(created),
				TimeSinceResolvedMs: timeSince(resolved),
				ContentLength:       contentLength,
			},
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func zeroIfNil(val *int64) int64 {
	if val == nil {
		return 0
	}
	return *val
}

func timeSince(val *int64) int64 {
	if val == nil {
		return -1
	}
	now := time.Now().UnixMilli()
	return now - *val
}
