package domain

// ProjectTeam يمثل فريق العمل الخاص بالمشروع
type ProjectTeam struct {
	ID            string    `json:"id"`
	Name          string    `json:"name,omitempty"`
	Icon          string    `json:"icon,omitempty"`
	Project       *Project  `json:"project,omitempty"`
	AuditTargetID string    `json:"auditTargetId,omitempty"`
	IsUpdatable   bool      `json:"isUpdatable,omitempty"`
	IsRemovable   bool      `json:"isRemovable,omitempty"`
	Users         []*User   `json:"users,omitempty"`
	Type          string    `json:"$type,omitempty"`
}

// ProjectPlugins ملحقات وتكاملات المشروع
type ProjectPlugins struct {
	TimeTrackingSettings   *ProjectTimeTrackingSettings `json:"timeTrackingSettings,omitempty"`
	HelpDeskSettings       *ProjectHelpDeskSettings     `json:"helpDeskSettings,omitempty"`
	VcsIntegrationSettings *ProjectVcsIntegration       `json:"vcsIntegrationSettings,omitempty"`
	Grazie                 *ProjectGraziePlugin         `json:"grazie,omitempty"`
	Type                   string                       `json:"$type,omitempty"`
}

// ProjectGraziePlugin تكامل JetBrains Grazie للتدقيق
type ProjectGraziePlugin struct {
	Disabled bool   `json:"disabled"`
	Type     string `json:"$type,omitempty"`
}

// ProjectHelpDeskSettings إعدادات HelpDesk للمشروع
type ProjectHelpDeskSettings struct {
	ID          string `json:"id"`
	DefaultForm any    `json:"defaultForm,omitempty"`
	Type        string `json:"$type,omitempty"`
}

// ProjectVcsIntegration تكاملات نظام VCS للمشروع
type ProjectVcsIntegration struct {
	HasVcsIntegrations bool `json:"hasVcsIntegrations"`
	Type               string `json:"$type,omitempty"`
}

// Organization يمثل المنظمة التي تضم المشروع
type Organization struct {
	ID            string `json:"id"`
	Key           string `json:"key,omitempty"`
	Name          string `json:"name,omitempty"`
	IconURL       string `json:"iconUrl,omitempty"`
	ProjectsCount int    `json:"projectsCount,omitempty"`
	Type          string `json:"$type,omitempty"`
}

// Project يمثل الكيان الرئيسي للمشروع في YouTrack
type Project struct {
	ID                       string             `json:"id"`
	Name                     string             `json:"name"`
	ShortName                string             `json:"shortName"`
	ProjectType              *ProjectType       `json:"projectType,omitempty"`
	Pinned                   bool               `json:"pinned,omitempty"`
	IconURL                  string             `json:"iconUrl,omitempty"`
	Template                 bool               `json:"template,omitempty"`
	Archived                 bool               `json:"archived,omitempty"`
	Restricted               bool               `json:"restricted,omitempty"`
	HasArticles              bool               `json:"hasArticles,omitempty"`
	IsDemo                   bool               `json:"isDemo,omitempty"`
	FieldsSorted             bool               `json:"fieldsSorted,omitempty"`
	Query                    string             `json:"query,omitempty"`
	IssuesURL                string             `json:"issuesUrl,omitempty"`
	Description              string             `json:"description,omitempty"`
	Leader                   *User              `json:"leader,omitempty"`
	Team                     *ProjectTeam       `json:"team,omitempty"`
	Organization             *Organization      `json:"organization,omitempty"`
	CreationTime             int64              `json:"creationTime,omitempty"`
	FromEmail                string             `json:"fromEmail,omitempty"`
	FromPersonal             string             `json:"fromPersonal,omitempty"`
	ReplyToEmail             string             `json:"replyToEmail,omitempty"`
	EmailDelimiter           string             `json:"emailDelimiter,omitempty"`
	UseEmailDelimiter        bool               `json:"useEmailDelimiter,omitempty"`
	SupportsEmailDelimiter   bool               `json:"supportsEmailDelimiter,omitempty"`
	DefaultVisibilityGroup   *UserGroup         `json:"defaultVisibilityGroup,omitempty"`
	RelevantVisibilityGroups []*UserGroup       `json:"relevantVisibilityGroups,omitempty"`
	DefaultSmtp              bool               `json:"defaultSmtp,omitempty"`
	SourceTemplate           string             `json:"sourceTemplate,omitempty"`
	AuditTargetID            string             `json:"auditTargetId,omitempty"`
	HistoricalShortNames     []string           `json:"historicalShortNames,omitempty"`
	Plugins                  *ProjectPlugins    `json:"plugins,omitempty"`
	CreatedAt                int64              `json:"createdAt,omitempty"`
	UpdatedAt                int64              `json:"updatedAt,omitempty"`
	Type                     string             `json:"$type,omitempty"`
}

