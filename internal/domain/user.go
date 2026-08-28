package domain

import "time"

// UserType نوع حساب المستخدم
type UserType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"$type,omitempty"`
}

// User يمثل نموذج المستخدم في النظام
type User struct {
	ID                string         `json:"id"`
	Login             string         `json:"login,omitempty"`
	Email             string         `json:"email,omitempty"`
	Name              string         `json:"name,omitempty"`
	FullName          string         `json:"fullName,omitempty"`
	LocalizedName     string         `json:"localizedName,omitempty"`
	AvatarURL         string         `json:"avatarUrl,omitempty"`
	Online            bool           `json:"online,omitempty"`
	Banned            bool           `json:"banned,omitempty"`
	BanBadge          *string        `json:"banBadge,omitempty"`
	CanReadProfile    bool           `json:"canReadProfile,omitempty"`
	IsLocked          bool           `json:"isLocked,omitempty"`
	IsEmailVerified   bool           `json:"isEmailVerified,omitempty"`
	Guest             bool           `json:"guest,omitempty"`
	UserType          *UserType      `json:"userType,omitempty"`
	IssueRelatedGroup *UserGroup     `json:"issueRelatedGroup,omitempty"`
	Profiles          *UserProfiles  `json:"profiles,omitempty"`
	FeatureFlags      []*FeatureFlag `json:"featureFlags,omitempty"`
	Widgets           []*Widget      `json:"widgets,omitempty"`
	Type              string         `json:"$type,omitempty"`
	CreatedAt         time.Time      `json:"created_at,omitempty"`
	UpdatedAt         time.Time      `json:"updated_at,omitempty"`
}

// UserGroup يمثل مجموعة المستخدمين أو الفريق
type UserGroup struct {
	ID             string       `json:"id"`
	Name           string       `json:"name,omitempty"`
	Description    string       `json:"description,omitempty"`
	AuditTargetID  string       `json:"auditTargetId,omitempty"`
	AllUsersGroup  bool         `json:"allUsersGroup,omitempty"`
	Icon           string       `json:"icon,omitempty"`
	TeamForProject *ProjectTeam `json:"teamForProject,omitempty"`
	IsUpdatable    bool         `json:"isUpdatable,omitempty"`
	IsRemovable    bool         `json:"isRemovable,omitempty"`
	Type           string       `json:"$type,omitempty"`
}

// FeatureFlag يمثل ميزة تجريبية مفعلة في النظام
type FeatureFlag struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"$type,omitempty"`
}

// UserProfiles يجمع كافة ملفات التفضيلات الخاصة بالمستخدم
type UserProfiles struct {
	General       *GeneralUserProfile       `json:"general,omitempty"`
	Appearance    *AppearanceUserProfile    `json:"appearance,omitempty"`
	IssuesList    *IssuesListUserProfile    `json:"issuesList,omitempty"`
	Articles      *ArticlesUserProfile      `json:"articles,omitempty"`
	Notifications *NotificationsUserProfile `json:"notifications,omitempty"`
	TimeTracking  *TimeTrackingUserProfile  `json:"timetracking,omitempty"`
	Tips          *TipsUserProfile          `json:"tips,omitempty"`
	Helpdesk      *HelpdeskUserProfile      `json:"helpdesk,omitempty"`
	AI            *AiUserProfile            `json:"ai,omitempty"`
	Grazie        *GrazieUserProfile        `json:"grazie,omitempty"`
	Type          string                    `json:"$type,omitempty"`
}

// GrazieUserProfile إعدادات التدقيق اللغوي والذكاء الاصطناعي (JetBrains Grazie) للمستخدم
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
	Type                          string `json:"$type,omitempty"`
}

