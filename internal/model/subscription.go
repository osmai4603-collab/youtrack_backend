package model

// IssueListSubscriptionBean يمثّل اشتراكًا في تحديثات قائمة المشاكل
// (Request #18: /api/issueListSubscription).
type IssueListSubscriptionBean struct {
	ID          string                       `json:"id,omitempty" db:"id"`
	Ticket      string                       `json:"ticket,omitempty" db:"ticket"`
	Query       string                       `json:"query,omitempty" db:"query"`
	Subscribe   bool                         `json:"subscribe" db:"subscribe"`
	ContextType string                       `json:"-" db:"context_type"`
	ContextID   string                       `json:"-" db:"context_id"`
	Context     *SubscriptionContext         `json:"context,omitempty"`
	Issues      []*IssueListSubscriptionItem `json:"issues,omitempty"`
	FolderID    string                       `json:"folderId,omitempty" db:"folder_id"`
	Type        string                       `json:"$type" db:"type"`
}

// SubscriptionContext يصف سياق الاشتراك (مشروع أو مجلد مثلاً).
type SubscriptionContext struct {
	Type string `json:"$type"`
	ID   string `json:"id"`
}

// IssueListSubscriptionItem يصف مشكلة مرتبطة بالاشتراك ضمن قائمة المشاكل.
type IssueListSubscriptionItem struct {
	ID      string `json:"id" db:"issue_id"`
	Matches bool   `json:"matches" db:"matches"`
	Type    string `json:"$type"`
}

// IssueListSubscriptionRequest يمثّل جسم طلب إنشاء/تحديث الاشتراك.
type IssueListSubscriptionRequest struct {
	Query     string                       `json:"query"`
	Issues    []*IssueListSubscriptionItem `json:"issues"`
	Subscribe *bool                        `json:"subscribe"`
	Context   *SubscriptionContext         `json:"context"`
	FolderID  string                       `json:"folderId"`
}
