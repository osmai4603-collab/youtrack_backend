package model

// SavedQuery يمثّل استعلامًا محفوظًا في النظام (Request #15).
type SavedQuery struct {
	ID                    string                     `json:"id" db:"id"`
	IssuesURL             string                     `json:"issuesUrl,omitempty" db:"issues_url"`
	Name                  string                     `json:"name,omitempty" db:"name"`
	Query                 string                     `json:"query,omitempty" db:"query"`
	PinnedByDefault       bool                       `json:"pinnedByDefault" db:"pinned_by_default"`
	Pinned                bool                       `json:"pinned" db:"pinned"`
	PinnedInHelpdesk      bool                       `json:"pinnedInHelpdesk" db:"pinned_in_helpdesk"`
	IsUpdatable           bool                       `json:"isUpdatable" db:"is_updatable"`
	IsDeletable           bool                       `json:"isDeletable" db:"is_deletable"`
	IsShareable           bool                       `json:"isShareable" db:"is_shareable"`
	OwnerID               string                     `json:"-" db:"owner_id"`
	Owner                 *User                      `json:"owner,omitempty"`
	VisibleForID          string                     `json:"-" db:"visible_for_id"`
	ReadSharingSettings   *SavedQuerySharingSettings `json:"readSharingSettings,omitempty"`
	UpdateableByID        string                     `json:"-" db:"updateable_by_id"`
	UpdateSharingSettings *SavedQuerySharingSettings `json:"updateSharingSettings,omitempty"`
	SortOrderSortable     bool                       `json:"-" db:"sort_order_sortable"`
	SortOrder             *SavedQueryIssueOrder      `json:"sortOrder,omitempty"`
	Type                  string                     `json:"$type,omitempty"`
}

// SavedQuerySharingSettings يمثّل إعدادات المشاركة لاستعلام محفوظ
// (نظام WatchFolderSharingSettings في YouTrack).
type SavedQuerySharingSettings struct {
	PermittedGroups []*SavedQueryGroup `json:"permittedGroups"`
	PermittedUsers  []*User            `json:"permittedUsers"`
	Type            string             `json:"$type"`
}

// SavedQueryIssueOrder يمثّل ترتيب إصدارات الاستعلام المحفوظ.
type SavedQueryIssueOrder struct {
	IsSortable bool   `json:"isSortable"`
	Type       string `json:"$type"`
}

// SavedQueryGroup يمثّل مجموعة مستخدمين مشاركة في استعلام محفوظ
// مع الفريق المرتبط بها (يضمّن UserGroup دون تعديل مخططها القديم).
type SavedQueryGroup struct {
	UserGroup
	TeamForProjectID string      `json:"-" db:"team_for_project_id"`
	TeamForProject   *ProjectRef `json:"teamForProject,omitempty"`
}
