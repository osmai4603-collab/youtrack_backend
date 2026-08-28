package domain

// AgileBoard يمثل لوحة الأجايل
type AgileBoard struct {
	ID                         string               `json:"id"`
	Name                       string               `json:"name,omitempty"`
	IsDemo                     bool                 `json:"isDemo,omitempty"`
	IsUpdatable                bool                 `json:"isUpdatable,omitempty"`
	Favorite                   bool                 `json:"favorite,omitempty"`
	FlatBacklog                bool                 `json:"flatBacklog,omitempty"`
	HideOrphansSwimlane        bool                 `json:"hideOrphansSwimlane,omitempty"`
	OrphansAtTheTop            bool                 `json:"orphansAtTheTop,omitempty"`
	ColorizeCustomFields       bool                 `json:"colorizeCustomFields,omitempty"`
	CardOnSeveralSprints       bool                 `json:"cardOnSeveralSprints,omitempty"`
	Owner                      *User                `json:"owner,omitempty"`
	EstimationField            *FieldDefinition     `json:"estimationField,omitempty"`
	OriginalEstimationField    *FieldDefinition     `json:"originalEstimationField,omitempty"`
	Backlog                    *SavedQuery          `json:"backlog,omitempty"`
	Projects                   []*Project           `json:"projects,omitempty"`
	Sprints                    []*Sprint            `json:"sprints,omitempty"`
	ColumnSettings             *ColumnSettings      `json:"columnSettings,omitempty"`
	SwimlaneSettings           *SwimlaneSettings    `json:"swimlaneSettings,omitempty"`
	CreatedWithOriginal        bool                 `json:"createdWithOriginal,omitempty"`
	Type                       string               `json:"$type,omitempty"`
}

// Sprint يمثل السباق داخل اللوحة
type Sprint struct {
	ID         string       `json:"id"`
	Agile      *AgileBoard  `json:"agile,omitempty"`
	Name       string       `json:"name,omitempty"`
	Start      *int64       `json:"start,omitempty"`
	Finish     *int64       `json:"finish,omitempty"`
	Goal       string       `json:"goal,omitempty"`
	Ordinal    int          `json:"ordinal,omitempty"`
	Archived   bool         `json:"archived,omitempty"`
	IsDefault  bool         `json:"isDefault,omitempty"`
	IsStarted  bool         `json:"isStarted,omitempty"`
	ReportID   string       `json:"reportId,omitempty"`
	Board      *AgileBoard  `json:"board,omitempty"`
	Type       string       `json:"$type,omitempty"`
}

// ColumnSettings إعدادات الأعمدة للوحة
type ColumnSettings struct {
	ID                 string         `json:"id"`
	Columns            []*BoardColumn `json:"columns,omitempty"`
	Field              *FieldDefinition `json:"field,omitempty"`
	ShowBundleWarning  bool           `json:"showBundleWarning,omitempty"`
	Type               string         `json:"$type,omitempty"`
}

// BoardColumn يمثل العمود في لوحة الأجايل
type BoardColumn struct {
	ID          string                     `json:"id"`
	Agile       *AgileBoard                `json:"agile,omitempty"`
	Collapsed   bool                       `json:"collapsed,omitempty"`
	Ordinal     int                        `json:"ordinal,omitempty"`
	IsResolved  bool                       `json:"isResolved,omitempty"`
	IsVisible   bool                       `json:"isVisible,omitempty"`
	Color       *FieldStyle                `json:"color,omitempty"`
	Parent      *BoardColumn               `json:"parent,omitempty"`
	WipLimit    *WipLimit                  `json:"wipLimit,omitempty"`
	FieldValues []*BoardColumnFieldValue   `json:"fieldValues,omitempty"`
	Type        string                     `json:"$type,omitempty"`
}

// WipLimit يحدد سقف المهام تحت الإجراء في العمود
type WipLimit struct {
	Min  int    `json:"min"`
	Max  int    `json:"max"`
	Type string `json:"$type,omitempty"`
}