// AiUserProfile تفضيلات وإعدادات مساعد الذكاء الاصطناعي
type AiUserProfile struct {
	DisableChat         bool   `json:"disableChat"`
	ChatSidebarShow     bool   `json:"chatSidebarShow"`
	ChatsListShow       bool   `json:"chatsListShow"`
	ChatMode            string `json:"chatMode,omitempty"` // e.g. "docked", "floating"
	ChatSidebarWidth    int    `json:"chatSidebarWidth,omitempty"`
	ChatFloatingAnchor  string `json:"chatFloatingAnchor,omitempty"`
	ChatFloatingOffsetX int    `json:"chatFloatingOffsetX,omitempty"`
	ChatFloatingOffsetY int    `json:"chatFloatingOffsetY,omitempty"`
	ChatFloatingWidth   int    `json:"chatFloatingWidth,omitempty"`
	ChatFloatingHeight  int    `json:"chatFloatingHeight,omitempty"`
	Type                string `json:"$type,omitempty"`
}

// GeneralUserProfile الإعدادات العامة للمستخدم (المنطقة الزمنية، اللغة، صيغة التاريخ)
type GeneralUserProfile struct {
	Timezone                  *TimeZoneDescriptor   `json:"timezone,omitempty"`
	DateFieldFormat           *DateFormatDescriptor `json:"dateFieldFormat,omitempty"`
	Locale                    *LocaleDescriptor     `json:"locale,omitempty"`
	SemanticSearchForArticles bool             `json:"semanticSearchForArticles"`
	LastCreatedIssue          *Issue           `json:"lastCreatedIssue,omitempty"`
	SearchContext             *HelpdeskContext `json:"searchContext,omitempty"`
	HelpdeskContext           *HelpdeskContext `json:"helpdeskContext,omitempty"`
	Type                      string           `json:"$type,omitempty"`
}

// TimeZoneDescriptor واصف المنطقة الزمنية
type TimeZoneDescriptor struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// DateFormatDescriptor واصف صيغة التاريخ
type DateFormatDescriptor struct {
	Pattern     string `json:"pattern,omitempty"`
	DatePattern string `json:"datePattern,omitempty"`
	Type        string `json:"$type,omitempty"`
}

// LocaleDescriptor واصف اللغة والموقع الجغرافي
type LocaleDescriptor struct {
	ID        string `json:"id"`
	Language  string `json:"language,omitempty"`
	Locale    string `json:"locale,omitempty"`
	Name      string `json:"name,omitempty"`
	Community bool   `json:"community,omitempty"`
	Type      string `json:"$type,omitempty"`
}

