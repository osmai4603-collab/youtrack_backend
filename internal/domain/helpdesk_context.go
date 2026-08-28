package domain

// HelpdeskContext يمثل سياق مكتب المساعدة (Helpdesk) المرتبط بمستخدم أو بحث.
// يستخدم ضمن GeneralUserProfile في الحقلين: searchContext و helpdeskContext،
// ويطابق الهيكل المرجعي @helpdeskContext في request1.txt:
// id,name,issuesUrl,pinned,pinnedInHelpdesk,owner($permittedUsers),$type,query,isUpdatable,shortName
type HelpdeskContext struct {
	ID               string `json:"id"`
	Name             string `json:"name,omitempty"`
	IssuesURL        string `json:"issuesUrl,omitempty"`
	Pinned           bool   `json:"pinned,omitempty"`
	PinnedInHelpdesk bool   `json:"pinnedInHelpdesk,omitempty"`
	Owner            *User  `json:"owner,omitempty"`
	Query            string `json:"query,omitempty"`
	IsUpdatable      bool   `json:"isUpdatable,omitempty"`
	ShortName        string `json:"shortName,omitempty"`
	Type             string `json:"$type,omitempty"`
}
