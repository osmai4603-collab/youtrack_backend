package model

// IssueFolder يصف المجلد (مشروع/استعلام محفوظ/وسم) الذي يُحسب عليه عدد المشاكل.
type IssueFolder struct {
	ID               string `json:"id,omitempty" db:"id"`
	Name             string `json:"name,omitempty" db:"name"`
	ShortName        string `json:"shortName,omitempty" db:"short_name"`
	Query            string `json:"query,omitempty" db:"query"`
	Pinned           *bool  `json:"pinned,omitempty" db:"pinned"`
	PinnedInHelpdesk *bool  `json:"pinnedInHelpdesk,omitempty" db:"pinned_in_helpdesk"`
	Type             string `json:"$type,omitempty"`
}

// IssueCountResponse يمثّل استجابة /api/issuesGetter/count (مطابق لـ request17.txt).
type IssueCountResponse struct {
	Count  int64        `json:"count"`
	Folder *IssueFolder `json:"folder,omitempty"`
	Type   string       `json:"$type,omitempty"`
}

// Normalize يجعل الـ $type ثابتاً على كل المستويات.
func (r *IssueCountResponse) Normalize() {
	r.Type = "IssueCountResponse"
	if r.Folder != nil {
		if r.Folder.Type == "" {
			r.Folder.Type = "Project"
		}
	}
}