// AppearanceUserProfile إعدادات المظهر وواجهة المستخدم
type AppearanceUserProfile struct {
	FirstDayOfWeek                    int      `json:"firstDayOfWeek"`
	ShowTooltips                      bool     `json:"showTooltips"`
	ShowRecentEntities                bool     `json:"showRecentEntities"`
	ShowSidebarResizerTip             bool     `json:"showSidebarResizerTip"`
	ShowToolbar                       bool     `json:"showToolbar"`
	ShowInlineEditorToolbar           bool     `json:"showInlineEditorToolbar"`
	UseMarkdownEditor                 bool     `json:"useMarkdownEditor"`
	UseSummaryInIssueLinks            bool     `json:"useSummaryInIssueLinks"`
	ShowQuickView                     bool     `json:"showQuickView"`
	LastUsedColor                     string   `json:"lastUsedColor,omitempty"`
	TableViewColumns                  []string `json:"tableViewColumns"`
	NaturalCommentsOrder              bool     `json:"naturalCommentsOrder"`
	ShowCommentsInActivityStream      bool     `json:"showCommentsInActivityStream"`
	ShowVcsChangesInActivityStream    bool     `json:"showVcsChangesInActivityStream"`
	ShowWorkItemsInActivityStream     bool     `json:"showWorkItemsInActivityStream"`
	ShowHistoryInActivityStream       bool     `json:"showHistoryInActivityStream"`
	ExpandChangesInActivityStream     bool     `json:"expandChangesInActivityStream"`
	UseAbsoluteDates                  bool     `json:"useAbsoluteDates"`
	ShowSimilarIssues                 bool     `json:"showSimilarIssues"`
	ExceptionsExpanded                bool     `json:"exceptionsExpanded"`
	KnowledgeBaseSidebarWidth         int      `json:"knowledgeBaseSidebarWidth,omitempty"`
	SivSidebarWidth                   int      `json:"sivSidebarWidth,omitempty"`
	QuickViewSidebarWidth             int      `json:"quickViewSidebarWidth,omitempty"`
	ModalSidebarWidth                 int      `json:"modalSidebarWidth,omitempty"`
	IssueListSidebarWidth             int      `json:"issueListSidebarWidth,omitempty"`
	QuickViewWidth                    int      `json:"quickViewWidth,omitempty"`
	ShowKnowledgeBaseSidebar          bool     `json:"showKnowledgeBaseSidebar"`
	ShowSIVSidebar                    bool     `json:"showSIVSidebar"`
	AttachmentsCollapsed              bool     `json:"attachmentsCollapsed"`
	ShowLinksUnderDescription         bool     `json:"showLinksUnderDescription"`
	OpenCwOnTyping                    bool     `json:"openCwOnTyping"`
	CompactMode                       bool     `json:"compactMode"`
	RecognizedTextSidebarExpanded     bool     `json:"recognizedTextSidebarExpanded"`
	AttachmentsSorting                string   `json:"attachmentsSorting,omitempty"`
	HideEmbeddedAttachments           bool     `json:"hideEmbeddedAttachments"`
	HideCommentAttachments            bool     `json:"hideCommentAttachments"`
	AttachmentsListLayout             bool     `json:"attachmentsListLayout"`
	ExpandNavigation                  bool     `json:"expandNavigation"`
	IssuesTableViewMode               bool     `json:"issuesTableViewMode"`
	SidebarQuickViewMode              bool     `json:"sidebarQuickViewMode"`
	DashboardHeaderCollapsed          bool     `json:"dashboardHeaderCollapsed"`
	KnowledgeBaseSubarticlesCollapsed bool     `json:"knowledgeBaseSubarticlesCollapsed"`
	AiLinkSuggestionsCollapsed        bool     `json:"aiLinkSuggestionsCollapsed"`
	OnboardingTourPanelWidth          int      `json:"onboardingTourPanelWidth,omitempty"`
	Type                              string   `json:"$type,omitempty"`
}

// IssueListView طريقة عرض وتفاصيل قائمة المهام
type IssueListView struct {
	TreeView          bool   `json:"treeView"`
	TreeViewCollapsed bool   `json:"treeViewCollapsed"`
	DetailLevel       int    `json:"detailLevel"`
	Type              string `json:"$type,omitempty"`
}

// IssuesListUserProfile إعدادات قائمة وتصفية المهام
type IssuesListUserProfile struct {
	SortTextByRelevance   bool           `json:"sortTextByRelevance"`
	ShowSidebar           bool           `json:"showSidebar"`
	UnresolvedIssuesOnly  bool           `json:"unresolvedIssuesOnly"`
	QueryMode             bool           `json:"queryMode"`
	IssueListView         *IssueListView `json:"issueListView,omitempty"`
	ProjectsExpanded      bool           `json:"projectsExpanded"`
	SavedSearchesExpanded bool           `json:"savedSearchesExpanded"`
	TagsExpanded          bool           `json:"tagsExpanded"`
	Type                  string         `json:"$type,omitempty"`
}

// ArticlesUserProfile تفضيلات عرض المقالات وقاعدة المعرفة
type ArticlesUserProfile struct {
	ShowComments       bool     `json:"showComments"`
	ShowInlineComments bool     `json:"showInlineComments"`
	ShowHistory        bool     `json:"showHistory"`
	QueryMode          bool     `json:"queryMode"`
	LastVisitedArticle *Article `json:"lastVisitedArticle,omitempty"`
	Type               string   `json:"$type,omitempty"`
}

