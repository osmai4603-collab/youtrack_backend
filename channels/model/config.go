package model

// Config يمثّل إعدادات النظام العامة (Frontend Config).
// مستوحى من بنية Mattermost حيث يكون التكوين نموذجًا واحدًا قابلًا للأهداف الافتراضية.
type Config struct {
	Version           string `json:"version,omitempty"`
	Build             string `json:"build,omitempty"`
	ReleaseDate       int64  `json:"releaseDate,omitempty"`
	DefaultPage       string `json:"defaultPage,omitempty"`
	ContextPath       string `json:"contextPath,omitempty"`
	ReadOnly          bool   `json:"readOnly,omitempty"`
	StatisticsEnabled bool   `json:"statisticsEnabled,omitempty"`
	HelpdeskEnabled   bool   `json:"helpdeskEnabled,omitempty"`
	MaxUploadFileSize int64  `json:"maxUploadFileSize,omitempty"`
	MaxExportItems    int    `json:"maxExportItems,omitempty"`
	SSEPingTimeoutMs  int64  `json:"ssePingTimeoutMs,omitempty"`
	LogoURL           string `json:"logoUrl,omitempty"`
	WideLogoURL       string `json:"wideLogoUrl,omitempty"`
	SquareLogoURL     string `json:"squareLogoUrl,omitempty"`
}

// SetDefaults يطبّق قيمًا افتراضية لحقول التكوين غير المعبأة.
func (c *Config) SetDefaults() {
	if c.Version == "" {
		c.Version = "2025.1"
	}
	if c.Build == "" {
		c.Build = "0"
	}
	if c.DefaultPage == "" {
		c.DefaultPage = "/projects"
	}
	if c.ContextPath == "" {
		c.ContextPath = "/"
	}
	if c.MaxUploadFileSize == 0 {
		c.MaxUploadFileSize = 50 * 1024 * 1024
	}
	if c.MaxExportItems == 0 {
		c.MaxExportItems = 1000
	}
	if c.SSEPingTimeoutMs == 0 {
		c.SSEPingTimeoutMs = 10000
	}
}

// GlobalSettings يمثّل بعض الإعدادات الإدارية العامة.
type GlobalSettings struct {
	ID                int    `json:"-" db:"id"`
	Version           string `json:"version,omitempty" db:"version"`
	Build             string `json:"build,omitempty" db:"build"`
	ReadOnly          bool   `json:"readOnly,omitempty" db:"read_only"`
	HelpdeskEnabled   bool   `json:"helpdeskEnabled,omitempty" db:"helpdesk_enabled"`
	MaxUploadFileSize int64  `json:"maxUploadFileSize,omitempty" db:"max_upload_file_size"`
	MaxExportItems    int    `json:"maxExportItems,omitempty" db:"max_export_items"`
}