// ProjectType نوع المشروع
type ProjectType struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// ProjectPeople أعضاء المشروع وصلاحياتهم
type ProjectPeople struct {
	UsersInTeam              []*User        `json:"usersInTeam,omitempty"`
	GroupsInTeam             []*UserGroup   `json:"groupsInTeam,omitempty"`
	OtherUsersWithAccess     []*User        `json:"otherUsersWithAccess,omitempty"`
	OtherGroupsWithAccess    []*UserGroup   `json:"otherGroupsWithAccess,omitempty"`
	TotalUsersInTeamCount    int            `json:"totalUsersInTeamCount,omitempty"`
	Type                     string         `json:"$type,omitempty"`
}

// ProjectDashboard لوحة تحكم المشروع
type ProjectDashboard struct {
	Widgets []*DashboardWidgetEmbedding `json:"widgets,omitempty"`
	Type    string                      `json:"$type,omitempty"`
}

// DashboardWidgetEmbedding أداة مدمجة في لوحة تحكم المشروع
type DashboardWidgetEmbedding struct {
	ID       string       `json:"id"`
	Key      string       `json:"key,omitempty"`
	X        int          `json:"x,omitempty"`
	Y        int          `json:"y,omitempty"`
	Width    int          `json:"width,omitempty"`
	Height   int          `json:"height,omitempty"`
	Widget   *WidgetRef   `json:"widget,omitempty"`
	Settings string       `json:"settings,omitempty"`
	Type     string       `json:"$type,omitempty"`
}

// WidgetRef مرجع لأداة Dashboard
type WidgetRef struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// ProjectSavedQueries الاستعلام المحفوظة الخاص بالمشروع
type ProjectSavedQueries struct {
	ID     string       `json:"id"`
	Queries []*SavedQuery `json:"queries,omitempty"`
	Type   string       `json:"$type,omitempty"`
}

// ProjectCustomFields حقول مخصصة خاصة بالمشروع
type ProjectCustomFields struct {
	ID     string                `json:"id"`
	Fields []*ProjectCustomField `json:"fields,omitempty"`
	Type   string                `json:"$type,omitempty"`
}

// ProjectIssueLinkTypes أنواع ربط التذاكر في المشروع
type ProjectIssueLinkTypes struct {
	ID       string           `json:"id"`
	LinkTypes []*IssueLinkType `json:"linkTypes,omitempty"`
	Type     string           `json:"$type,omitempty"`
}

// ProjectBuilds أجهزة البناء في المشروع
type ProjectBuilds struct {
	ID        string         `json:"id"`
	BuildTypes []*BuildType  `json:"buildTypes,omitempty"`
	Type      string         `json:"$type,omitempty"`
}

// BuildType نوع جهاز البناء
type BuildType struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"$type,omitempty"`
}

// ProjectVcsChanges معلومات VCS Changes للمشروع
type ProjectVcsChanges struct {
	ID      string `json:"id"`
	Type    string `json:"$type,omitempty"`
}
