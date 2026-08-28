package domain

// UserRepository واجهة التعامل مع بيانات المستخدمين
type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByLogin(login string) (*User, error)
	GetCurrentUser() (*User, error)
	// GetCurrentUserDetail يجلب مخطط المستخدم الحالي وفقاً للأقسام المطلوبة فقط
	// (sel nil = كل شيء). يقلّل الجلب من قاعدة البيانات بناءً على معامل fields.
	GetCurrentUserDetail(id string, sel *UserFieldSelect) (*CurrentUser, error)
	Create(user *User) error
	Update(user *User) error
}

// ProjectRepository واجهة التعامل مع المشاريع
type ProjectRepository interface {
	GetByID(id string) (*Project, error)
	List() ([]*Project, error)
	Create(project *Project) error
}

// IssueRepository واجهة التعامل مع التذاكر والروابط والأنشطة
type IssueRepository interface {
	GetByID(id string) (*Issue, error)
	GetLinkTypes() ([]*IssueLinkType, error)
	GetActivities(issueID string) (*ActivityCursorPage, error)
	Create(issue *Issue) error
}

// AdminRepository واجهة إعدادات الإدارة وساعات العمل
type AdminRepository interface {
	GetWorkTimeSettings() (*WorkTimeSettings, error)
	GetGlobalSettings() (*GlobalSettings, error)
	GetWidgets() ([]*WidgetView, error)
}

// NotificationRepository واجهة التعامل مع الإشعارات والبريد
type NotificationRepository interface {
	GetInboxFolders() ([]*InboxFolder, error)
	GetThreads(folderID string) ([]*InboxThread, error)
}

// SearchRepository واجهة التعامل مع البحث
type SearchRepository interface {
	SearchAssist(query string) (*SearchAssist, error)
}

// ConfigRepository واجهة التعامل مع إعدادات النظام
type ConfigRepository interface {
	GetFrontendConfig() (*FrontendConfig, error)
	GetPublicSettings() (*PublicSettings, error)
}

// AgileRepository واجهة التعامل مع لوحات الأجايل
type AgileRepository interface {
	GetUserProfile() (*AgileUserProfile, error)
	GetBoardExtensions(boardID string) (*Extensions, error)
}

// TimeTrackingRepository واجهة التعامل مع تتبع الوقت
type TimeTrackingRepository interface {
	GetAttributePrototypes() ([]*AttributePrototype, error)
	GetAttributePrototype(id string) (*AttributePrototype, error)
	GetBoardTimeTrackingData(boardID string) (*BoardTimeTrackingData, error)
}

// ProjectPeopleRepository واجهة التعامل مع أعضاء المشروع
type ProjectPeopleRepository interface {
	GetProjectPeople(projectID string, transitiveRolesQuery string) (*ProjectPeople, error)
	GetProjectDashboard(projectID string) (*ProjectDashboard, error)
}

// VCSRepository واجهة التعامل مع خوادم VCS
type VCSRepository interface {
	GetVCSServers() ([]*VCSServer, error)
}

// AppRepository واجهة التعامل مع التطبيقات والخدمات
type AppRepository interface {
	GetServicesPage() (*ServicesPage, error)
}

// HubRepository واجهة التعامل مع Hub
type HubRepository interface {
	GetHubUser() (*HubUser, error)
}