// BoardColumnFieldValue قيم الحقول المخصصة التي تحدد العمود
type BoardColumnFieldValue struct {
	ID           string `json:"id"`
	Name         string `json:"name,omitempty"`
	Presentation string `json:"presentation,omitempty"`
	Ordinal      int    `json:"ordinal,omitempty"`
	IsResolved   bool   `json:"isResolved,omitempty"`
	CanUpdate    bool   `json:"canUpdate,omitempty"`
	Type         string `json:"$type,omitempty"`
}

// BoardCell يمثل الخلية في لوحة الأجايل (نقطة تقاطع العمود مع المسار)
type BoardCell struct {
	ID            string       `json:"id"`
	Column        *BoardColumn `json:"column,omitempty"`
	Row           any          `json:"row,omitempty"`
	IssuesCount   int          `json:"issuesCount,omitempty"`
	TooManyIssues bool         `json:"tooManyIssues,omitempty"`
	Issues        []*Issue     `json:"issues,omitempty"`
	Type          string       `json:"$type,omitempty"`
}

// SwimlaneSettings إعدادات مسارات السباحة (الأفقية) على اللوحة
type SwimlaneSettings struct {
	ID              string           `json:"id"`
	Agile           *AgileBoard      `json:"agile,omitempty"`
	Enabled         bool             `json:"enabled,omitempty"`
	Field           *FieldDefinition `json:"field,omitempty"`
	DefaultCardType any              `json:"defaultCardType,omitempty"`
	Type            string           `json:"$type,omitempty"`
}

// AgileUserProfile ملف تعريف المستخدم الخاص بلوحة الأجايل
type AgileUserProfile struct {
	CardDetailLevel int              `json:"cardDetailLevel,omitempty"`
	DefaultAgile    *AgileBoard      `json:"defaultAgile,omitempty"`
	VisitedSprints  []*Sprint        `json:"visitedSprints,omitempty"`
	Type            string           `json:"$type,omitempty"`
}

// AgileStatus حالة لوحة الأجايل
type AgileStatus struct {
	Valid  bool     `json:"valid,omitempty"`
	Errors []string `json:"errors,omitempty"`
	Type   string   `json:"$type,omitempty"`
}

// SprintRef مرجع سباق
type SprintRef struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// AgileRef مرجع لوحة أجايل
type AgileRef struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// BoardIssueLink ربط تذكرة في لوحة الأجايل
type BoardIssueLink struct {
	ID      string `json:"id"`
	Dir     string `json:"dir,omitempty"`
	Link    *IssueLinkType `json:"link,omitempty"`
	issues []*Issue `json:"issues,omitempty"`
	Type    string  `json:"$type,omitempty"`
}

// BoardSwimlane مسار سباحة في لوحة الأجايل
type BoardSwimlane struct {
	ID             string       `json:"id"`
	Name           string       `json:"name,omitempty"`
	Issues         []*Issue     `json:"issues,omitempty"`
	IssuesCount    int          `json:"issuesCount,omitempty"`
	ParentSwimlane *BoardSwimlane `json:"parentSwimlane,omitempty"`
	SubIssuesCount int          `json:"subIssuesCount,omitempty"`
	Type           string       `json:"$type,omitempty"`
}

// ReportSettings إعدادات التقارير
type ReportSettings struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Report *Report `json:"report,omitempty"`
	Type   string `json:"$type,omitempty"`
}

// Report تقرير
type Report struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
	Type       string `json:"$type,omitempty"`
}

// Extensions إضافات لوحة الأجايل
type Extensions struct {
	ID            string `json:"id"`
	TimeTracking  *BoardTimeTrackingData `json:"timeTracking,omitempty"`
	Type          string `json:"$type,omitempty"`
}

// Ownership امتلاك
type Ownership struct {
	Type string `json:"$type,omitempty"`
}
