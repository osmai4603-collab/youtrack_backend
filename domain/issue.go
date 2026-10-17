package domain

import (
	"errors"
	"time"
)

// Issue-scoped sentinels follow the same rule as the project ones: they are
// declared once here and aliased by the storage and use-case packages, so
// errors.Is keeps matching across the layer boundary.
var (
	// ErrIssueIDRequired is returned when an empty issue ID is used.
	ErrIssueIDRequired = errors.New("issue ID is required")
	// ErrIssueNotVisible is returned when the issue does not exist or its
	// project is outside the caller's scope. The two cases share one sentinel
	// so the endpoints cannot be used to probe which issue IDs are real.
	ErrIssueNotVisible = errors.New("issue is not visible to the user")
)

// Issue states, mirroring the client's IssueStateEnum. The values are validated
// by a CHECK constraint on the column, because the client parses them through
// IssueStateEnum.of(String), which takes a non-nullable String and would throw
// on anything it does not recognise.
const (
	IssueStateToDo       = "to-do"
	IssueStateInProgress = "in-progress"
	IssueStateDone       = "done"
)

// Issue priorities, mirroring IssuePriorityTypeEnum.
const (
	IssuePriorityShowStopper = "show-stopper"
	IssuePriorityCritical    = "critical"
	IssuePriorityMajor       = "major"
	IssuePriorityNormal      = "normal"
	IssuePriorityMinor       = "minor"
)

// Issue types, mirroring IssueTypeEnum.
const (
	IssueTypeBug                = "bug"
	IssueTypeCosmetic           = "cosmetic"
	IssueTypeException          = "exception"
	IssueTypeFeature            = "feature"
	IssueTypeTask               = "task"
	IssueTypeUsabilityProblem   = "usability-problem"
	IssueTypePerformanceProblem = "performance-problem"
	IssueTypeEpic               = "epic"
)

// Issue is a single issue as the read endpoints present it.
//
// The JSON names are snake_case and deliberately not the column names: the
// client's IssueModel.fromJson reads issue_key and issue_sequence where the
// table has id_readable and number_in_project, and its other reads are
// project_id, summary, state, priority, issue_type. Timestamps are RFC3339
// strings rather than the epoch milliseconds the columns hold, because the
// client parses them with DateTime.parse, which rejects a bare integer.
//
// Fields the client renders but the schema cannot source are present as zero
// values on purpose, so the shape stays stable: Tags, Sprints, Links, Build and
// Visibility are emitted as empty rather than omitted, and AssigneeID, ParentID,
// DueDate, Estimation and SpentTime are null when unset. See
// internal/httphandlers/handles/issues for why tags are not serialized yet.
type Issue struct {
	ID          string `json:"id" db:"id"`
	ProjectID   string `json:"project_id" db:"project_id"`
	IssueKey    string `json:"issue_key" db:"id_readable"`
	IssueNumber int    `json:"issue_sequence" db:"number_in_project"`
	Summary     string `json:"summary" db:"summary"`
	Description string `json:"description" db:"description"`

	State     string `json:"state" db:"state"`
	Priority  string `json:"priority" db:"priority"`
	IssueType string `json:"issue_type" db:"issue_type"`

	// AssigneeID is nil when the issue is unassigned. It is a pointer rather
	// than a string so the JSON stays null instead of an empty id, which the
	// client would render as a user that exists. The client's Issue.assigneeId
	// is nullable, so null is safe here.
	AssigneeID *string `json:"assignee_id" db:"assignee_id"`
	// ReporterID is NOT a pointer, unlike AssigneeID. The client's Issue
	// declares reporterId as a non-nullable required String and passes
	// json['reporter_id'] straight into it, so a null here is a runtime type
	// error in Dart, not a graceful degradation. The column allows NULL, so the
	// repository coalesces to "" — the same treatment project.leader_id gets.
	ReporterID string  `json:"reporter_id" db:"reporter_id"`
	ParentID   *string `json:"parent_id" db:"-"`

	SubsystemID *string `json:"subsystem_id" db:"subsystem_id"`
	FixVersions string  `json:"fix_versions" db:"fix_versions"`

	// CreatedAt and UpdatedAt are epoch milliseconds on the row; the handler
	// formats them to RFC3339 on the way out.
	CreatedAt int64  `json:"-" db:"created"`
	UpdatedAt int64  `json:"-" db:"updated"`
	DueDate   *int64 `json:"-" db:"due_date"`

	// Estimation and SpentTime are minutes, matching Duration(minutes: ...).
	Estimation *int `json:"estimation" db:"estimation"`
	SpentTime  *int `json:"spent_time" db:"spent_time"`

	Votes       int  `json:"votes" db:"votes"`
	IsStarred   bool `json:"is_starred" db:"-"`
	IsDraft     bool `json:"is_draft" db:"is_draft"`
	Watchers    int  `json:"watchers_count" db:"-"`
	Attachments int  `json:"attachments_count" db:"-"`
	Comments    int  `json:"comments_count" db:"-"`

	// Tags, Sprints and Links are always emitted as empty arrays. The tables
	// behind them exist, but the client's TagModel parses created_at with
	// DateTime.parse and the tags table has no such column, so populating them
	// from the current schema would throw instead of rendering. Emitting [] is
	// honest about what the server can source and keeps the client's list
	// building.
	Tags    []map[string]any `json:"tags" db:"-"`
	Sprints []map[string]any `json:"sprints" db:"-"`
	Links   []map[string]any `json:"issue_links" db:"-"`

	// Visibility is emitted as a map of group name to member ids. The client
	// renders it through a helper that tolerates a missing key, but sending an
	// explicit empty pair documents that nothing is restricted.
	Visibility map[string][]string `json:"visibility" db:"-"`
}

// IsResolved reports whether the legacy `resolved` column marks the issue as
// finished. It is kept in step with State by the seed and by any future write
// path, so a done issue agrees with both columns.
func (i *Issue) IsResolved() bool {
	return i.State == IssueStateDone
}

// CreatedAtTime and UpdatedAtTime expose the epoch columns as time.Time for the
// handler, which formats them to RFC3339. A zero or negative column yields the
// zero time rather than a bogus 1970 instant.
func (i *Issue) CreatedAtTime() time.Time {
	return epochMillisToTime(i.CreatedAt)
}

func (i *Issue) UpdatedAtTime() time.Time {
	return epochMillisToTime(i.UpdatedAt)
}

func epochMillisToTime(millis int64) time.Time {
	if millis <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(millis).UTC()
}
