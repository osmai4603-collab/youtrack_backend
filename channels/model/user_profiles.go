package model

// FeatureFlag يمثّل ميزة تجريبية أو تفعيلية في YouTrack.
type FeatureFlag struct {
	ID      string `json:"id" db:"id"`
	Enabled bool   `json:"enabled" db:"enabled"`
	Type    string `json:"$type"`
}

// DashboardWidget يمثّل ودجة في لوحة القيادة أو تفاصيل المستخدم.
type DashboardWidget struct {
	ID              string  `json:"id" db:"id"`
	Key             string  `json:"key,omitempty" db:"key"`
	AppID           string  `json:"appId,omitempty" db:"app_id"`
	Description     string  `json:"description,omitempty" db:"description"`
	AppName         string  `json:"appName,omitempty" db:"app_name"`
	AppTitle        string  `json:"appTitle,omitempty" db:"app_title"`
	Name            string  `json:"name,omitempty" db:"name"`
	Collapsed       bool    `json:"collapsed,omitempty" db:"collapsed"`
	Configurable    bool    `json:"configurable,omitempty" db:"configurable"`
	IndexPath       string  `json:"indexPath,omitempty" db:"index_path"`
	ExtensionPoint  string  `json:"extensionPoint,omitempty" db:"extension_point"`
	IconPath        string  `json:"iconPath,omitempty" db:"icon_path"`
	Guard           string  `json:"guard,omitempty" db:"guard"`
	AppIconPath     string  `json:"appIconPath,omitempty" db:"app_icon_path"`
	AppDarkIconPath string  `json:"appDarkIconPath,omitempty" db:"app_dark_icon_path"`
	DefaultHeight   *string `json:"defaultHeight,omitempty" db:"default_height"`
	DefaultWidth    *string `json:"defaultWidth,omitempty" db:"default_width"`
	ExpectedHeight  *string `json:"expectedHeight,omitempty" db:"expected_height"`
	ExpectedWidth   *string `json:"expectedWidth,omitempty" db:"expected_width"`
	VendorName      string  `json:"vendorName,omitempty" db:"vendor_name"`
	VendorEmail     string  `json:"vendorEmail,omitempty" db:"vendor_email"`
	VendorURL       string  `json:"vendorUrl,omitempty" db:"vendor_url"`
	MarketplaceID   *int64  `json:"marketplaceId,omitempty" db:"marketplace_id"`
	ShowHeader      bool    `json:"showHeader,omitempty" db:"show_header"`
	Borderless      bool    `json:"borderless,omitempty" db:"borderless"`
	Type            string  `json:"$type,omitempty"`
}

// LocaleDescriptor يصف لغة واجهة المستخدم.
type LocaleDescriptor struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Community bool   `json:"community"`
	Locale    string `json:"locale"`
	Language  string `json:"language"`
	Type      string `json:"$type"`
}

// TimeZoneDescriptor يصف المنطقة الزمنية.
type TimeZoneDescriptor struct {
	ID   string `json:"id"`
	Type string `json:"$type"`
}

// DateFormatDescriptor يصف صيغة التاريخ والوقت.
type DateFormatDescriptor struct {
	Pattern     string `json:"pattern"`
	DatePattern string `json:"datePattern"`
	Type        string `json:"$type"`
}

// GeneralUserProfile يحوي الإعدادات العامة لملف المستخدم.
type GeneralUserProfile struct {
	ID                        string                `json:"id,omitempty"`
	Timezone                  *TimeZoneDescriptor   `json:"timezone"`
	DateFormat                *DateFormatDescriptor `json:"dateFieldFormat"`
	Locale                    *LocaleDescriptor     `json:"locale"`
	SemanticSearchForArticles bool                  `json:"semanticSearchForArticles"`
	LastCreatedIssue          any                   `json:"lastCreatedIssue"`
	SearchContext             any                   `json:"searchContext"`
	HelpdeskContext           any                   `json:"helpdeskContext"`
	Type                      string                `json:"$type"`
}

