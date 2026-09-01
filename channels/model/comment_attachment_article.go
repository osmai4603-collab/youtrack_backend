package model

// Comment يمثّل تعليقًا عاماً ضمن خيط/نشاط.
type Comment struct {
	ID           string `json:"id" db:"id"`
	IssueID      string `json:"issueId,omitempty" db:"issue_id"`
	AuthorID     string `json:"-" db:"author_id"`
	Author       *User  `json:"author,omitempty"`
	Text         string `json:"text" db:"text"`
	Created      int64  `json:"created,omitempty" db:"created"`
	Updated      int64  `json:"updated,omitempty" db:"updated"`
	URL          string `json:"url,omitempty" db:"url"`
	IsDeleted    bool   `json:"deleted,omitempty" db:"is_deleted"`
	ParentID     string `json:"parentId,omitempty" db:"parent_comment_id"`
	VisibilityID string `json:"-" db:"visibility_type"`
	Type         string `json:"$type,omitempty"`
}

// Attachment يمثّل مرفقاً ضمن قضية أو تعليق أو مقال.
type Attachment struct {
	ID           string `json:"id" db:"id"`
	IssueID      string `json:"issueId,omitempty" db:"issue_id"`
	CommentID    string `json:"commentId,omitempty" db:"comment_id"`
	ArticleID    string `json:"articleId,omitempty" db:"article_id"`
	AuthorID     string `json:"-" db:"author_id"`
	Author       *User  `json:"author,omitempty"`
	Name         string `json:"name" db:"name"`
	Size         int64  `json:"size,omitempty" db:"size"`
	MimeType     string `json:"mimeType,omitempty" db:"mime_type"`
	URL          string `json:"url,omitempty" db:"url"`
	Created      int64  `json:"created,omitempty" db:"created"`
	VisibilityID string `json:"-" db:"visibility_type"`
	Type         string `json:"$type,omitempty"`
}

// Article يمثّل مقالاً في قاعدة المعرفة.
type Article struct {
	ID                    string `json:"id" db:"id"`
	IDReadable            string `json:"idReadable,omitempty" db:"id_readable"`
	Summary               string `json:"summary,omitempty" db:"summary"`
	ProjectID             string `json:"-" db:"project_id"`
	ReporterID            string `json:"-" db:"reporter_id"`
	Reporter              *User  `json:"reporter,omitempty"`
	ParentArticleID       string `json:"parentArticleId,omitempty" db:"parent_article_id"`
	Ordinal               int    `json:"ordinal,omitempty" db:"ordinal"`
	HasUnpublishedChanges bool   `json:"hasUnpublishedChanges,omitempty" db:"has_unpublished_changes"`
	HasChildren           bool   `json:"hasChildren,omitempty" db:"has_children"`
	HasStar               bool   `json:"hasStar,omitempty" db:"has_star"`
	IsUpdatable           bool   `json:"isUpdatable,omitempty" db:"is_updatable"`
	IsDeletable           bool   `json:"isDeletable,omitempty" db:"is_deletable"`
	CollaborativeDraftID  string `json:"collaborativeDraftId,omitempty" db:"collaborative_draft_id"`
	Updated               int64  `json:"updated,omitempty" db:"updated"`
	Type                  string `json:"$type,omitempty"`
}
