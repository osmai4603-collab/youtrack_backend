package model

// HubService يصف خدمة مسجّلة في Hub (مطابق لـ request24.txt).
// الحقل Type يُسلسَل كـ "type" وليس "$type" كما في استجابة Hub الأصلية.
type HubService struct {
	Type                         string `json:"type,omitempty" db:"-"`
	ID                           string `json:"id,omitempty" db:"id"`
	Name                         string `json:"name,omitempty" db:"name"`
	Key                          string `json:"key,omitempty" db:"key"`
	HomeURL                      string `json:"homeUrl,omitempty" db:"home_url"`
	ApplicationName              string `json:"applicationName,omitempty" db:"application_name"`
	Vendor                       string `json:"vendor,omitempty" db:"vendor"`
	Version                      string `json:"version,omitempty" db:"version"`
	Trusted                      *bool  `json:"trusted,omitempty" db:"trusted"`
	IconURL                      string `json:"iconUrl,omitempty" db:"icon_url"`
	UserUriPattern               string `json:"userUriPattern,omitempty" db:"user_uri_pattern"`
	GroupUriPattern              string `json:"groupUriPattern,omitempty" db:"group_uri_pattern"`
	Audience                     string `json:"audience,omitempty" db:"audience"`
	Immutable                    *bool  `json:"immutable,omitempty" db:"immutable"`
	ClientCredentialsFlowEnabled *bool  `json:"clientCredentialsFlowEnabled,omitempty" db:"client_credentials_flow_enabled"`
	AuthCodeFlowEnabled          *bool  `json:"authCodeFlowEnabled,omitempty" db:"auth_code_flow_enabled"`
	ImplicitFlowEnabled          *bool  `json:"implicitFlowEnabled,omitempty" db:"implicit_flow_enabled"`
}

// Normalize يضبط القيم الافتراضية لكائن الخدمة.
func (s *HubService) Normalize() {
	if s.Type == "" {
		s.Type = "service"
	}
}

// ServicesPage يصف استجابة قائمة الخدمات (مطابق لـ request24.txt).
type ServicesPage struct {
	Type     string        `json:"type"`
	Skip     int           `json:"skip,omitempty"`
	Top      int           `json:"top,omitempty"`
	Total    int           `json:"total,omitempty"`
	Services []*HubService `json:"services,omitempty"`
}