// ArticlesUserProfile يحوي إعدادات المقالات وقاعدة المعرفة.
type ArticlesUserProfile struct {
	LastVisitedArticle any    `json:"lastVisitedArticle"`
	ShowHistory        bool   `json:"showHistory"`
	ShowComments       bool   `json:"showComments"`
	QueryMode          any    `json:"queryMode"`
	ShowInlineComments bool   `json:"showInlineComments"`
	Type               string `json:"$type"`
}

// PeriodFieldFormat يصف صيغة المدة الزمنية في تتبع الوقت.
type PeriodFieldFormat struct {
	ID   string `json:"id"`
	Type string `json:"$type"`
}

// TimeTrackingUserProfile يحوي إعدادات تتبع الوقت للمستخدم.
type TimeTrackingUserProfile struct {
	PeriodFormat            *PeriodFieldFormat `json:"periodFormat"`
	OwnTimesheets           bool               `json:"ownTimesheets"`
	PeriodFieldPattern      string             `json:"periodFieldPattern"`
	IsTimeTrackingAvailable bool               `json:"isTimeTrackingAvailable"`
	Type                    string             `json:"$type"`
}

// TipsUserProfile يحوي كافة حالات التلميحات والأونبوردينغ للمستخدم.
type TipsUserProfile struct {
	IssueListPageCompleted              bool   `json:"issueListPageCompleted"`
	IssuePageCompleted                  bool   `json:"issuePageCompleted"`
	HelpdeskTeamTipShown                bool   `json:"helpdeskTeamTipShown"`
	ProjectOverviewTipShown             bool   `json:"projectOverviewTipShown"`
	ProjectSettingsTipShown             bool   `json:"projectSettingsTipShown"`
	IssueAiActionsTipShown              bool   `json:"issueAiActionsTipShown"`
	ProjectContentTipShown              bool   `json:"projectContentTipShown"`
	AgileBoardCardsTipShown             bool   `json:"agileBoardCardsTipShown"`
	AppsProjectTabTipShown              bool   `json:"appsProjectTabTipShown"`
	ChangeRuleTypeTipShown              bool   `json:"changeRuleTypeTipShown"`
	AgileBoardVisibilityTipShown        bool   `json:"agileBoardVisibilityTipShown"`
	AgileBoardSwimlanesTipShown         bool   `json:"agileBoardSwimlanesTipShown"`
	AgileBoardColumnsTipShown           bool   `json:"agileBoardColumnsTipShown"`
	ArticleAiAssistantTipShown          bool   `json:"articleAiAssistantTipShown"`
	ArticleVisibilityTipShown           bool   `json:"articleVisibilityTipShown"`
	ArticleInlineCommentsTipShown       bool   `json:"articleInlineCommentsTipShown"`
	ProjectSettingsPeopleTipShown       bool   `json:"projectSettingsPeopleTipShown"`
	ProjectSettingsFieldsTipShown       bool   `json:"projectSettingsFieldsTipShown"`
	ProjectSettingsVcsTipShown          bool   `json:"projectSettingsVcsTipShown"`
	HelpdeskPinnedCommentsTipShown      bool   `json:"helpdeskPinnedCommentsTipShown"`
	PricingAdminPopupShown              bool   `json:"pricingAdminPopupShown"`
	TextCompletionPromoShown            bool   `json:"textCompletionPromoShown"`
	TextRecognitionTipsShown            bool   `json:"textRecognitionTipsShown"`
	ArticleSidebarTipShown              bool   `json:"articleSidebarTipShown"`
	ArticleCommentsTipShown             bool   `json:"articleCommentsTipShown"`
	ActivityTypesTipShown               bool   `json:"activityTypesTipShown"`
	IssueFieldsTipShown                 bool   `json:"issueFieldsTipShown"`
	CommandsTipShown                    bool   `json:"commandsTipShown"`
	TimeTrackingTipShown                bool   `json:"timeTrackingTipShown"`
	VisibleFieldsTipShown               bool   `json:"visibleFieldsTipShown"`
	SurveyShown                         bool   `json:"surveyShown"`
	VotesTipShown                       bool   `json:"votesTipShown"`
	AiPromoShown                        bool   `json:"aiPromoShown"`
	CollapsibleSidebarTipShown          bool   `json:"collapsibleSidebarTipShown"`
	PmfShown                            bool   `json:"pmfShown"`
	AiWritingAssistantPromoShown        bool   `json:"aiWritingAssistantPromoShown"`
	InlineCommentPromoShown             bool   `json:"inlineCommentPromoShown"`
	ProjectSettingsTimeTrackingTipShown bool   `json:"projectSettingsTimeTrackingTipShown"`
	ProjectSettingsTeamcityTipShown     bool   `json:"projectSettingsTeamcityTipShown"`
	ProjectSettingsWorkflowTipShown     bool   `json:"projectSettingsWorkflowTipShown"`
	ProjectSettingsAppsTipShown         bool   `json:"projectSettingsAppsTipShown"`
	IssueTextRecognitionTipShown        bool   `json:"issueTextRecognitionTipShown"`
	HelpdeskChannelsTipShown            bool   `json:"helpdeskChannelsTipShown"`
	HelpdeskOverviewTipShown            bool   `json:"helpdeskOverviewTipShown"`
	ProjectOverviewPageCompleted        bool   `json:"projectOverviewPageCompleted"`
	ProjectSettingsPageCompleted        bool   `json:"projectSettingsPageCompleted"`
	DelayedDemoModalShown               bool   `json:"delayedDemoModalShown"`
	SavedSearchesTipShown               bool   `json:"savedSearchesTipShown"`
	SearchOptionsTipShown               bool   `json:"searchOptionsTipShown"`
	VisibilityRestrictionsTipShown      bool   `json:"visibilityRestrictionsTipShown"`
	AgileBoardBacklogTipShown           bool   `json:"agileBoardBacklogTipShown"`
	TextRecognitionPromoShown           bool   `json:"textRecognitionPromoShown"`
	AiTipsShown                         bool   `json:"aiTipsShown"`
	AgileBoardPageCompleted             bool   `json:"agileBoardPageCompleted"`
	ArticlesPageCompleted               bool   `json:"articlesPageCompleted"`
	HelpdeskSlaTipShown                 bool   `json:"helpdeskSlaTipShown"`
	OnboardingTourState                 string `json:"onboardingTourState"`
	HelpdeskProjectPageCompleted        bool   `json:"helpdeskProjectPageCompleted"`
	OnboardingTourAIBlockDismissed      bool   `json:"onboardingTourAIBlockDismissed"`
	Type                                string `json:"$type"`
}

