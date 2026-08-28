package domain

// BannersConfig إعدادات التنبيهات والإعلانات العامة للنظام
type BannersConfig struct {
	GlobalBannerEnabled bool     `json:"globalBannerEnabled"`
	GlobalBanner        string   `json:"globalBanner,omitempty"`
	SystemEventsBanners []string `json:"systemEventsBanners"`
	Type                string   `json:"$type,omitempty"`
}

// FrontendConfig إعدادات واجهة YouTrack العامة (/api/config)
type FrontendConfig struct {
	Banners              *BannersConfig    `json:"banners,omitempty"`
	Ring                 *RingFrontendConfig `json:"ring,omitempty"`
	L10n                 *L10NFrontendConfig `json:"l10n,omitempty"`
	LicenseError         any               `json:"licenseError,omitempty"`
	System               *SystemFrontendConfig `json:"system,omitempty"`
	Shortcuts            *ShortcutScheme   `json:"shortcuts,omitempty"`
	ReadOnly             bool              `json:"readOnly,omitempty"`
	Hosted               *HostedFrontendConfig `json:"hosted,omitempty"`
	Konnector            *KonnectorConfig  `json:"konnector,omitempty"`
	StatisticsEnabled    bool              `json:"statisticsEnabled,omitempty"`
	ContextPath          string            `json:"contextPath,omitempty"`
	DefaultPage          string            `json:"defaultPage,omitempty"`
	LogoURL              string            `json:"logoUrl,omitempty"`
	WideLogoURL          string            `json:"wideLogoUrl,omitempty"`
	SquareLogoURL        string            `json:"squareLogoUrl,omitempty"`
	Version              string            `json:"version,omitempty"`
	Build                string            `json:"build,omitempty"`
	ReleaseDate          int64             `json:"releaseDate,omitempty"`
	FeaturesURL          string            `json:"featuresUrl,omitempty"`
	MarketplaceBaseURL   string            `json:"marketplaceBaseUrl,omitempty"`
	RedirectToWelcomeForm bool             `json:"redirectToWelcomeForm,omitempty"`
	HelpdeskEnabled      bool              `json:"helpdeskEnabled,omitempty"`
	Type                 string            `json:"$type,omitempty"`
}

// RingFrontendConfig إعدادات Hub/Ring
type RingFrontendConfig struct {
	ServiceID       string                `json:"serviceId,omitempty"`
	URL             string                `json:"url,omitempty"`
	SearchBotToken  *string               `json:"searchBotToken,omitempty"`
	Broken          bool                  `json:"broken,omitempty"`
	Enabled         bool                  `json:"enabled,omitempty"`
	HasEmbeddedHub  bool                  `json:"hasEmbeddedHub,omitempty"`
	Services        *RingServicesFrontendConfig `json:"services,omitempty"`
	Type            string                `json:"$type,omitempty"`
}

// RingServicesFrontendConfig خدمات Hub المتصلة
type RingServicesFrontendConfig struct {
	Dashboard any    `json:"dashboard,omitempty"`
	Type      string `json:"$type,omitempty"`
}

// L10NFrontendConfig إعدادات اللغة والتعريب
type L10NFrontendConfig struct {
	IsRTL            bool              `json:"isRTL,omitempty"`
	Locale           string            `json:"locale,omitempty"`
	Language         string            `json:"language,omitempty"`
	PredefinedQueries map[string]string `json:"predefinedQueries,omitempty"`
	Type             string            `json:"$type,omitempty"`
}

// SystemFrontendConfig إعدادات النظام على واجهة المستخدم
type SystemFrontendConfig struct {
	MaxUploadFileSize int    `json:"maxUploadFileSize,omitempty"`
	MaxExportItems    int    `json:"maxExportItems,omitempty"`
	SsePingTimeoutMs  int    `json:"ssePingTimeoutMs,omitempty"`
	Type              string `json:"$type,omitempty"`
}

