package domain

// CustomFieldBundle تفاصيل حزمة الحقل المخصص
type CustomFieldBundle struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// CustomFieldType نوع الحقل المخصص
type CustomFieldType struct {
	ID           string `json:"id"`
	Presentation string `json:"presentation,omitempty"`
	IsBundleType bool   `json:"isBundleType,omitempty"`
	ValueType    string `json:"valueType,omitempty"`
	IsMultiValue bool   `json:"isMultiValue,omitempty"`
}

// FieldDefinition تعريف الحقل
type FieldDefinition struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Ordinal       int              `json:"ordinal,omitempty"`
	Aliases       string           `json:"aliases,omitempty"`
	LocalizedName string           `json:"localizedName,omitempty"`
	FieldType     *CustomFieldType `json:"fieldType,omitempty"`
}

// ProjectCustomField إعدادات الحقل المخصص على مستوى المشروع
type ProjectCustomField struct {
	ID             string             `json:"id"`
	Field          *FieldDefinition   `json:"field,omitempty"`
	Bundle         *CustomFieldBundle `json:"bundle,omitempty"`
	CanBeEmpty     bool               `json:"canBeEmpty,omitempty"`
	EmptyFieldText string             `json:"emptyFieldText,omitempty"`
	HasRunningJob  bool               `json:"hasRunningJob,omitempty"`
	Ordinal        int                `json:"ordinal,omitempty"`
	IsSpentTime    bool               `json:"isSpentTime,omitempty"`
	IsEstimation   bool               `json:"isEstimation,omitempty"`
	IsPublic       bool               `json:"isPublic,omitempty"`
	Type           string             `json:"$type,omitempty"`
}

// IssueCustomField قيمة الحقل المخصص المرتبط بتذكرة
type IssueCustomField struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name,omitempty"`
	Value              any                 `json:"value,omitempty"`
	ProjectCustomField *ProjectCustomField `json:"projectCustomField,omitempty"`
	HasStateMachine    bool                `json:"hasStateMachine,omitempty"`
	IsUpdatable        bool                `json:"isUpdatable,omitempty"`
	Type               string              `json:"$type,omitempty"`
}

// PullRequest تفاصيل طلب السحب المرتبط بالتذكرة
type PullRequest struct {
	ID         string  `json:"id"`
	IDReadable string  `json:"idReadable,omitempty"`
	IDExternal string  `json:"idExternal,omitempty"`
	Title      string  `json:"title,omitempty"`
	Text       string  `json:"text,omitempty"`
	URL        string  `json:"url,omitempty"`
	Branch     string  `json:"branch,omitempty"`
	Date       int64   `json:"date,omitempty"`
	Author     *User   `json:"author,omitempty"`
	Type       string  `json:"$type,omitempty"`
}

// Article مقال قاعدة المعرفة.
// البنية كاملة وفق الهيكل المرجعي المطلوب في آخر مقال تمت زيارته
// (lastVisitedArticle) في request1.txt:
// id,idReadable,reporter($permittedUsers),summary,project(...),parentArticle(idReadable),
// ordinal,visibility(...),updatersSettings(permittedGroups,permittedUsers),
// hasUnpublishedChanges,isUpdatable,isDeletable,collaborativeDraftId,hasChildren,
// tags(...),hasStar,updated
type Article struct {
	ID                   string                    `json:"id"`
	IDReadable           string                    `json:"idReadable,omitempty"`
	Summary              string                    `json:"summary,omitempty"`
	Reporter             *User                     `json:"reporter,omitempty"`
	Project              *Project                  `json:"project,omitempty"`
	ParentArticle        *ArticleParentRef         `json:"parentArticle,omitempty"`
	Ordinal              int                       `json:"ordinal,omitempty"`
	Visibility           *Visibility               `json:"visibility,omitempty"`
	UpdatersSettings     *ArticleUpdatersSettings  `json:"updatersSettings,omitempty"`
	HasUnpublishedChanges bool                     `json:"hasUnpublishedChanges,omitempty"`
	IsUpdatable          bool                      `json:"isUpdatable,omitempty"`
	IsDeletable          bool                      `json:"isDeletable,omitempty"`
	CollaborativeDraftID string                    `json:"collaborativeDraftId,omitempty"`
	HasChildren          bool                      `json:"hasChildren,omitempty"`
	Tags                 []*Tag                    `json:"tags,omitempty"`
	HasStar              bool                      `json:"hasStar,omitempty"`
	Updated              int64                     `json:"updated,omitempty"`
	Type                 string                    `json:"$type,omitempty"`
}

// ArticleParentRef مرجع مبسط للمقال الأم، ويُطلب حقل idReadable فقط ضمن lastVisitedArticle.
type ArticleParentRef struct {
	IDReadable string `json:"idReadable,omitempty"`
	Type       string `json:"$type,omitempty"`
}

