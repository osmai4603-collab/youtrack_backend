package model

// ProjectType يصف نوع المشروع (DEFAULT, HELPDESK, SCRUM, KANBAN)
type ProjectType struct {
	ID   string `json:"id"`
	Type string `json:"$type"`
}

// Project يصف مشروعًا في النظام بكامل تفاصيله (Request #12).
type Project struct {
	ID                     string                `json:"id" db:"id"`
	Name                   string                `json:"name" db:"name"`
	ShortName              string                `json:"shortName" db:"short_name"`
	ProjectTypeID          string                `json:"-" db:"project_type_id"`
	ProjectType            *ProjectType          `json:"projectType,omitempty"`
	Pinned                 bool                  `json:"pinned,omitempty" db:"pinned"`
	Template               bool                  `json:"template,omitempty" db:"template"`
	Archived               bool                  `json:"archived,omitempty" db:"archived"`
	Restricted             bool                  `json:"restricted,omitempty" db:"restricted"`
	HasArticles            bool                  `json:"hasArticles,omitempty" db:"has_articles"`
	IsDemo                 bool                  `json:"isDemo,omitempty" db:"is_demo"`
	LeaderID               string                `json:"-" db:"leader_id"`
	Leader                 *User                 `json:"leader,omitempty"`
	OrganizationID         string                `json:"-" db:"organization_id"`
	Description            string                `json:"description,omitempty" db:"description"`
	CreatedAt              int64                 `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt              int64                 `json:"updatedAt,omitempty" db:"updated_at"`
	IconURL                string                `json:"iconUrl,omitempty" db:"icon_url"`
	SourceTemplate         string                `json:"sourceTemplate,omitempty" db:"source_template"`
	FieldsSorted           bool                  `json:"fieldsSorted,omitempty" db:"fields_sorted"`
	Query                  string                `json:"query,omitempty" db:"query"`
	IssuesURL              string                `json:"issuesUrl,omitempty" db:"issues_url"`
	CreationTime           int64                 `json:"creationTime,omitempty" db:"creation_time"`
	Widgets                []*DashboardWidget    `json:"widgets,omitempty"`
	Plugins                *ProjectPlugins       `json:"plugins,omitempty"`
	DefaultVisibilityGroup *UserGroup            `json:"defaultVisibilityGroup,omitempty"`
	RelevantVisibilityGroups []*UserGroup        `json:"relevantVisibilityGroups,omitempty"`
	HistoricalShortNames   []string              `json:"historicalShortNames,omitempty" db:"historical_short_names"`
	AuditTargetID          string                `json:"auditTargetId,omitempty" db:"audit_target_id"`
	Organization           *Organization         `json:"organization,omitempty"`
	DefaultSmtp            bool                  `json:"defaultSmtp,omitempty" db:"default_smtp"`
	FromEmail              string                `json:"fromEmail,omitempty" db:"from_email"`
	FromPersonal           string                `json:"fromPersonal,omitempty" db:"from_personal"`
	ReplyToEmail           string                `json:"replyToEmail,omitempty" db:"reply_to_email"`
	SupportsEmailDelimiter bool                  `json:"supportsEmailDelimiter,omitempty" db:"supports_email_delimiter"`
	EmailDelimiter         string                `json:"emailDelimiter,omitempty" db:"email_delimiter"`
	UseEmailDelimiter      bool                  `json:"useEmailDelimiter,omitempty" db:"use_email_delimiter"`
	Team                   *ProjectTeamDetailed  `json:"team,omitempty"`
	Type                   string                `json:"$type,omitempty"`
}

func (p *Project) Normalize() {
	if p.Type == "" {
		p.Type = "Project"
	}
	if p.ProjectType != nil && p.ProjectType.Type == "" {
		p.ProjectType.Type = "ProjectType"
	}
	if p.Team != nil && p.Team.Type == "" {
		p.Team.Type = "ProjectTeam"
	}
	if p.Organization != nil && p.Organization.Type == "" {
		p.Organization.Type = "Organization"
	}
	if p.Plugins != nil && p.Plugins.Type == "" {
		p.Plugins.Type = "ProjectPlugins"
	}
	if p.Plugins != nil {
		if p.Plugins.TimeTrackingSettings != nil && p.Plugins.TimeTrackingSettings.Type == "" {
			p.Plugins.TimeTrackingSettings.Type = "ProjectTimeTrackingSettings"
		}
		if p.Plugins.HelpDeskSettings != nil && p.Plugins.HelpDeskSettings.Type == "" {
			p.Plugins.HelpDeskSettings.Type = "ProjectHelpDeskSettings"
		}
		if p.Plugins.VcsIntegrationSettings != nil && p.Plugins.VcsIntegrationSettings.Type == "" {
			p.Plugins.VcsIntegrationSettings.Type = "ProjectVcsIntegrationSettings"
		}
		if p.Plugins.Grazie != nil && p.Plugins.Grazie.Type == "" {
			p.Plugins.Grazie.Type = "ProjectGraziePlugin"
		}
	}
	if p.Team != nil {
		if p.Team.TeamForProject != nil && p.Team.TeamForProject.Type == "" {
			p.Team.TeamForProject.Type = "Project"
		}
		if len(p.Team.Users) == 0 {
			p.Team.Users = nil
		}
	}
}

// ProjectTeamDetailed يصف فريق المشروع بالتفصيل حسب طلب #12.
type ProjectTeamDetailed struct {
	ID             string      `json:"id" db:"id"`
	Name           string      `json:"name" db:"name"`
	AuditTargetID  string      `json:"auditTargetId,omitempty" db:"audit_target_id"`
	Description    string      `json:"description,omitempty" db:"description"`
	AllUsersGroup  bool        `json:"allUsersGroup" db:"all_users_group"`
	Icon           string      `json:"icon,omitempty" db:"icon"`
	TeamForProject *ProjectRef `json:"teamForProject,omitempty"`
	IsUpdatable    bool        `json:"isUpdatable" db:"is_updatable"`
	IsRemovable    bool        `json:"isRemovable" db:"is_removable"`
	Users          []*User     `json:"users,omitempty"`
	Type           string      `json:"$type,omitempty"`
}

// ProjectRef مرجع بسيط للمشروع يستخدم داخل الفريق.
type ProjectRef struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name,omitempty" db:"name"`
	Icon string `json:"icon,omitempty" db:"icon"`
	Type string `json:"$type,omitempty"`
}

