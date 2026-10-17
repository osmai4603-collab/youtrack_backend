package domain

// AgileBoardData is the full response payload for a board view. It groups the
// project's issues into swimlanes (one per subsystem, plus a default lane) and
// columns (the three issue states). The structure mirrors the JSON shape the
// frontend's AgileBoardViewPage expects.
type AgileBoardData struct {
	ProjectID    string            `json:"projectId"`
	Headers      []string          `json:"headers"`
	ColumnCounts map[string]int    `json:"columnCounts"`
	Swimlanes    []BoardSwimlane   `json:"swimlanes"`
	Sprints      []BoardSprint     `json:"sprints"`
	ActiveSprint *BoardSprint      `json:"activeSprint"`
}

// BoardSwimlane is one horizontal row of the board, grouping cards by subsystem.
type BoardSwimlane struct {
	Subsystem BoardSubsystem `json:"subsystem"`
	Columns   []BoardColumn  `json:"columns"`
}

// BoardSubsystem identifies the subsystem a swimlane represents.
type BoardSubsystem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   string `json:"projectId"`
	Color       int64  `json:"color"`
	FirstLetter string `json:"firstLetter"`
}

// BoardColumn is one vertical column inside a swimlane.
type BoardColumn struct {
	State string      `json:"state"`
	Name  string      `json:"name"`
	Cards []BoardCard `json:"cards"`
}

// BoardCard is the minimal issue representation rendered on the board.
type BoardCard struct {
	ID        string `json:"id"`
	IssueKey  string `json:"issueKey"`
	Summary   string `json:"summary"`
	State     string `json:"state"`
	Priority  string `json:"priority"`
	IssueType string `json:"issueType"`
}

// BoardSprint is a sprint entry returned alongside the board data.
type BoardSprint struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AgileID   string `json:"agileId"`
	Start     *int64 `json:"start"`
	Finish    *int64 `json:"finish"`
	Goal      string `json:"goal"`
	Archived  bool   `json:"archived"`
	IsStarted bool   `json:"isStarted"`
}

// BoardHeaders are the three fixed columns the frontend displays.
var BoardHeaders = []string{IssueStateToDo, IssueStateInProgress, IssueStateDone}

// ValidIssueState reports whether s is one of the three accepted issue states.
func ValidIssueState(s string) bool {
	switch s {
	case IssueStateToDo, IssueStateInProgress, IssueStateDone:
		return true
	}
	return false
}