// NotificationsUserProfile إعدادات استقبال الإشعارات
type NotificationsUserProfile struct {
	EmailNotificationsEnabled              bool    `json:"emailNotificationsEnabled"`
	DisabledDirect                         bool    `json:"disabledDirect"`
	DisabledSubscription                   bool    `json:"disabledSubscription"`
	DisabledSystem                         bool    `json:"disabledSystem"`
	ShowSystem                             bool    `json:"showSystem"`
	ShowUnreadOnly                         bool    `json:"showUnreadOnly"`
	EmailBlocked                           bool    `json:"emailBlocked"`
	EmailBlockReason                       *string `json:"emailBlockReason,omitempty"`
	UsePlainTextEmails                     bool    `json:"usePlainTextEmails"`
	NotifyOnOwnChanges                     bool    `json:"notifyOnOwnChanges"`
	MentionNotificationsEnabled            bool    `json:"mentionNotificationsEnabled"`
	DuplicateClusterNotificationsEnabled   bool    `json:"duplicateClusterNotificationsEnabled"`
	MailboxIntegrationNotificationsEnabled bool    `json:"mailboxIntegrationNotificationsEnabled"`
	AutoWatchOnComment                     bool    `json:"autoWatchOnComment"`
	AutoWatchOnCreate                      bool    `json:"autoWatchOnCreate"`
	AutoWatchOnUpdate                      bool    `json:"autoWatchOnUpdate"`
	AutoWatchOnFieldSet                    bool    `json:"autoWatchOnFieldSet"`
	AutoWatchOnVote                        bool    `json:"autoWatchOnVote"`
	Type                                   string  `json:"$type,omitempty"`
}

// PeriodFieldFormat صيغة حقل الفترة الزمنية
type PeriodFieldFormat struct {
	ID   string `json:"id"`
	Type string `json:"$type,omitempty"`
}

// TimeTrackingUserProfile إعدادات تتبع الوقت
type TimeTrackingUserProfile struct {
	PeriodFormat            *PeriodFieldFormat `json:"periodFormat,omitempty"`
	PeriodFieldPattern      string             `json:"periodFieldPattern,omitempty"`
	IsTimeTrackingAvailable bool               `json:"isTimeTrackingAvailable"`
	OwnTimesheets           bool               `json:"ownTimesheets"`
	Type                    string             `json:"$type,omitempty"`
}