// ProjectPlugins يجمع إعدادات الإضافات للمشروع.
type ProjectPlugins struct {
	TimeTrackingSettings   *TimeTrackingSettings   `json:"timeTrackingSettings,omitempty"`
	HelpDeskSettings       *HelpDeskSettings       `json:"helpDeskSettings,omitempty"`
	VcsIntegrationSettings *VcsIntegrationSettings `json:"vcsIntegrationSettings,omitempty"`
	Grazie                 *GrazieSettings         `json:"grazie,omitempty"`
	Type                   string                  `json:"$type,omitempty"`
}

type TimeTrackingSettings struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"$type,omitempty"`
}

type HelpDeskSettings struct {
	ID          string        `json:"id"`
	DefaultForm *HelpDeskForm `json:"defaultForm,omitempty"`
	Type        string        `json:"$type,omitempty"`
}

type HelpDeskForm struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Type  string `json:"$type,omitempty"`
}

type VcsIntegrationSettings struct {
	HasVcsIntegrations bool   `json:"hasVcsIntegrations"`
	Type               string `json:"$type,omitempty"`
}

type GrazieSettings struct {
	Disabled bool   `json:"disabled"`
	Type     string `json:"$type,omitempty"`
}

// ProjectTeam يصف فريق مشروع (مبسط).
type ProjectTeam struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	ProjectID string `json:"projectId" db:"project_id"`
}

// Organization يصف منظمة (مجموعة مشاريع).
type Organization struct {
	ID            string     `json:"id" db:"id"`
	Key           string     `json:"key,omitempty" db:"key"`
	Name          string     `json:"name,omitempty" db:"name"`
	IconURL       *string    `json:"iconUrl,omitempty" db:"icon_url"`
	ProjectsCount int        `json:"projectsCount,omitempty" db:"projects_count"`
	AuditTargetID string     `json:"auditTargetId,omitempty" db:"audit_target_id"`
	Description   string     `json:"description,omitempty" db:"description"`
	Projects      []*Project `json:"projects,omitempty"`
	Type          string     `json:"$type,omitempty"`
}

// Normalize يضبط القيم الافتراضية لكائن المنظمة.
func (o *Organization) Normalize() {
	if o.Type == "" {
		o.Type = "Organization"
	}
}
