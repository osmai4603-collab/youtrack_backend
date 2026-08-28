package domain

// ActivityCategory تصنيف النشاط (مثل IssueCreatedCategory أو CommentsCategory)
type ActivityCategory struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// ActivityField الحقل المرتبط بالنشاط
type ActivityField struct {
	ID           string           `json:"id"`
	Presentation string           `json:"presentation,omitempty"`
	CustomField  *CustomFieldType `json:"customField,omitempty"`
	Type         string           `json:"$type,omitempty"`
}

// ActivityTarget الهدف المرتبط بالنشاط
type ActivityTarget struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// ActivityItem يمثل الحدث الواحد في سجل الأنشطة
type ActivityItem struct {
	ID           string            `json:"id"`
	Timestamp    int64             `json:"timestamp"`
	Author       *User             `json:"author,omitempty"`
	Category     *ActivityCategory `json:"category,omitempty"`
	Field        *ActivityField    `json:"field,omitempty"`
	Target       *ActivityTarget   `json:"target,omitempty"`
	TargetMember *string           `json:"targetMember,omitempty"`
	Type         string            `json:"type,omitempty"` // e.g., ADD, REMOVE
	Added        []any             `json:"added,omitempty"`
	Removed      []any             `json:"removed,omitempty"`
	ItemType     string            `json:"$type,omitempty"` // e.g., IssueCreatedActivityItem, CommentActivityItem
}

// ActivityCursorPage يمثل الصفحة الكاملة للأنشطة بنظام المؤشرات (Cursor Pagination)
type ActivityCursorPage struct {
	Cursor       string          `json:"cursor"`
	BeforeCursor string          `json:"beforeCursor,omitempty"`
	AfterCursor  string          `json:"afterCursor,omitempty"`
	HasBefore    bool            `json:"hasBefore"`
	HasAfter     bool            `json:"hasAfter"`
	Activities   []*ActivityItem `json:"activities"`
	Type         string          `json:"$type,omitempty"`
}
