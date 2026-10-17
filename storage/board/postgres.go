package board

import (
	"context"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// PostgresRepository implements Repository against the Postgres schema.
type PostgresRepository struct {
	db postgres.DBTX
}

// NewRepository creates a board repository backed by the given database handle.
func NewRepository(db postgres.DBTX) Repository {
	if db == nil {
		panic("board: db must not be nil")
	}
	return &PostgresRepository{db: db}
}

// ---------------------------------------------------------------------------
// GetBoardDetails
// ---------------------------------------------------------------------------

func (r *PostgresRepository) GetBoardDetails(ctx context.Context, projectID string, sprintID string) (*domain.AgileBoardData, error) {
	// 1. Fetch issues for the project (optionally filtered by sprint).
	issues, err := r.fetchIssues(ctx, projectID, sprintID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch subsystems for the project.
	subsystems, err := r.fetchSubsystems(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 3. Fetch sprints that belong to agile boards associated with this project.
	sprints, activeSprint, err := r.fetchSprints(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 4. Assemble swimlanes.
	board := assembleBoard(projectID, issues, subsystems, sprints, activeSprint)
	requestlogger.LogStorageResult(ctx, "board.GetBoardDetails", board)
	return board, nil
}

// ---------------------------------------------------------------------------
// UpdateIssueState
// ---------------------------------------------------------------------------

const updateIssueStateSQL = `
	UPDATE issues
	SET state   = $1,
	    updated = $2
	WHERE id = $3
`

func (r *PostgresRepository) UpdateIssueState(ctx context.Context, issueID string, newState string) error {
	now := time.Now().UnixMilli()
	tag, err := r.db.Exec(ctx, updateIssueStateSQL, newState, now, issueID)
	if err != nil {
		return postgres.TranslateError("board.UpdateIssueState", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIssueNotVisible
	}
	requestlogger.LogStorageResult(ctx, "board.UpdateIssueState", map[string]any{"updated": true})
	return nil
}

// ---------------------------------------------------------------------------
// Internal queries
// ---------------------------------------------------------------------------

// issueRow is a lightweight scan target — only the columns the board card needs.
type issueRow struct {
	ID          string
	IssueKey    string
	Summary     string
	State       string
	Priority    string
	IssueType   string
	SubsystemID *string
}

const fetchIssuesSQL = `
	SELECT i.id,
	       i.id_readable,
	       i.summary,
	       i.state,
	       i.priority,
	       i.issue_type,
	       i.subsystem_id
	FROM issues i
	WHERE i.project_id = $1
	ORDER BY i.updated DESC
`

// fetchIssues loads every issue of the project. Sprint filtering is left as a
// stub for now because the issues table has no sprint_id column; when that
// relationship is modelled (e.g. an issue_sprints join table), an extra
// predicate can be added here without changing the rest of the pipeline.
func (r *PostgresRepository) fetchIssues(ctx context.Context, projectID string, _ string) ([]issueRow, error) {
	rows, err := r.db.Query(ctx, fetchIssuesSQL, projectID)
	if err != nil {
		return nil, postgres.TranslateError("board.fetchIssues", err)
	}
	defer rows.Close()

	var issues []issueRow
	for rows.Next() {
		var row issueRow
		if err := rows.Scan(
			&row.ID,
			&row.IssueKey,
			&row.Summary,
			&row.State,
			&row.Priority,
			&row.IssueType,
			&row.SubsystemID,
		); err != nil {
			return nil, err
		}
		issues = append(issues, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	requestlogger.LogStorageResult(ctx, "board.fetchIssues", issues)
	return issues, nil
}

const fetchSubsystemsSQL = `
	SELECT id, project_id, name, color
	FROM project_subsystems
	WHERE project_id = $1
	ORDER BY name
`

func (r *PostgresRepository) fetchSubsystems(ctx context.Context, projectID string) ([]domain.ProjectSubsystem, error) {
	rows, err := r.db.Query(ctx, fetchSubsystemsSQL, projectID)
	if err != nil {
		return nil, postgres.TranslateError("board.fetchSubsystems", err)
	}
	defer rows.Close()

	var subs []domain.ProjectSubsystem
	for rows.Next() {
		var s domain.ProjectSubsystem
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &s.Color); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	requestlogger.LogStorageResult(ctx, "board.fetchSubsystems", subs)
	return subs, nil
}

const fetchSprintsSQL = `
	SELECT s.id, s.name, COALESCE(s.agile_id, ''), s.start, s.finish,
	       COALESCE(s.goal, ''), COALESCE(s.archived, false), COALESCE(s.is_started, false)
	FROM sprints s
	JOIN agile_board_projects abp ON abp.agile_id = s.agile_id
	WHERE abp.project_id = $1
	ORDER BY s.ordinal
`

func (r *PostgresRepository) fetchSprints(ctx context.Context, projectID string) ([]domain.BoardSprint, *domain.BoardSprint, error) {
	rows, err := r.db.Query(ctx, fetchSprintsSQL, projectID)
	if err != nil {
		return nil, nil, postgres.TranslateError("board.fetchSprints", err)
	}
	defer rows.Close()

	var sprints []domain.BoardSprint
	var active *domain.BoardSprint
	for rows.Next() {
		var sp domain.BoardSprint
		if err := rows.Scan(
			&sp.ID, &sp.Name, &sp.AgileID,
			&sp.Start, &sp.Finish,
			&sp.Goal, &sp.Archived, &sp.IsStarted,
		); err != nil {
			return nil, nil, err
		}
		sprints = append(sprints, sp)
		if sp.IsStarted && !sp.Archived {
			copied := sp
			active = &copied
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	requestlogger.LogStorageResult(ctx, "board.fetchSprints", map[string]any{
		"sprints":       sprints,
		"active_sprint": active,
	})
	return sprints, active, nil
}

// ---------------------------------------------------------------------------
// Assembly logic  (pure, no I/O)
// ---------------------------------------------------------------------------

func assembleBoard(
	projectID string,
	issues []issueRow,
	subsystems []domain.ProjectSubsystem,
	sprints []domain.BoardSprint,
	activeSprint *domain.BoardSprint,
) *domain.AgileBoardData {
	headers := domain.BoardHeaders

	// Map subsystem ID → subsystem for quick lookup.
	subMap := make(map[string]domain.ProjectSubsystem, len(subsystems))
	for _, s := range subsystems {
		subMap[s.ID] = s
	}

	// Bucket issues by (subsystemID, state).
	type laneKey struct {
		subsystemID string
	}
	type bucketKey struct {
		subsystemID string
		state       string
	}
	buckets := make(map[bucketKey][]domain.BoardCard)
	lanesOrdered := make([]string, 0)
	lanesSeen := make(map[string]bool)

	for _, issue := range issues {
		subID := "default"
		if issue.SubsystemID != nil && *issue.SubsystemID != "" {
			subID = *issue.SubsystemID
		}
		if !lanesSeen[subID] {
			lanesSeen[subID] = true
			lanesOrdered = append(lanesOrdered, subID)
		}
		card := domain.BoardCard{
			ID:        issue.ID,
			IssueKey:  issue.IssueKey,
			Summary:   issue.Summary,
			State:     issue.State,
			Priority:  issue.Priority,
			IssueType: issue.IssueType,
		}
		key := bucketKey{subsystemID: subID, state: issue.State}
		buckets[key] = append(buckets[key], card)
	}

	// If there are no subsystems and no issues at all, produce one default lane.
	if len(lanesOrdered) == 0 {
		lanesOrdered = append(lanesOrdered, "default")
	}

	// Column counts across all lanes.
	columnCounts := make(map[string]int, len(headers))
	for _, h := range headers {
		columnCounts[h] = 0
	}

	swimlanes := make([]domain.BoardSwimlane, 0, len(lanesOrdered))
	for _, subID := range lanesOrdered {
		sub := boardSubsystemFor(subID, projectID, subMap)
		columns := make([]domain.BoardColumn, 0, len(headers))
		for _, state := range headers {
			cards := buckets[bucketKey{subsystemID: subID, state: state}]
			if cards == nil {
				cards = []domain.BoardCard{}
			}
			columnCounts[state] += len(cards)
			columns = append(columns, domain.BoardColumn{
				State: state,
				Name:  state,
				Cards: cards,
			})
		}
		swimlanes = append(swimlanes, domain.BoardSwimlane{
			Subsystem: sub,
			Columns:   columns,
		})
	}

	sprintsCopy := sprints
	if sprintsCopy == nil {
		sprintsCopy = []domain.BoardSprint{}
	}

	return &domain.AgileBoardData{
		ProjectID:    projectID,
		Headers:      headers,
		ColumnCounts: columnCounts,
		Swimlanes:    swimlanes,
		Sprints:      sprintsCopy,
		ActiveSprint: activeSprint,
	}
}

func boardSubsystemFor(subID, projectID string, subMap map[string]domain.ProjectSubsystem) domain.BoardSubsystem {
	if sub, ok := subMap[subID]; ok {
		firstLetter := ""
		if len(sub.Name) > 0 {
			firstLetter = string([]rune(sub.Name)[0])
		}
		return domain.BoardSubsystem{
			ID:          sub.ID,
			Name:        sub.Name,
			ProjectID:   sub.ProjectID,
			Color:       sub.Color,
			FirstLetter: firstLetter,
		}
	}
	// Default / unassigned subsystem.
	return domain.BoardSubsystem{
		ID:          "default",
		Name:        "General",
		ProjectID:   projectID,
		Color:       2201331,
		FirstLetter: "G",
	}
}
