package domain

// App يمثل التطبيق أو Workflow في YouTrack
type App struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Title                 string            `json:"title,omitempty"`
	Version               string            `json:"version,omitempty"`
	Language              *PackageLanguage  `json:"language,omitempty"`
	Icon                  string            `json:"icon,omitempty"`
	DarkIcon              string            `json:"darkIcon,omitempty"`
	AutoAttach            bool              `json:"autoAttach,omitempty"`
	Permanent             bool              `json:"permanent,omitempty"`
	FromMarketplace       bool              `json:"fromMarketplace,omitempty"`
	HasWidgetOrHttp       bool              `json:"hasWidgetOrHttp,omitempty"`
	CanBeAttached         bool              `json:"canBeAttached,omitempty"`
	CanUpdateContent      bool              `json:"canUpdateContent,omitempty"`
	CanDeleteContent      bool              `json:"canDeleteContent,omitempty"`
	HasBrokenUsages       bool              `json:"hasBrokenUsages,omitempty"`
	Model                 string            `json:"model,omitempty"`
	Updated               int64             `json:"updated,omitempty"`
	UpdatedBy             *User             `json:"updatedBy,omitempty"`
	MarketplaceID         *int              `json:"marketplaceId,omitempty"`
	Vendor                *AppVendor        `json:"vendor,omitempty"`
	GlobalConfig          *AppConfiguration `json:"globalConfig,omitempty"`
	Widgets               []*Widget         `json:"widgets,omitempty"`
	Usages                []*ProjectAppConf `json:"usages,omitempty"`
	Tags                  []*AppTag         `json:"tags,omitempty"`
	Type                  string            `json:"$type,omitempty"`
}

// PackageLanguage لغة برمجة التطبيق
type PackageLanguage struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// AppVendor الشركة الموفرة للتطبيق
type AppVendor struct {
	Name  string `json:"name"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
	Type  string `json:"$type,omitempty"`
}

// AppConfiguration إعدادات التطبيق العامة
type AppConfiguration struct {
	Enabled                 bool `json:"enabled"`
	MissingRequiredSettings bool `json:"missingRequiredSettings"`
	Type                    string `json:"$type,omitempty"`
}

// AppTag وسم التطبيق
type AppTag struct {
	Name string `json:"name"`
	Type string `json:"$type,omitempty"`
}

// ProjectAppConf إعدادات التطبيق الخاصة بمشروع معين
type ProjectAppConf struct {
	ID       string   `json:"id"`
	Project  *Project `json:"project,omitempty"`
	CanUpdate bool    `json:"canUpdate,omitempty"`
	Type     string   `json:"$type,omitempty"`
}

// ProjectAppConfiguration تكوين التطبيق المفعل للمشروع
type ProjectAppConfiguration struct {
	ID                      string `json:"id"`
	Project                 *Project `json:"project,omitempty"`
	App                     *App     `json:"app,omitempty"`
	Enabled                 bool     `json:"enabled"`
	IsBroken                bool     `json:"isBroken"`
	CanUpdate               bool     `json:"canUpdate"`
	MissingRequiredSettings bool     `json:"missingRequiredSettings"`
	ProjectSettings         string   `json:"projectSettings,omitempty"` // JSON string
	Type                    string   `json:"$type,omitempty"`
}

// Widget يمثل الأداة البرمجية في لوحات التحكم
type Widget struct {
	ID               string `json:"id"`
	Name             string `json:"name,omitempty"`
	Key              string `json:"key,omitempty"`
	AppID            string `json:"appId,omitempty"`
	AppName          string `json:"appName,omitempty"`
	AppTitle         string `json:"appTitle,omitempty"`
	Description      string `json:"description,omitempty"`
	ExtensionPoint   string `json:"extensionPoint,omitempty"`
	IconPath         string `json:"iconPath,omitempty"`
	IndexPath        string `json:"indexPath,omitempty"`
	Configurable     bool   `json:"configurable,omitempty"`
	Collapsed        bool   `json:"collapsed,omitempty"`
	Borderless       bool   `json:"borderless,omitempty"`
	ShowHeader       bool   `json:"showHeader,omitempty"`
	DefaultHeight    string `json:"defaultHeight,omitempty"`
	DefaultWidth     string `json:"defaultWidth,omitempty"`
	ExpectedHeight   int    `json:"expectedHeight,omitempty"`
	ExpectedWidth    int    `json:"expectedWidth,omitempty"`
	VendorName       string `json:"vendorName,omitempty"`
	VendorEmail      string `json:"vendorEmail,omitempty"`
	VendorURL        string `json:"vendorUrl,omitempty"`
	MarketplaceID    int    `json:"marketplaceId,omitempty"`
	AppIconPath      string `json:"appIconPath,omitempty"`
	AppDarkIconPath  string `json:"appDarkIconPath,omitempty"`
	Guard            string `json:"guard,omitempty"`
	Type             string `json:"$type,omitempty"`
}

// ProjectDashboardWidget أداة تحكم داخل لوحة المشروع وموقعها الجغرافي
type ProjectDashboardWidget struct {
	ID        string   `json:"id"`
	Project   *Project `json:"project,omitempty"`
	Widget    *Widget  `json:"widget,omitempty"`
	Key       string   `json:"key,omitempty"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Settings  string   `json:"settings,omitempty"` // JSON string
	Type      string   `json:"$type,omitempty"`
}