// AppearanceUserProfile يحوي كافة إعدادات الواجهة والمظهر.
type AppearanceUserProfile struct {
	OpenCwOnTyping                      bool   `json:"openCwOnTyping"`
	SivSidebarWidth                     int    `json:"sivSidebarWidth"`
	ShowToolbar                         bool   `json:"showToolbar"`
	ShowQuickView                       bool   `json:"showQuickView"`
	ShowSimilarIssues                   bool   `json:"showSimilarIssues"`
	ShowKnowledgeBaseSidebar            bool   `json:"showKnowledgeBaseSidebar"`
	ShowSIVSidebar                      bool   `json:"showSIVSidebar"`
	DashboardHeaderCollapsed            bool   `json:"dashboardHeaderCollapsed"`
	KnowledgeBaseSubarticlesCollapsed   bool   `json:"knowledgeBaseSubarticlesCollapsed"`
	AiLinkSuggestionsCollapsed          bool   `json:"aiLinkSuggestionsCollapsed"`
	ShowCommentsInActivityStream        bool   `json:"showCommentsInActivityStream"`
	KnowledgeBaseSidebarWidth           int    `json:"knowledgeBaseSidebarWidth"`
	LastUsedColor                       string `json:"lastUsedColor"`
	AttachmentsListLayout               bool   `json:"attachmentsListLayout"`
	ShowTooltips                        bool   `json:"showTooltips"`
	ExpandNavigation                    bool   `json:"expandNavigation"`
	HideCommentAttachments              bool   `json:"hideCommentAttachments"`
	SidebarQuickViewMode                bool   `json:"sidebarQuickViewMode"`
	ExpandChangesInActivityStream       bool   `json:"expandChangesInActivityStream"`
	NaturalCommentsOrder                bool   `json:"naturalCommentsOrder"`
	UseAbsoluteDates                    bool   `json:"useAbsoluteDates"`
	UseMarkdownEditor                   bool   `json:"useMarkdownEditor"`
	ShowInlineEditorToolbar             bool   `json:"showInlineEditorToolbar"`
	ShowVcsChangesInActivityStream      bool   `json:"showVcsChangesInActivityStream"`
	TableViewColumns                    []any  `json:"tableViewColumns"`
	IssueListSidebarWidth               int    `json:"issueListSidebarWidth"`
	AttachmentsCollapsed                bool   `json:"attachmentsCollapsed"`
	UseSummaryInIssueLinks              bool   `json:"useSummaryInIssueLinks"`
	ShowRecentEntities                  bool   `json:"showRecentEntities"`
	ExceptionsExpanded                  bool   `json:"exceptionsExpanded"`
	OnboardingTourPanelWidth            int    `json:"onboardingTourPanelWidth"`
	ShowLinksUnderDescription           bool   `json:"showLinksUnderDescription"`
	RecognizedTextSidebarExpanded       bool   `json:"recognizedTextSidebarExpanded"`
	QuickViewSidebarWidth               int    `json:"quickViewSidebarWidth"`
	ModalSidebarWidth                   int    `json:"modalSidebarWidth"`
	ShowSidebarResizerTip               bool   `json:"showSidebarResizerTip"`
	IssuesTableViewMode                 bool   `json:"issuesTableViewMode"`
	AttachmentsSorting                  string `json:"attachmentsSorting"`
	HideEmbeddedAttachments             bool   `json:"hideEmbeddedAttachments"`
	CompactMode                         bool   `json:"compactMode"`
	QuickViewWidth                      int    `json:"quickViewWidth"`
	ShowHistoryInActivityStream         bool   `json:"showHistoryInActivityStream"`
	ShowWorkItemsInActivityStream       bool   `json:"showWorkItemsInActivityStream"`
	FirstDayOfWeek                      int    `json:"firstDayOfWeek"`
	Type                                string `json:"$type"`
}

