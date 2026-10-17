package usershandler

import (
	"errors"
	"log/slog"
	"net/http"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	usersusecase "youtrack_backend/internal/usecases/users"
)

type ParamsRequest struct {
}

type FeatureFlag struct {
	ID      string `json:"id" db:"id"`
	Enabled bool   `json:"enabled" db:"enabled"`
	Type    string `json:"$type"`
}

type UserType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"$type"`
}

type UserProfiles struct {
	Timetracking  *TimeTrackingUserProfile  `json:"timetracking,omitempty"`
	General       *GeneralUserProfile       `json:"general,omitempty"`
	Appearance    *AppearanceUserProfile    `json:"appearance,omitempty"`
	IssuesList    *IssuesListUserProfile    `json:"issuesList,omitempty"`
	Articles      *ArticlesUserProfile      `json:"articles,omitempty"`
	Notifications *NotificationsUserProfile `json:"notifications,omitempty"`
	Type          string                    `json:"$type,omitempty"`
}

// AppearanceUserProfile يحوي كافة إعدادات الواجهة والمظهر.
type AppearanceUserProfile struct {
	OpenCwOnTyping                    bool   `json:"openCwOnTyping"`
	SivSidebarWidth                   int    `json:"sivSidebarWidth"`
	ShowToolbar                       bool   `json:"showToolbar"`
	ShowQuickView                     bool   `json:"showQuickView"`
	ShowSimilarIssues                 bool   `json:"showSimilarIssues"`
	ShowKnowledgeBaseSidebar          bool   `json:"showKnowledgeBaseSidebar"`
	ShowSIVSidebar                    bool   `json:"showSIVSidebar"`
	DashboardHeaderCollapsed          bool   `json:"dashboardHeaderCollapsed"`
	KnowledgeBaseSubarticlesCollapsed bool   `json:"knowledgeBaseSubarticlesCollapsed"`
	AiLinkSuggestionsCollapsed        bool   `json:"aiLinkSuggestionsCollapsed"`
	ShowCommentsInActivityStream      bool   `json:"showCommentsInActivityStream"`
	KnowledgeBaseSidebarWidth         int    `json:"knowledgeBaseSidebarWidth"`
	LastUsedColor                     string `json:"lastUsedColor"`
	AttachmentsListLayout             bool   `json:"attachmentsListLayout"`
	ShowTooltips                      bool   `json:"showTooltips"`
	ExpandNavigation                  bool   `json:"expandNavigation"`
	HideCommentAttachments            bool   `json:"hideCommentAttachments"`
	SidebarQuickViewMode              bool   `json:"sidebarQuickViewMode"`
	ExpandChangesInActivityStream     bool   `json:"expandChangesInActivityStream"`
	NaturalCommentsOrder              bool   `json:"naturalCommentsOrder"`
	UseAbsoluteDates                  bool   `json:"useAbsoluteDates"`
	UseMarkdownEditor                 bool   `json:"useMarkdownEditor"`
	ShowInlineEditorToolbar           bool   `json:"showInlineEditorToolbar"`
	ShowVcsChangesInActivityStream    bool   `json:"showVcsChangesInActivityStream"`
	TableViewColumns                  []any  `json:"tableViewColumns"`
	IssueListSidebarWidth             int    `json:"issueListSidebarWidth"`
	AttachmentsCollapsed              bool   `json:"attachmentsCollapsed"`
	UseSummaryInIssueLinks            bool   `json:"useSummaryInIssueLinks"`
	ShowRecentEntities                bool   `json:"showRecentEntities"`
	ExceptionsExpanded                bool   `json:"exceptionsExpanded"`
	OnboardingTourPanelWidth          int    `json:"onboardingTourPanelWidth"`
	ShowLinksUnderDescription         bool   `json:"showLinksUnderDescription"`
	RecognizedTextSidebarExpanded     bool   `json:"recognizedTextSidebarExpanded"`
	QuickViewSidebarWidth             int    `json:"quickViewSidebarWidth"`
	ModalSidebarWidth                 int    `json:"modalSidebarWidth"`
	ShowSidebarResizerTip             bool   `json:"showSidebarResizerTip"`
	IssuesTableViewMode               bool   `json:"issuesTableViewMode"`
	AttachmentsSorting                string `json:"attachmentsSorting"`
	HideEmbeddedAttachments           bool   `json:"hideEmbeddedAttachments"`
	CompactMode                       bool   `json:"compactMode"`
	QuickViewWidth                    int    `json:"quickViewWidth"`
	ShowHistoryInActivityStream       bool   `json:"showHistoryInActivityStream"`
	ShowWorkItemsInActivityStream     bool   `json:"showWorkItemsInActivityStream"`
	FirstDayOfWeek                    int    `json:"firstDayOfWeek"`
	Type                              string `json:"$type"`
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

// PeriodFieldFormat يصف صيغة المدة الزمنية في تتبع الوقت.
type PeriodFieldFormat struct {
	ID   string `json:"id"`
	Type string `json:"$type"`
}

// DateFormatDescriptor يصف صيغة التاريخ والوقت.
type DateFormatDescriptor struct {
	Pattern     string `json:"pattern"`
	DatePattern string `json:"datePattern"`
	Type        string `json:"$type"`
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

// TimeTrackingUserProfile يحوي إعدادات تتبع الوقت للمستخدم.
type TimeTrackingUserProfile struct {
	PeriodFormat            *PeriodFieldFormat `json:"periodFormat"`
	OwnTimesheets           bool               `json:"ownTimesheets"`
	PeriodFieldPattern      string             `json:"periodFieldPattern"`
	IsTimeTrackingAvailable bool               `json:"isTimeTrackingAvailable"`
	Type                    string             `json:"$type"`
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

// IssueListView يصف خيارات عرض قائمة المشاكل.
type IssueListView struct {
	DetailLevel       int    `json:"detailLevel"`
	TreeView          bool   `json:"treeView"`
	TreeViewCollapsed bool   `json:"treeViewCollapsed"`
	Type              string `json:"$type"`
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

type ParamsResponse struct {
	Widgets         []DashboardWidget `json:"widgets"`
	FeatureFlags    []FeatureFlag     `json:"featureFlags"`
	Profiles        UserProfiles      `json:"profiles"`
	Guest           bool              `json:"guest"`
	Banned          bool              `json:"banned"`
	Login           string            `json:"login"`
	Email           string            `json:"email"`
	UserType        UserType          `json:"userType"`
	AvatarURL       string            `json:"avatarUrl"`
	Online          bool              `json:"online"`
	IsEmailVerified bool              `json:"isEmailVerified"`
	BanBadge        string            `json:"banBadge"`
	CanReadProfile  bool              `json:"canReadProfile"`
	FullName        string            `json:"fullName"`
	Name            string            `json:"name"`
	IsLocked        bool              `json:"isLocked"`
	ID              string            `json:"id"`
	Type            string            `json:"$type"`
}

const meType = "jetbrains.jetpass.tracker.User"

func (hlr *Handler) handleMe(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(request.Context())
	if !ok {
		response.Error(writer, platErrors.Unauthorized("me", "authentication required", errors.New("missing user identity")))
		return
	}
	roles, ok := []string{}, false // middlewares.GetCurrentUserRoles(request.Context())
	if !ok {
		roles = []string{"user"}
	}

	result, err := hlr.users.Me(request.Context(), &usersusecase.MeParams{UserID: userID, Roles: roles})
	if err != nil {
		hlr.logger.LogAttrs(request.Context(), slog.LevelWarn, "me rejected",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		writeMeError(writer, err)
		return
	}

	u := result.User
	response.JSON(writer, http.StatusOK, ParamsResponse{
		Widgets:         []DashboardWidget{},
		FeatureFlags:    []FeatureFlag{},
		Profiles:        UserProfiles{Type: "UserProfiles"},
		Guest:           u.Guest,
		Banned:          u.Banned,
		Login:           u.Login,
		Email:           u.Email,
		UserType:        UserType{ID: u.UserTypeID, Name: u.UserTypeID, Type: "UserType"},
		AvatarURL:       u.AvatarURL,
		Online:          u.Online,
		IsEmailVerified: u.IsEmailVerified,
		BanBadge:        u.BanBadge,
		CanReadProfile:  u.CanReadProfile,
		FullName:        u.FullName,
		Name:            u.Name,
		IsLocked:        u.IsLocked,
		ID:              u.ID,
		Type:            meType,
	})
}

func writeMeError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usersusecase.ErrUserNotFound):
		response.Error(writer, platErrors.Unauthorized("me", "user not found", err))
	case errors.Is(err, usersusecase.ErrUserUnavailable):
		response.Error(writer, platErrors.Forbidden("me", "user account is unavailable", err))
	default:
		response.Error(writer, platErrors.Internal("me", "me unavailable", err))
	}
}
