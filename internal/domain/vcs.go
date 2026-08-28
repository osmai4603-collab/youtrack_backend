package domain

// VCSServer خادم نظام التحكم بال版本 (VCS)
type VCSServer struct {
	ID                string       `json:"id"`
	URL               string       `json:"url,omitempty"`
	ApplicationID     *string      `json:"applicationId,omitempty"`
	AppID             *string      `json:"appId,omitempty"`
	AppName           string       `json:"appName,omitempty"`
	IsPredefined      bool         `json:"isPredefined,omitempty"`
	SSLKey            *SSLKey      `json:"sslKey,omitempty"`
	ChangesProcessors []any        `json:"changesProcessors,omitempty"`
	Type              string       `json:"$type,omitempty"`
}

// SSLKey مفتاح SSL الخاص بخادم VCS
type SSLKey struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"$type,omitempty"`
}