// IssueListView يصف خيارات عرض قائمة المشاكل.
type IssueListView struct {
	DetailLevel       int    `json:"detailLevel"`
	TreeView          bool   `json:"treeView"`
	TreeViewCollapsed bool   `json:"treeViewCollapsed"`
	Type              string `json:"$type"`
}

// IssuesListUserProfile يحوي إعدادات قائمة المشاكل.
type IssuesListUserProfile struct {
	SavedSearchesExpanded bool           `json:"savedSearchesExpanded"`
	UnresolvedIssuesOnly  bool           `json:"unresolvedIssuesOnly"`
	ShowSidebar           bool           `json:"showSidebar"`
	ProjectsExpanded      bool           `json:"projectsExpanded"`
	SortTextByRelevance   bool           `json:"sortTextByRelevance"`
	TagsExpanded          bool           `json:"tagsExpanded"`
	QueryMode             bool           `json:"queryMode"`
	IssueListView         *IssueListView `json:"issueListView"`
	Type                  string         `json:"$type"`
}

// HelpdeskUserProfile يحوي إعدادات وتعيينات الدعم الفني للمستخدم.
type HelpdeskUserProfile struct {
	IsReporter             bool       `json:"isReporter"`
	TableViewTicketColumns []any      `json:"tableViewTicketColumns"`
	IsAgent                bool       `json:"isAgent"`
	AgentInProjects        []*Project `json:"agentInProjects"`
	ReporterInProjects     []*Project `json:"reporterInProjects"`
	Type                   string     `json:"$type"`
}

