package domain

// DefaultProfiles يبني مخطط البروفايلات الكامل بقيم افتراضية تطابق الشكل المرجعي
// في request1.txt لملف /api/users/me. تُدمج قيم قاعدة البيانات فوقه عند الحاجة
// (انظر loadCurrentUserProfiles)، لضمان خروج كافة الأقسام بالبنية الكاملة حتى
// عندما لا تكون مخزنة في قاعدة البيانات.
func DefaultProfiles() *CurrentUserProfiles {
	cp := &CurrentUserProfiles{
		General:       defaultGeneralProfile(),
		Articles:      defaultArticlesProfile(),
		TimeTracking:  defaultTimeTrackingProfile(),
		Tips:          defaultTipsProfile(),
		Appearance:    defaultAppearanceProfile(),
		IssuesList:    defaultIssuesListProfile(),
		Helpdesk:      defaultHelpdeskProfile(),
		AI:            defaultAIProfile(),
		Notifications: defaultNotificationsProfile(),
		Type:          "UserProfiles",
	}
	return cp
}

// DefaultUserProfiles يبني النسخة الموسّعة من البروفايلات (مع قسم Grazie)
// للمسار القديم GetCurrentUser في user_usecase.go.
func DefaultUserProfiles() *UserProfiles {
	cp := DefaultProfiles()
	return &UserProfiles{
		General:       cp.General,
		Appearance:    cp.Appearance,
		IssuesList:    cp.IssuesList,
		Articles:      cp.Articles,
		Notifications: cp.Notifications,
		TimeTracking:  cp.TimeTracking,
		Tips:          cp.Tips,
		Helpdesk:      cp.Helpdesk,
		AI:            cp.AI,
		Grazie: &GrazieUserProfile{
			ExcludedIssueTypes:            "",
			EnableSpellChecker:            true,
			HasMoreTokens:                 true,
			EnableTextCompletion:          true,
			CycleRestart:                  1787653099967,
			FreeLicense:                   false,
			SpellCheckerEnabledInSystem:   true,
			TextCompletionEnabledInSystem: true,
			Enabled:                       true,
			Type:                          "GrazieUserProfile",
		},
		Type: "UserProfiles",
	}
}

func defaultGeneralProfile() *GeneralUserProfile {
	return &GeneralUserProfile{
		Timezone: &TimeZoneDescriptor{
			ID:   "Europe/Prague",
			Type: "TimeZoneDescriptor",
		},
		DateFieldFormat: &DateFormatDescriptor{
			Pattern:     "d MMM yyyy HH:mm",
			DatePattern: "d MMM yyyy",
			Type:        "DateFormatDescriptor",
		},
		Locale: &LocaleDescriptor{
			ID:        "en_US",
			Language:  "en",
			Locale:    "en_US",
			Name:      "English",
			Community: false,
			Type:      "LocaleDescriptor",
		},
		SemanticSearchForArticles: false,
		Type:                      "GeneralUserProfile",
	}
}

func defaultArticlesProfile() *ArticlesUserProfile {
	return &ArticlesUserProfile{
		ShowComments:       true,
		ShowInlineComments: false,
		ShowHistory:        false,
		QueryMode:          false,
		Type:               "ArticlesUserProfile",
	}
}

func defaultTimeTrackingProfile() *TimeTrackingUserProfile {
	return &TimeTrackingUserProfile{
		PeriodFormat: &PeriodFieldFormat{
			ID:   "FULL",
			Type: "PeriodFieldFormat",
		},
		PeriodFieldPattern:      `((\d+w\s*)?(\s*)?(\d+d\s*)?(\s*)?(\d+h\s*)?(\s*)?(\d+m\s*)?|\d+)`,
		IsTimeTrackingAvailable: false,
		OwnTimesheets:           false,
		Type:                    "TimeTrackingUserProfile",
	}
}

func defaultTipsProfile() *TipsUserProfile {
	return &TipsUserProfile{
		OnboardingTourState:                 "idle",
		InlineCommentPromoShown:             true,
		IssueListPageCompleted:              false,
		IssuePageCompleted:                  false,
		AgileBoardPageCompleted:             false,
		ArticlesPageCompleted:               false,
		ProjectOverviewPageCompleted:        false,
		ProjectSettingsPageCompleted:        false,
		HelpdeskProjectPageCompleted:        false,
		OnboardingTourAIBlockDismissed:      false,
		HelpdeskTeamTipShown:                false,
		HelpdeskSlaTipShown:                 false,
		HelpdeskChannelsTipShown:            false,
		HelpdeskOverviewTipShown:            false,
		ProjectOverviewTipShown:             false,
		ProjectSettingsTipShown:             false,
		ProjectSettingsPeopleTipShown:       false,
		ProjectSettingsFieldsTipShown:       false,
		ProjectSettingsVcsTipShown:          false,
		ProjectSettingsTimeTrackingTipShown: false,
		ProjectSettingsTeamcityTipShown:     false,
		ProjectSettingsWorkflowTipShown:     false,
		ProjectSettingsAppsTipShown:         false,
		ProjectContentTipShown:              false,
		IssueAiActionsTipShown:              false,
		IssueTextRecognitionTipShown:        false,
		IssueFieldsTipShown:                 false,
		CommandsTipShown:                    false,
		TimeTrackingTipShown:                false,
		ActivityTypesTipShown:               false,
		AgileBoardBacklogTipShown:           false,
		AgileBoardVisibilityTipShown:        false,
		AgileBoardCardsTipShown:             false,
		AgileBoardSwimlanesTipShown:         false,
		AgileBoardColumnsTipShown:           false,
		ArticleSidebarTipShown:              false,
		ArticleCommentsTipShown:             false,
		ArticleAiAssistantTipShown:          false,
		ArticleVisibilityTipShown:           false,
		ArticleInlineCommentsTipShown:       false,
		HelpdeskPinnedCommentsTipShown:      false,
		PricingAdminPopupShown:              false,
		TextCompletionPromoShown:            false,
		TextRecognitionTipsShown:            false,
		TextRecognitionPromoShown:           false,
		VisibleFieldsTipShown:               false,
		SurveyShown:                         false,
		VotesTipShown:                       false,
		AiPromoShown:                        false,
		AiTipsShown:                         false,
		AiWritingAssistantPromoShown:        false,
		DelayedDemoModalShown:               false,
		SavedSearchesTipShown:               false,
		SearchOptionsTipShown:               false,
		VisibilityRestrictionsTipShown:      false,
		CollapsibleSidebarTipShown:          false,
		PmfShown:                            false,
		ChangeRuleTypeTipShown:              false,
		Type:                                "TipsUserProfile",
	}
}

