package model

// Issue يصف قضية (تذكرة عمل) في النظام.
type Issue struct {
	ID              string          `json:"id" db:"id"`
	IDReadable      string          `json:"idReadable" db:"id_readable"`
	NumberInProject int             `json:"numberInProject,omitempty" db:"number_in_project"`
	Summary         string          `json:"summary" db:"summary"`
	Description     string          `json:"description,omitempty" db:"description"`
	ProjectID       string          `json:"-" db:"project_id"`
	Project         *Project        `json:"project,omitempty"`
	ReporterID      string          `json:"-" db:"reporter_id"`
	Reporter        *User           `json:"reporter,omitempty"`
	CreatorID       string          `json:"-" db:"creator_id"`
	Creator         *User           `json:"creator,omitempty"`
	UpdaterID       string          `json:"-" db:"updater_id"`
	Created         int64           `json:"created,omitempty" db:"created"`
	Updated         int64           `json:"updated,omitempty" db:"updated"`
	Resolved        int64           `json:"resolved,omitempty" db:"resolved"`
	Votes           int             `json:"votes,omitempty" db:"votes"`
	IsDraft         bool            `json:"isDraft,omitempty" db:"is_draft"`
	Comments        []*IssueComment `json:"comments,omitempty"`
	Tags            []*Tag          `json:"tags,omitempty"`
	Type            string          `json:"$type,omitempty"`
}

func (i *Issue) Normalize() {
	i.Type = "Issue"
	if i.Project != nil {
		i.Project.Normalize()
	}
	if i.Reporter != nil {
		i.Reporter.Normalize()
	}
	if i.Creator != nil {
		i.Creator.Normalize()
	}
}

// IssueComment يصف تعليقًا على قضية.
type IssueComment struct {
	ID        string `json:"id" db:"id"`
	IssueID   string `json:"issueId" db:"issue_id"`
	AuthorID  string `json:"-" db:"author_id"`
	Author    *User  `json:"author,omitempty"`
	Text      string `json:"text" db:"text"`
	Created   int64  `json:"created,omitempty" db:"created"`
	Updated   int64  `json:"updated,omitempty" db:"updated"`
	IsDeleted bool   `json:"deleted,omitempty" db:"is_deleted"`
	Type      string `json:"$type,omitempty"`
}

func (c *IssueComment) Normalize() {
	c.Type = "IssueComment"
	if c.Author != nil {
		c.Author.Normalize()
	}
}

// IssueLink يصف رابطًا بين قضيتين.
type IssueLink struct {
	ID            string `json:"id" db:"id"`
	SourceIssueID string `json:"-" db:"source_issue_id"`
	TargetIssueID string `json:"-" db:"target_issue_id"`
	LinkType      string `json:"linkType" db:"link_type"`
	Direction     string `json:"direction,omitempty" db:"direction"`
	Type          string `json:"$type,omitempty"`
}

// Tag يصف وسمًا.
type Tag struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	ColorID     string `json:"-" db:"color_id"`
	IsDeletable bool   `json:"isDeletable,omitempty" db:"is_deletable"`
	IsUpdatable bool   `json:"isUpdatable,omitempty" db:"is_updatable"`
	IsUsable    bool   `json:"isUsable,omitempty" db:"is_usable"`
}