// AiUserProfile يحوي إعدادات مساعد الذكاء الاصطناعي ونافذة الدردشة.
type AiUserProfile struct {
	ChatsListShow       bool   `json:"chatsListShow"`
	ChatFloatingWidth   int    `json:"chatFloatingWidth"`
	ChatMode            string `json:"chatMode"`
	ChatSidebarShow     bool   `json:"chatSidebarShow"`
	ChatFloatingOffsetY int    `json:"chatFloatingOffsetY"`
	ChatFloatingOffsetX int    `json:"chatFloatingOffsetX"`
	ChatFloatingAnchor  string `json:"chatFloatingAnchor"`
	ChatFloatingHeight  int    `json:"chatFloatingHeight"`
	ChatSidebarWidth    int    `json:"chatSidebarWidth"`
	DisableChat         bool   `json:"disableChat"`
	Type                string `json:"$type"`
}

// NotificationsUserProfile يحوي إعدادات التنبيهات وإشعارات البريد.
type NotificationsUserProfile struct {
	ShowUnreadOnly                         bool    `json:"showUnreadOnly"`
	MentionNotificationsEnabled            bool    `json:"mentionNotificationsEnabled"`
	DuplicateClusterNotificationsEnabled   bool    `json:"duplicateClusterNotificationsEnabled"`
	ShowSystem                             bool    `json:"showSystem"`
	NotifyOnOwnChanges                     bool    `json:"notifyOnOwnChanges"`
	AutoWatchOnFieldSet                    bool    `json:"autoWatchOnFieldSet"`
	AutoWatchOnCreate                      bool    `json:"autoWatchOnCreate"`
	AutoWatchOnComment                     bool    `json:"autoWatchOnComment"`
	AutoWatchOnUpdate                      bool    `json:"autoWatchOnUpdate"`
	AutoWatchOnVote                        bool    `json:"autoWatchOnVote"`
	EmailBlocked                           bool    `json:"emailBlocked"`
	EmailBlockReason                       *string `json:"emailBlockReason"`
	EmailNotificationsEnabled              bool    `json:"emailNotificationsEnabled"`
	UsePlainTextEmails                     bool    `json:"usePlainTextEmails"`
	DisabledDirect                         bool    `json:"disabledDirect"`
	DisabledSubscription                   bool    `json:"disabledSubscription"`
	DisabledSystem                         bool    `json:"disabledSystem"`
	MailboxIntegrationNotificationsEnabled bool    `json:"mailboxIntegrationNotificationsEnabled"`
	Type                                   string  `json:"$type"`
}

// GrazieUserProfile يصف إعدادات الذكاء الاصطناعي والتدقيق اللغوي Grazie.
type GrazieUserProfile struct {
	ExcludedIssueTypes            string `json:"excludedIssueTypes"`
	EnableSpellChecker            bool   `json:"enableSpellChecker"`
	HasMoreTokens                 bool   `json:"hasMoreTokens"`
	EnableTextCompletion          bool   `json:"enableTextCompletion"`
	CycleRestart                  int64  `json:"cycleRestart"`
	FreeLicense                   bool   `json:"freeLicense"`
	SpellCheckerEnabledInSystem   bool   `json:"spellCheckerEnabledInSystem"`
	TextCompletionEnabledInSystem bool   `json:"textCompletionEnabledInSystem"`
	Enabled                       bool   `json:"enabled"`
	Type                          string `json:"$type"`
}

// QuestionnaireUserProfile يصف حالات الاستبيانات واستطلاعات الرأي.
type QuestionnaireUserProfile struct {
	ShowSurvey               bool   `json:"showSurvey"`
	ShowPmfSurvey            bool   `json:"showPmfSurvey"`
	DemoEligibilityTimestamp *int64 `json:"demoEligibilityTimestamp"`
	Type                     string `json:"$type"`
}

// RecentIssue يمثّل قضية/مشكلة تم زيارتها مؤخراً.
type RecentIssue struct {
	ID     string `json:"id" db:"id"`
	Pinned bool   `json:"pinned" db:"pinned"`
	Date   int64  `json:"date" db:"date"`
	Issue  *Issue `json:"issue,omitempty"`
	Type   string `json:"$type"`
}