// TipsUserProfile تتبع التلميحات والنوافذ الإرشادية التي شاهدها المستخدم
type TipsUserProfile struct {
	IssueListPageCompleted              bool   `json:"issueListPageCompleted"`
	IssuePageCompleted                  bool   `json:"issuePageCompleted"`
	AgileBoardPageCompleted             bool   `json:"agileBoardPageCompleted"`
	ArticlesPageCompleted               bool   `json:"articlesPageCompleted"`
	ProjectOverviewPageCompleted        bool   `json:"projectOverviewPageCompleted"`
	ProjectSettingsPageCompleted        bool   `json:"projectSettingsPageCompleted"`
	HelpdeskProjectPageCompleted        bool   `json:"helpdeskProjectPageCompleted"`
	OnboardingTourState                 string `json:"onboardingTourState,omitempty"` // e.g. "paused", "completed"
	OnboardingTourAIBlockDismissed      bool   `json:"onboardingTourAIBlockDismissed"`
	HelpdeskTeamTipShown                bool   `json:"helpdeskTeamTipShown"`
	HelpdeskSlaTipShown                 bool   `json:"helpdeskSlaTipShown"`
	HelpdeskChannelsTipShown            bool   `json:"helpdeskChannelsTipShown"`
	HelpdeskOverviewTipShown            bool   `json:"helpdeskOverviewTipShown"`
	ProjectOverviewTipShown             bool   `json:"projectOverviewTipShown"`
	ProjectSettingsTipShown             bool   `json:"projectSettingsTipShown"`
	ProjectSettingsPeopleTipShown       bool   `json:"projectSettingsPeopleTipShown"`
	ProjectSettingsFieldsTipShown       bool   `json:"projectSettingsFieldsTipShown"`
	ProjectSettingsVcsTipShown          bool   `json:"projectSettingsVcsTipShown"`
	ProjectSettingsTimeTrackingTipShown bool   `json:"projectSettingsTimeTrackingTipShown"`
	ProjectSettingsTeamcityTipShown     bool   `json:"projectSettingsTeamcityTipShown"`
	ProjectSettingsWorkflowTipShown     bool   `json:"projectSettingsWorkflowTipShown"`
	ProjectSettingsAppsTipShown         bool   `json:"projectSettingsAppsTipShown"`
	ProjectContentTipShown              bool   `json:"projectContentTipShown"`
	IssueAiActionsTipShown              bool   `json:"issueAiActionsTipShown"`
	IssueTextRecognitionTipShown        bool   `json:"issueTextRecognitionTipShown"`
	IssueFieldsTipShown                 bool   `json:"issueFieldsTipShown"`
	CommandsTipShown                    bool   `json:"commandsTipShown"`
	TimeTrackingTipShown                bool   `json:"timeTrackingTipShown"`
	ActivityTypesTipShown               bool   `json:"activityTypesTipShown"`
	AgileBoardBacklogTipShown           bool   `json:"agileBoardBacklogTipShown"`
	AgileBoardVisibilityTipShown        bool   `json:"agileBoardVisibilityTipShown"`
	AgileBoardCardsTipShown             bool   `json:"agileBoardCardsTipShown"`
	AgileBoardSwimlanesTipShown         bool   `json:"agileBoardSwimlanesTipShown"`
	AgileBoardColumnsTipShown           bool   `json:"agileBoardColumnsTipShown"`
	ArticleSidebarTipShown              bool   `json:"articleSidebarTipShown"`
	ArticleCommentsTipShown             bool   `json:"articleCommentsTipShown"`
	ArticleAiAssistantTipShown          bool   `json:"articleAiAssistantTipShown"`
	ArticleVisibilityTipShown           bool   `json:"articleVisibilityTipShown"`
	ArticleInlineCommentsTipShown       bool   `json:"articleInlineCommentsTipShown"`
	HelpdeskPinnedCommentsTipShown      bool   `json:"helpdeskPinnedCommentsTipShown"`
	PricingAdminPopupShown              bool   `json:"pricingAdminPopupShown"`
	TextCompletionPromoShown            bool   `json:"textCompletionPromoShown"`
	TextRecognitionTipsShown            bool   `json:"textRecognitionTipsShown"`
	TextRecognitionPromoShown           bool   `json:"textRecognitionPromoShown"`
	VisibleFieldsTipShown               bool   `json:"visibleFieldsTipShown"`
	SurveyShown                         bool   `json:"surveyShown"`
	VotesTipShown                       bool   `json:"votesTipShown"`
	AiPromoShown                        bool   `json:"aiPromoShown"`
	AiTipsShown                         bool   `json:"aiTipsShown"`
	AiWritingAssistantPromoShown        bool   `json:"aiWritingAssistantPromoShown"`
	InlineCommentPromoShown             bool   `json:"inlineCommentPromoShown"`
	DelayedDemoModalShown               bool   `json:"delayedDemoModalShown"`
	SavedSearchesTipShown               bool   `json:"savedSearchesTipShown"`
	SearchOptionsTipShown               bool   `json:"searchOptionsTipShown"`
	VisibilityRestrictionsTipShown      bool   `json:"visibilityRestrictionsTipShown"`
	CollapsibleSidebarTipShown          bool   `json:"collapsibleSidebarTipShown"`
	PmfShown                            bool   `json:"pmfShown"`
	ChangeRuleTypeTipShown              bool   `json:"changeRuleTypeTipShown"`
	Type                                string `json:"$type,omitempty"`
}
