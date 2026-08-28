package domain

// IssueListSubscription اشتراك قائمة التذاكر
type IssueListSubscription struct {
	Ticket string `json:"ticket,omitempty"`
	Type   string `json:"$type,omitempty"`
}

// PublicSettings إعدادات الحساب العامة
type PublicSettings struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// HubUser بيانات المستخدم من Hub
type HubUser struct {
	ID              string       `json:"id"`
	Name            string       `json:"name,omitempty"`
	Login           string       `json:"login,omitempty"`
	Guest           bool         `json:"guest,omitempty"`
	Profile         *HubProfile  `json:"profile,omitempty"`
	RequiredTwoFactorAuthentication bool `json:"requiredTwoFactorAuthentication,omitempty"`
	Type            string       `json:"type,omitempty"`
}

// HubProfile ملف المستخدم في Hub
type HubProfile struct {
	Avatar *HubAvatar `json:"avatar,omitempty"`
	Email  *HubEmail  `json:"email,omitempty"`
}

// HubAvatar صورة المستخدم في Hub
type HubAvatar struct {
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
}

// HubEmail البريد الإلكتروني للمستخدم في Hub
type HubEmail struct {
	Email    string `json:"email,omitempty"`
	Verified bool   `json:"verified,omitempty"`
	Type     string `json:"type,omitempty"`
}

// ServicesPage صفحة الخدمات
type ServicesPage struct {
	ID string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// QuestionnaireUserProfile استبيان ملف المستخدم
type QuestionnaireUserProfile struct {
	ShowSurvey               bool   `json:"showSurvey,omitempty"`
	ShowPmfSurvey            bool   `json:"showPmfSurvey,omitempty"`
	DemoEligibilityTimestamp  *int64 `json:"demoEligibilityTimestamp,omitempty"`
	Type                     string `json:"$type,omitempty"`
}

// License 许可证
type License struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// Application 应用
type Application struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"$type,omitempty"`
}