// RecentArticle يمثّل مقالاً تم زيارته مؤخراً.
type RecentArticle struct {
	ID      string `json:"id" db:"id"`
	Pinned  bool   `json:"pinned" db:"pinned"`
	Date    int64  `json:"date" db:"date"`
	Article any    `json:"article,omitempty"`
	Type    string `json:"$type"`
}

// InboxFolder يمثّل مجلدًا في صندوق الوارد (مطابق لـ request8.txt).
type InboxFolder struct {
	ID           string `json:"id" db:"id"`
	LastNotified int64  `json:"lastNotified" db:"last_notified"`
	LastSeen     int64  `json:"lastSeen" db:"last_seen"`
	Enabled      bool   `json:"enabled" db:"enabled"`
	Type         string `json:"$type"`
}

// HubAvatar يصف صورة المستخدم في Hub.
type HubAvatar struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

// HubProfile يصف الملف التعريفي في Hub.
type HubProfile struct {
	Avatar *HubAvatar `json:"avatar,omitempty"`
	Email  string     `json:"email,omitempty"`
}

// Hub2FA يصف المصادقة الثنائية.
type Hub2FA struct {
	Enabled bool `json:"enabled"`
}

// HubWebauthn يصف جهاز WebAuthn.
type HubWebauthn struct {
	Enabled bool `json:"enabled"`
}

// HubUser يمثّل ملف المستخدم في خدمة Hub.
type HubUser struct {
	Guest                           bool         `json:"guest"`
	ID                              string       `json:"id"`
	Name                            string       `json:"name"`
	Login                           string       `json:"login"`
	Profile                         *HubProfile  `json:"profile,omitempty"`
	RequiredTwoFactorAuthentication bool         `json:"requiredTwoFactorAuthentication"`
	TwoFactorAuthentication         *Hub2FA      `json:"twoFactorAuthentication,omitempty"`
	WebauthnDevice                  *HubWebauthn `json:"webauthnDevice,omitempty"`
	Type                            string       `json:"$type,omitempty"`
}

// UserProfiles يجمّع الـ 9 ملفات الشخصية المتداخلة لكائن Me.
type UserProfiles struct {
	AI            *AiUserProfile            `json:"ai,omitempty"`
	Tips          *TipsUserProfile          `json:"tips,omitempty"`
	Helpdesk      *HelpdeskUserProfile      `json:"helpdesk,omitempty"`
	Timetracking  *TimeTrackingUserProfile  `json:"timetracking,omitempty"`
	General       *GeneralUserProfile       `json:"general,omitempty"`
	Appearance    *AppearanceUserProfile    `json:"appearance,omitempty"`
	IssuesList    *IssuesListUserProfile    `json:"issuesList,omitempty"`
	Articles      *ArticlesUserProfile      `json:"articles,omitempty"`
	Notifications *NotificationsUserProfile `json:"notifications,omitempty"`
	Type          string                    `json:"$type,omitempty"`
}

