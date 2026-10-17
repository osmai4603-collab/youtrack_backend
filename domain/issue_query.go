package domain

// IssueSortField is the set of orderings the issues list may request.
//
// It is a closed enum rather than a free string so the repository can map it to
// a fixed SQL fragment. Interpolating a caller-supplied column name would make
// the sort parameter injectable, and no amount of validating the other query
// parameters would cover that.
type IssueSortField string

const (
	// IssueSortUpdated orders by the last update, the client's default.
	IssueSortUpdated IssueSortField = "updated"
	// IssueSortCreated orders by creation time.
	IssueSortCreated IssueSortField = "created"
	// IssueSortPriority orders by severity, most severe first regardless of
	// the ascending flag, because "ascending by priority" has no useful meaning.
	IssueSortPriority IssueSortField = "priority"
	// IssueSortVotes orders by the vote count.
	IssueSortVotes IssueSortField = "votes"
	// IssueSortSummary orders alphabetically by summary.
	IssueSortSummary IssueSortField = "summary"
)

// IssueQuery is the filter and paging state of an issues listing.
//
// Every field is optional: the zero value asks for the identity's readable
// issues with the server's default ordering, which is what GET /api/issues
// sends with no query string at all.
type IssueQuery struct {
	// ProjectID narrows the listing to one project. GET /api/issues/{issueID}
	// and the per-issue read do not use it; GET /api/projects/{projectID}/issues
	// sets it and additionally requires the project to be visible.
	ProjectID string

	// ProjectIDAlias is the same filter arriving under YouTrack's camelCase
	// `projectId` spelling, which is the name the client's own
	// ProjectIdFilter.toJson() emits. It is resolved into ProjectID by
	// ResolveProject before the query reaches storage, so repositories only
	// ever deal with ProjectID and cannot disagree about precedence.
	ProjectIDAlias string

	// SearchQuery matches the issue key and the summary, case-insensitively.
	// The two are matched with OR because users type a key when they mean one
	// and words when they mean the other.
	SearchQuery string

	State     string
	Priority  string
	IssueType string

	// AssigneeID filters on the assignee. The sentinel UnassignedAssignee
	// selects issues with no assignee, which is otherwise unreachable because
	// an absent filter and an empty assignee are the same JSON value.
	AssigneeID string

	// TagID keeps only issues carrying this tag.
	TagID string

	SortField IssueSortField
	Ascending bool

	// Limit and Offset page the result. A non-positive Limit means "the server
	// default"; Limit is capped by the repository so a caller cannot ask for an
	// unbounded response.
	Limit  int
	Offset int
}

// UnassignedAssignee is the AssigneeID value that selects unassigned issues.
//
// It is a distinct string rather than "" so the two cases cannot be confused
// in the query: "" means no assignee filter at all, UnassignedAssignee means
// assignee_id IS NULL.
const UnassignedAssignee = "__unassigned__"

// ResolveProject folds ProjectIDAlias into ProjectID so the rest of the stack
// has one field to read. ProjectID wins when both are set, because it comes
// from the path parameter of the per-project route, which is the more specific
// statement of intent.
func (q *IssueQuery) ResolveProject() {
	if q == nil {
		return
	}
	if q.ProjectID == "" {
		q.ProjectID = q.ProjectIDAlias
	}
	q.ProjectIDAlias = ""
}
