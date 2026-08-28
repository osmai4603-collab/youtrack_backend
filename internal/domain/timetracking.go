package domain

// WorkTimeSettings إعدادات وقت العمل العامة
type WorkTimeSettings struct {
	ID                      int    `json:"id,omitempty"`
	DaysAWeek               int    `json:"daysAWeek"`
	MinutesADay             int    `json:"minutesADay"`
	WorkDays                []int  `json:"workDays,omitempty"`
	MinutesADayPresentation string `json:"minutesADayPresentation,omitempty"`
	FirstDayOfWeek          int    `json:"firstDayOfWeek"`
	Type                    string `json:"$type,omitempty"`
}

// WorkItemType نوع العمل المسجل (مثل تطوير، فحص، توثيق)
type WorkItemType struct {
	ID          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Color       *FieldStyle `json:"color,omitempty"`
	AutoAttach  bool        `json:"autoAttach,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"$type,omitempty"`
}

// ProjectTimeTrackingSettings إعدادات تتبع الوقت الخاصة بمشروع معين
type ProjectTimeTrackingSettings struct {
	ID             string              `json:"id"`
	Project        *Project            `json:"project,omitempty"`
	Enabled        bool                `json:"enabled"`
	EstimateField  *ProjectCustomField `json:"estimate,omitempty"`
	TimeSpentField *ProjectCustomField `json:"timeSpent,omitempty"`
	Type           string              `json:"$type,omitempty"`
}

// WorkItem بند العمل المسجل (الزمن المستغرق والوصف والتاريخ)
type WorkItem struct {
	ID              string        `json:"id"`
	Issue           *Issue        `json:"issue,omitempty"`
	Author          *User         `json:"author,omitempty"`
	Creator         *User         `json:"creator,omitempty"`
	WorkType        *WorkItemType `json:"workType,omitempty"`
	DurationMinutes int           `json:"durationMinutes"`
	Date            int64         `json:"date"`
	Text            string        `json:"text,omitempty"`
	Created         int64         `json:"created,omitempty"`
	Updated         int64         `json:"updated,omitempty"`
	Type            string        `json:"$type,omitempty"`
}

// AttributePrototype نموذج بروتوكول الحقل المخصص لتسجيل الوقت
type AttributePrototype struct {
	ID             string                      `json:"id"`
	Name           string                      `json:"name,omitempty"`
	HasRunningJobs bool                        `json:"hasRunningJobs,omitempty"`
	Values         []*WorkItemAttributeValue   `json:"values,omitempty"`
	Instances      []*WorkItemProjectAttribute `json:"instances,omitempty"`
	Type           string                      `json:"$type,omitempty"`
}

// WorkItemAttributeValue قيمة حقل تسجيل الوقت
type WorkItemAttributeValue struct {
	ID             string                      `json:"id"`
	Name           string                      `json:"name,omitempty"`
	AutoAttach     bool                        `json:"autoAttach,omitempty"`
	Description    *string                     `json:"description,omitempty"`
	HasRunningJobs bool                        `json:"hasRunningJobs,omitempty"`
	Color          *FieldStyle                 `json:"color,omitempty"`
	Attributes     []*WorkItemProjectAttribute `json:"attributes,omitempty"`
	Type           string                      `json:"$type,omitempty"`
}

// WorkItemProjectAttribute ربط الحقل بمشروع معين
type WorkItemProjectAttribute struct {
	ID                   string                       `json:"id"`
	TimeTrackingSettings *ProjectTimeTrackingSettings `json:"timeTrackingSettings,omitempty"`
	Type                 string                       `json:"$type,omitempty"`
}

// AggregatedUser بيانات مستخدم مجمّعة في تقرير تتبع الوقت
type AggregatedUser struct {
	User      *User            `json:"user,omitempty"`
	TotalTime int              `json:"totalTime,omitempty"`
	WorkItems []*WorkItem      `json:"workItems,omitempty"`
	Summary   *WorkItemSummary `json:"summary,omitempty"`
	Type      string           `json:"$type,omitempty"`
}

// WorkItemSummary ملخص بيانات تسجيل الوقت
type WorkItemSummary struct {
	TotalTime    int    `json:"totalTime,omitempty"`
	Presentation string `json:"presentation,omitempty"`
	Type         string `json:"$type,omitempty"`
}

// BoardTimeTrackingData بيانات تتبع الوقت الخاصة بلوحة الأجايل
type BoardTimeTrackingData struct {
	ID    string           `json:"id"`
	Agile *AgileBoard      `json:"agile,omitempty"`
	Field *FieldDefinition `json:"field,omitempty"`
	Type  string           `json:"$type,omitempty"`
}

// WorkItemActivity سجل نشاط تسجيل الوقت
type WorkItemActivity struct {
	ID        string        `json:"id"`
	Author    *User         `json:"author,omitempty"`
	Timestamp int64         `json:"timestamp,omitempty"`
	Text      string        `json:"text,omitempty"`
	Duration  *WorkDuration `json:"duration,omitempty"`
	Type      string        `json:"$type,omitempty"`
}

// WorkDuration مدة العمل
type WorkDuration struct {
	Presentation string `json:"presentation,omitempty"`
	Minutes      int    `json:"minutes,omitempty"`
	Type         string `json:"$type,omitempty"`
}