// DefaultUserProfiles يبني كائن UserProfiles كامل بالقيم الافتراضية الرسمية لـ YouTrack.
func DefaultUserProfiles() *UserProfiles {
	return &UserProfiles{
		Type: "UserProfiles",
		AI: &AiUserProfile{
			ChatsListShow:       true,
			ChatFloatingWidth:   400,
			ChatMode:            "docked",
			ChatSidebarShow:     false,
			ChatFloatingOffsetY: 100,
			ChatFloatingOffsetX: 100,
			ChatFloatingAnchor:  "top-right",
			ChatFloatingHeight:  600,
			ChatSidebarWidth:    400,
			DisableChat:         false,
			Type:                "AiUserProfile",
		},
		Tips: &TipsUserProfile{
			IssueListPageCompleted:              false,
			IssuePageCompleted:                  false,
			HelpdeskTeamTipShown:                false,
			ProjectOverviewTipShown:             false,
			ProjectSettingsTipShown:             false,
			IssueAiActionsTipShown:              false,
			ProjectContentTipShown:              false,
			AgileBoardCardsTipShown:             false,
			AppsProjectTabTipShown:              false,
			ChangeRuleTypeTipShown:              false,
			AgileBoardVisibilityTipShown:        false,
			AgileBoardSwimlanesTipShown:         false,
			AgileBoardColumnsTipShown:           false,
			ArticleAiAssistantTipShown:          false,
			ArticleVisibilityTipShown:           false,
			ArticleInlineCommentsTipShown:       false,
			ProjectSettingsPeopleTipShown:       false,
			ProjectSettingsFieldsTipShown:       false,
			ProjectSettingsVcsTipShown:          false,
			HelpdeskPinnedCommentsTipShown:      false,
			PricingAdminPopupShown:              false,
			TextCompletionPromoShown:            false,
			TextRecognitionTipsShown:            false,
			ArticleSidebarTipShown:              false,
			ArticleCommentsTipShown:             false,
			ActivityTypesTipShown:               false,
			IssueFieldsTipShown:                 false,
			CommandsTipShown:                    false,
			TimeTrackingTipShown:                false,
			VisibleFieldsTipShown:               false,
			SurveyShown:                         false,
			VotesTipShown:                       false,
			AiPromoShown:                        false,
			CollapsibleSidebarTipShown:          false,
			PmfShown:                            false,
			AiWritingAssistantPromoShown:        false,
			InlineCommentPromoShown:             true,
			ProjectSettingsTimeTrackingTipShown: false,
			ProjectSettingsTeamcityTipShown:     false,
			ProjectSettingsWorkflowTipShown:     false,
			ProjectSettingsAppsTipShown:         false,
			IssueTextRecognitionTipShown:        false,
			HelpdeskChannelsTipShown:            false,
			HelpdeskOverviewTipShown:            false,
			ProjectOverviewPageCompleted:        false,
			ProjectSettingsPageCompleted:        false,
			DelayedDemoModalShown:               false,
			SavedSearchesTipShown:               false,
			SearchOptionsTipShown:               false,
			VisibilityRestrictionsTipShown:      false,
			AgileBoardBacklogTipShown:           false,
			TextRecognitionPromoShown:           false,
			AiTipsShown:                         false,
			AgileBoardPageCompleted:             false,
			ArticlesPageCompleted:               false,
			HelpdeskSlaTipShown:                 false,
			OnboardingTourState:                 "idle",
			HelpdeskProjectPageCompleted:        false,
			OnboardingTourAIBlockDismissed:      false,
			Type:                                "TipsUserProfile",
		},
		Helpdesk: &HelpdeskUserProfile{
			IsReporter:             false,
			TableViewTicketColumns: []any{},
			IsAgent:                false,
			AgentInProjects:        []*Project{},
			ReporterInProjects:     []*Project{},
			Type:                   "HelpdeskUserProfile",
		},
		Timetracking: &TimeTrackingUserProfile{
			PeriodFormat: &PeriodFieldFormat{
				ID:   "FULL",
				Type: "PeriodFieldFormat",
			},
			OwnTimesheets:           false,
			PeriodFieldPattern:      `((\d+w\s*)?(\s*)?(\d+d\s*)?(\s*)?(\d+h\s*)?(\s*)?(\d+m\s*)?|\d+)`,
			IsTimeTrackingAvailable: false,
			Type:                    "TimeTrackingUserProfile",
		},
		General: &GeneralUserProfile{
			Timezone: &TimeZoneDescriptor{
				ID:   "Europe/Prague",
				Type: "TimeZoneDescriptor",
			},
			DateFormat: &DateFormatDescriptor{
				Pattern:     "d MMM yyyy HH:mm",
				DatePattern: "d MMM yyyy",
				Type:        "DateFormatDescriptor",
			},
			Locale: &LocaleDescriptor{
				Name:      "English",
				Community: false,
				Locale:    "en_US",
				ID:        "en_US",
				Language:  "en",
				Type:      "LocaleDescriptor",
			},
			SemanticSearchForArticles: false,
			LastCreatedIssue:          nil,
			SearchContext:             nil,
			HelpdeskContext:           nil,
			Type:                      "GeneralUserProfile",
		},
		Appearance: &AppearanceUserProfile{
			OpenCwOnTyping:                    true,
			SivSidebarWidth:                   240,
			ShowToolbar:                       true,
			ShowQuickView:                     true,
			ShowSimilarIssues:                 true,
			ShowKnowledgeBaseSidebar:          true,
			ShowSIVSidebar:                    false,
			DashboardHeaderCollapsed:          false,
			KnowledgeBaseSubarticlesCollapsed: false,
			AiLinkSuggestionsCollapsed:        false,
			ShowCommentsInActivityStream:      true,
			KnowledgeBaseSidebarWidth:         240,
			LastUsedColor:                     "default",
			AttachmentsListLayout:             false,
			ShowTooltips:                      true,
			ExpandNavigation:                  true,
			HideCommentAttachments:            true,
			SidebarQuickViewMode:              true,
			ExpandChangesInActivityStream:     false,
			NaturalCommentsOrder:              true,
			UseAbsoluteDates:                  false,
			UseMarkdownEditor:                 false,
			ShowInlineEditorToolbar:           true,
			ShowVcsChangesInActivityStream:    false,
			TableViewColumns:                  []any{},
			IssueListSidebarWidth:             240,
			AttachmentsCollapsed:              false,
			UseSummaryInIssueLinks:            true,
			ShowRecentEntities:                true,
			ExceptionsExpanded:                true,
			OnboardingTourPanelWidth:          400,
			ShowLinksUnderDescription:         true,
			RecognizedTextSidebarExpanded:     false,
			QuickViewSidebarWidth:             240,
			ModalSidebarWidth:                 240,
			ShowSidebarResizerTip:             true,
			IssuesTableViewMode:               true,
			AttachmentsSorting:                "default",
			HideEmbeddedAttachments:           true,
			CompactMode:                       false,
			QuickViewWidth:                    0,
			ShowHistoryInActivityStream:       false,
			ShowWorkItemsInActivityStream:     false,
			FirstDayOfWeek:                    0,
			Type:                              "AppearanceUserProfile",
		},
		IssuesList: &IssuesListUserProfile{
			SavedSearchesExpanded: true,
			UnresolvedIssuesOnly:  false,
			ShowSidebar:           true,
			ProjectsExpanded:      true,
			SortTextByRelevance:   true,
			TagsExpanded:          true,
			QueryMode:             false,
			IssueListView: &IssueListView{
				DetailLevel:       20,
				TreeView:          true,
				TreeViewCollapsed: false,
				Type:              "IssueListView",
			},
			Type: "IssuesListUserProfile",
		},
		Articles: &ArticlesUserProfile{
			LastVisitedArticle: nil,
			ShowHistory:        false,
			ShowComments:       true,
			QueryMode:          false,
			ShowInlineComments: false,
			Type:               "ArticlesUserProfile",
		},
		Notifications: &NotificationsUserProfile{
			ShowUnreadOnly:                         false,
			MentionNotificationsEnabled:            true,
			DuplicateClusterNotificationsEnabled:   false,
			ShowSystem:                             false,
			NotifyOnOwnChanges:                     false,
			AutoWatchOnFieldSet:                    true,
			AutoWatchOnCreate:                      true,
			AutoWatchOnComment:                     true,
			AutoWatchOnUpdate:                      false,
			AutoWatchOnVote:                        true,
			EmailBlocked:                           false,
			EmailBlockReason:                       nil,
			EmailNotificationsEnabled:              true,
			UsePlainTextEmails:                     false,
			DisabledDirect:                         false,
			DisabledSubscription:                   false,
			DisabledSystem:                         false,
			MailboxIntegrationNotificationsEnabled: true,
			Type:                                   "NotificationsUserProfile",
		},
	}
}
