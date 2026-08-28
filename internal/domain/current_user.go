package domain

// CurrentUser يمثل المخطط (scheme) الكامل لنقطة /api/users/me
// وفق البنية المرجعية المحفوظة في docs/requests/request1.txt.
// يحتوي الحقول الجذرية فقط المطلوبة، ويقسم الملفات الفرعية إلى scheme فرعية.
type CurrentUser struct {
	Widgets           []*Widget            `json:"widgets"`
	FeatureFlags      []*FeatureFlag       `json:"featureFlags"`
	Guest             bool                 `json:"guest"`
	Banned            bool                 `json:"banned"`
	BanBadge          *string              `json:"banBadge"`
	Login             string               `json:"login"`
	Email             string               `json:"email"`
	UserType          *UserType            `json:"userType"`
	AvatarURL         string               `json:"avatarUrl"`
	Online            bool                 `json:"online"`
	Profiles          *CurrentUserProfiles `json:"profiles"`
	IsEmailVerified   bool                 `json:"isEmailVerified"`
	IssueRelatedGroup *UserGroup           `json:"issueRelatedGroup"`
	CanReadProfile    bool                 `json:"canReadProfile"`
	FullName          string               `json:"fullName"`
	Name              string               `json:"name"`
	IsLocked          bool                 `json:"isLocked"`
	ID                string               `json:"id"`
	Type              string               `json:"$type"`
}

// CurrentUserProfiles يجمع كافة ملفات التفضيلات الخاصة بالمستخدم الحالي.
// البنية مأخوذة من ملف profiles في request1.txt.
type CurrentUserProfiles struct {
	General       *GeneralUserProfile       `json:"general"`
	Articles      *ArticlesUserProfile      `json:"articles"`
	TimeTracking  *TimeTrackingUserProfile  `json:"timetracking"`
	Tips          *TipsUserProfile          `json:"tips"`
	Appearance    *AppearanceUserProfile    `json:"appearance"`
	IssuesList    *IssuesListUserProfile    `json:"issuesList"`
	Helpdesk      *HelpdeskUserProfile      `json:"helpdesk"`
	AI            *AiUserProfile            `json:"ai"`
	Notifications *NotificationsUserProfile `json:"notifications"`
	Type          string                    `json:"$type"`
}