// ArticleUpdatersSettings إعدادات من يمكنه تعديل المقال، ويتضمن المجموعات والمستخدمين المسموحين.
type ArticleUpdatersSettings struct {
	PermittedGroups []*UserGroup `json:"permittedGroups,omitempty"`
	PermittedUsers  []*User      `json:"permittedUsers,omitempty"`
	Type            string       `json:"$type,omitempty"`
}

// IssueLinkType يمثل نوع رابط أو علاقة المهمة في YouTrack (مثل: Subtask, Duplicate, Relates)
type IssueLinkType struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	Directed                bool    `json:"directed"`
	Aggregation             bool    `json:"aggregation"`
	SourceToTarget          string  `json:"sourceToTarget,omitempty"`
	TargetToSource          string  `json:"targetToSource,omitempty"`
	LocalizedSourceToTarget *string `json:"localizedSourceToTarget,omitempty"`
	LocalizedTargetToSource *string `json:"localizedTargetToSource,omitempty"`
	Type                    string  `json:"$type,omitempty"`
}

// IssueLink يمثل الرابط الفعلي بين التذاكر مع المهام المرتبطة
type IssueLink struct {
	ID        string         `json:"id"`
	Direction string         `json:"direction,omitempty"`
	LinkType  *IssueLinkType `json:"linkType,omitempty"`
	Issues    []*Issue       `json:"issues,omitempty"`
	Trimmed   bool           `json:"trimmed,omitempty"`
	Type      string         `json:"$type,omitempty"`
}

// Issue يمثل المهمة أو المشكلة (Task / Issue / Bug)
type Issue struct {
	ID             string              `json:"id"`
	IDReadable     string              `json:"idReadable,omitempty"`
	Summary        string              `json:"summary,omitempty"`
	Description    string              `json:"description,omitempty"`
	Project        *Project            `json:"project,omitempty"`
	Reporter       *User               `json:"reporter,omitempty"`
	Updater        *User               `json:"updater,omitempty"`
	Resolved       *int64              `json:"resolved,omitempty"`
	Created        int64               `json:"created,omitempty"`
	Updated        int64               `json:"updated,omitempty"`
	Tags           []*Tag              `json:"tags,omitempty"`
	Fields         []*IssueCustomField `json:"fields,omitempty"`
	Comments       []*Comment          `json:"comments,omitempty"`
	Attachments    []*Attachment       `json:"attachments,omitempty"`
	Links          []*IssueLink        `json:"links,omitempty"`
	Visibility     *Visibility         `json:"visibility,omitempty"`
	Votes          int                 `json:"votes,omitempty"`
	Type           string              `json:"$type,omitempty"`
}

// Reaction التفاعلات بالرموز التعبيرية
type Reaction struct {
	ID       string `json:"id"`
	Reaction string `json:"reaction"`
	Author   *User  `json:"author,omitempty"`
	Type     string `json:"$type,omitempty"`
}

// Comment يمثل التعليق على التذكرة
type Comment struct {
	ID                   string        `json:"id"`
	Text                 string        `json:"text,omitempty"`
	TextPreview          string        `json:"textPreview,omitempty"`
	Created              int64         `json:"created,omitempty"`
	Updated              int64         `json:"updated,omitempty"`
	Author               *User         `json:"author,omitempty"`
	Visibility           *Visibility   `json:"visibility,omitempty"`
	Attachments          []*Attachment `json:"attachments,omitempty"`
	Reactions            []*Reaction   `json:"reactions,omitempty"`
	CanUndoComment       bool          `json:"canUndoComment,omitempty"`
	CanAddPublicComment  bool          `json:"canAddPublicComment,omitempty"`
	CanUpdateVisibility  bool          `json:"canUpdateVisibility,omitempty"`
	Deleted              bool          `json:"deleted,omitempty"`
	Pinned               bool          `json:"pinned,omitempty"`
	Type                 string        `json:"$type,omitempty"`
}

// SavedQuery يمثل الاستعلامات المحفوظة في النظام
type SavedQuery struct {
	ID                string     `json:"id"`
	Name              string     `json:"name,omitempty"`
	Query             string     `json:"query,omitempty"`
	FolderID          string     `json:"folderId,omitempty"`
	IsUpdatable       bool       `json:"isUpdatable,omitempty"`
	IsDeletable       bool       `json:"isDeletable,omitempty"`
	IsShareable       bool       `json:"isShareable,omitempty"`
	Pinned            bool       `json:"pinned,omitempty"`
	PinnedByDefault   bool       `json:"pinnedByDefault,omitempty"`
	PinnedInHelpdesk  bool       `json:"pinnedInHelpdesk,omitempty"`
	IssuesURL         string     `json:"issuesUrl,omitempty"`
	Owner             *User      `json:"owner,omitempty"`
	SortOrderSortable bool       `json:"sortOrderSortable,omitempty"`
	VisibleFor        *UserGroup `json:"visibleFor,omitempty"`
	UpdateableBy      *UserGroup `json:"updateableBy,omitempty"`
	Type              string     `json:"$type,omitempty"`
}