// HostedFrontendConfig إعدادات الاستضافة السحابية
type HostedFrontendConfig struct {
	Hosted          bool   `json:"hosted,omitempty"`
	Domain          string `json:"domain,omitempty"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
	Type            string `json:"$type,omitempty"`
}

// KonnectorConfig إعدادات Konnector (البوت والتكامل)
type KonnectorConfig struct {
	TelegramBotURL string `json:"telegramBotUrl,omitempty"`
	URL            string `json:"url,omitempty"`
	Type           string `json:"$type,omitempty"`
}

// ShortcutScheme مخطط اختصارات لوحة المفاتيح
type ShortcutScheme struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// GlobalSettings يمثل الإعدادات العامة للنظام وإعدادات الخادم
type GlobalSettings struct {
	ImageTextRecognitionSettings *ImageTextRecognitionSettings `json:"imageTextRecognitionSettings,omitempty"`
	NotificationSettings         *NotificationSettings         `json:"notificationSettings,omitempty"`
	RestSettings                 *RestSettings                 `json:"restSettings,omitempty"`
	SystemSettings               *SystemSettings               `json:"systemSettings,omitempty"`
	Type                         string                        `json:"$type,omitempty"`
}

// ImageTextRecognitionSettings إعدادات التعرف الضوئي على النصوص من الصور (OCR)
type ImageTextRecognitionSettings struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"$type,omitempty"`
}

// NotificationSettings إعدادات الإشعارات العامة للنظام
type NotificationSettings struct {
	EmailSettings *EmailSettings `json:"emailSettings,omitempty"`
	Type          string         `json:"$type,omitempty"`
}

// EmailSettings إعدادات خادم وإشعارات البريد الإلكتروني
type EmailSettings struct {
	IsEnabled bool   `json:"isEnabled"`
	Type      string `json:"$type,omitempty"`
}

// RestSettings إعدادات CORS وطلبات واجهة REST API
type RestSettings struct {
	AllowAllOrigins bool     `json:"allowAllOrigins"`
	AllowedOrigins  []string `json:"allowedOrigins,omitempty"`
	Type            string   `json:"$type,omitempty"`
}

// SystemSettings إعدادات إمكانيات النظام ودعم الميزات على مستوى الخادم
type SystemSettings struct {
	OcrSupported bool   `json:"ocrSupported"`
	Type         string `json:"$type,omitempty"`
}

// WidgetView يمثل أداة (Widget) متاحة للإضافة إلى لوحات التحكم أو المقالات في YouTrack
type WidgetView struct {
	ID              string  `json:"id"`
	Key             string  `json:"key"`
	Name            string  `json:"name"`
	Description     string  `json:"description,omitempty"`
	AppID           string  `json:"appId"`
	AppName         string  `json:"appName"`
	AppTitle        string  `json:"appTitle,omitempty"`
	ExtensionPoint  string  `json:"extensionPoint"`
	IndexPath       string  `json:"indexPath,omitempty"`
	IconPath        *string `json:"iconPath,omitempty"`
	AppIconPath     *string `json:"appIconPath,omitempty"`
	AppDarkIconPath *string `json:"appDarkIconPath,omitempty"`
	Configurable    bool    `json:"configurable"`
	Collapsed       bool    `json:"collapsed"`
	ShowHeader      bool    `json:"showHeader"`
	Borderless      bool    `json:"borderless"`
	DefaultHeight   *string `json:"defaultHeight,omitempty"`
	DefaultWidth    *string `json:"defaultWidth,omitempty"`
	ExpectedHeight  *int    `json:"expectedHeight,omitempty"`
	ExpectedWidth   *int    `json:"expectedWidth,omitempty"`
	Guard           *string `json:"guard,omitempty"`
	VendorName      *string `json:"vendorName,omitempty"`
	VendorEmail     *string `json:"vendorEmail,omitempty"`
	VendorURL       *string `json:"vendorUrl,omitempty"`
	MarketplaceID   *int    `json:"marketplaceId,omitempty"`
	Type            string  `json:"$type,omitempty"`
}

// HelpdeskUserProfile إعدادات وصلاحيات مكتب المساعدة (Helpdesk)
type HelpdeskUserProfile struct {
	IsAgent                bool       `json:"isAgent"`
	IsReporter             bool       `json:"isReporter"`
	AgentInProjects        []*Project `json:"agentInProjects"`
	ReporterInProjects     []*Project `json:"reporterInProjects"`
	TableViewTicketColumns []string   `json:"tableViewTicketColumns"`
	Type                   string     `json:"$type,omitempty"`
}