func defaultAppearanceProfile() *AppearanceUserProfile {
	return &AppearanceUserProfile{
		FirstDayOfWeek:                    0,
		ShowTooltips:                      true,
		ShowRecentEntities:                true,
		ShowSidebarResizerTip:             true,
		ShowToolbar:                       true,
		ShowInlineEditorToolbar:           true,
		UseMarkdownEditor:                 false,
		UseSummaryInIssueLinks:            true,
		ShowQuickView:                     true,
		LastUsedColor:                     "default",
		TableViewColumns:                  []string{},
		NaturalCommentsOrder:              true,
		ShowCommentsInActivityStream:      true,
		ShowVcsChangesInActivityStream:    false,
		ShowWorkItemsInActivityStream:     false,
		ShowHistoryInActivityStream:       false,
		ExpandChangesInActivityStream:     false,
		UseAbsoluteDates:                  false,
		ShowSimilarIssues:                 true,
		ExceptionsExpanded:                true,
		KnowledgeBaseSidebarWidth:         240,
		SivSidebarWidth:                   240,
		QuickViewSidebarWidth:             240,
		ModalSidebarWidth:                 240,
		IssueListSidebarWidth:             240,
		QuickViewWidth:                    0,
		ShowKnowledgeBaseSidebar:          true,
		ShowSIVSidebar:                    false,
		AttachmentsCollapsed:              false,
		ShowLinksUnderDescription:         true,
		OpenCwOnTyping:                    true,
		CompactMode:                       false,
		RecognizedTextSidebarExpanded:     false,
		AttachmentsSorting:                "default",
		HideEmbeddedAttachments:           true,
		HideCommentAttachments:            true,
		AttachmentsListLayout:             false,
		ExpandNavigation:                  true,
		IssuesTableViewMode:               true,
		SidebarQuickViewMode:              true,
		DashboardHeaderCollapsed:          false,
		KnowledgeBaseSubarticlesCollapsed: false,
		AiLinkSuggestionsCollapsed:        false,
		OnboardingTourPanelWidth:          400,
		Type:                              "AppearanceUserProfile",
	}
}

func defaultIssuesListProfile() *IssuesListUserProfile {
	return &IssuesListUserProfile{
		SortTextByRelevance:  true,
		ShowSidebar:          true,
		UnresolvedIssuesOnly: false,
		QueryMode:            false,
		IssueListView: &IssueListView{
			TreeView:          true,
			TreeViewCollapsed: false,
			DetailLevel:       20,
			Type:              "IssueListView",
		},
		ProjectsExpanded:      true,
		SavedSearchesExpanded: true,
		TagsExpanded:          true,
		Type:                  "IssuesListUserProfile",
	}
}

func defaultHelpdeskProfile() *HelpdeskUserProfile {
	return &HelpdeskUserProfile{
		IsReporter:             false,
		IsAgent:                false,
		AgentInProjects:        []*Project{},
		ReporterInProjects:     []*Project{},
		TableViewTicketColumns: []string{},
		Type:                   "HelpdeskUserProfile",
	}
}

func defaultAIProfile() *AiUserProfile {
	return &AiUserProfile{
		DisableChat:         false,
		ChatSidebarShow:     false,
		ChatsListShow:       true,
		ChatMode:            "docked",
		ChatSidebarWidth:    400,
		ChatFloatingAnchor:  "top-right",
		ChatFloatingOffsetX: 100,
		ChatFloatingOffsetY: 100,
		ChatFloatingWidth:   400,
		ChatFloatingHeight:  600,
		Type:                "AiUserProfile",
	}
}

func defaultNotificationsProfile() *NotificationsUserProfile {
	return &NotificationsUserProfile{
		EmailNotificationsEnabled:              true,
		DisabledDirect:                         false,
		DisabledSubscription:                   false,
		DisabledSystem:                         false,
		ShowSystem:                             false,
		ShowUnreadOnly:                         false,
		EmailBlocked:                           false,
		UsePlainTextEmails:                     false,
		NotifyOnOwnChanges:                     false,
		MentionNotificationsEnabled:            true,
		DuplicateClusterNotificationsEnabled:   false,
		MailboxIntegrationNotificationsEnabled: true,
		AutoWatchOnComment:                     true,
		AutoWatchOnCreate:                      true,
		AutoWatchOnUpdate:                      false,
		AutoWatchOnFieldSet:                    true,
		AutoWatchOnVote:                        true,
		Type:                                   "NotificationsUserProfile",
	}
}
