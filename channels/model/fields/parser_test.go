package fields

import (
	"testing"
)

func TestParseSimpleFields(t *testing.T) {
	query := "id,login,email,name"
	tree := Parse(query)

	if tree == nil {
		t.Fatalf("expected non-nil tree")
	}

	for _, f := range []string{"id", "login", "email", "name"} {
		if !tree.Has(f) {
			t.Errorf("expected tree to have %q", f)
		}
	}

	if tree.Has("avatarUrl") {
		t.Errorf("tree should not have avatarUrl")
	}
}

func TestParseNestedFields(t *testing.T) {
	query := "id,userType(id,name),profiles(general(timezone(id),locale(id,name)))"
	tree := Parse(query)

	if !tree.Has("id") {
		t.Errorf("expected id")
	}
	if !tree.Has("userType.id") || !tree.Has("userType.name") {
		t.Errorf("expected userType.id and userType.name")
	}
	if !tree.Has("profiles.general.timezone.id") {
		t.Errorf("expected profiles.general.timezone.id")
	}
	if !tree.Has("profiles.general.locale.name") {
		t.Errorf("expected profiles.general.locale.name")
	}
}

func TestParseWithTemplates(t *testing.T) {
	query := "id,searchContext(%40helpdeskContext);@helpdeskContext:id,name,issuesUrl,owner(%40permittedUsers);@permittedUsers:id,login,email"
	tree := Parse(query)

	if !tree.Has("id") {
		t.Errorf("expected id")
	}
	if !tree.Has("searchContext.id") {
		t.Errorf("expected searchContext.id")
	}
	if !tree.Has("searchContext.name") {
		t.Errorf("expected searchContext.name")
	}
	if !tree.Has("searchContext.owner.login") {
		t.Errorf("expected searchContext.owner.login")
	}
}

func TestParseRequest1Fields(t *testing.T) {
	req1 := `id,login,email,fullName,avatarUrl,userType(id,name),name,isEmailVerified,guest,online,banned,banBadge,canReadProfile,isLocked,issueRelatedGroup(%40permittedGroups),profiles(general(locale(id,community,language,locale,name),timezone(id),semanticSearchForArticles,dateFieldFormat(pattern,datePattern),lastCreatedIssue(project(id,shortName)),searchContext(%40helpdeskContext),helpdeskContext(%40helpdeskContext)),articles(showComments,showInlineComments,showHistory,queryMode,lastVisitedArticle(id,idReadable,reporter(%40permittedUsers),summary,project(id,name,shortName,projectType(id),pinned,iconUrl,template,archived,restricted,team(id)),parentArticle(idReadable),ordinal,visibility($type,implicitPermittedUsers(%40permittedUsers),permittedGroups(%40permittedGroups),permittedUsers(%40permittedUsers)),updatersSettings(permittedGroups(%40permittedGroups),permittedUsers(%40permittedUsers)),hasUnpublishedChanges,isUpdatable,isDeletable,collaborativeDraftId,hasChildren,tags(id,name,color(id,background,foreground),isDeletable,isUpdatable,isUsable),hasStar,updated)),timetracking(periodFormat(id),periodFieldPattern,isTimeTrackingAvailable,ownTimesheets),tips(aiTipsShown,surveyShown,pmfShown,textRecognitionTipsShown,aiPromoShown,inlineCommentPromoShown,textRecognitionPromoShown,changeRuleTypeTipShown,helpdeskPinnedCommentsTipShown,aiWritingAssistantPromoShown,textCompletionPromoShown,appsProjectTabTipShown,delayedDemoModalShown,pricingAdminPopupShown,collapsibleSidebarTipShown,savedSearchesTipShown,searchOptionsTipShown,visibleFieldsTipShown,votesTipShown,visibilityRestrictionsTipShown,issueFieldsTipShown,commandsTipShown,timeTrackingTipShown,activityTypesTipShown,agileBoardBacklogTipShown,agileBoardVisibilityTipShown,agileBoardCardsTipShown,agileBoardSwimlanesTipShown,agileBoardColumnsTipShown,articleSidebarTipShown,articleCommentsTipShown,articleAiAssistantTipShown,articleVisibilityTipShown,articleInlineCommentsTipShown,projectContentTipShown,projectOverviewTipShown,projectSettingsTipShown,projectSettingsPeopleTipShown,projectSettingsFieldsTipShown,projectSettingsVcsTipShown,projectSettingsTimeTrackingTipShown,projectSettingsTeamcityTipShown,projectSettingsWorkflowTipShown,projectSettingsAppsTipShown,issueAiActionsTipShown,issueTextRecognitionTipShown,helpdeskTeamTipShown,helpdeskSlaTipShown,helpdeskChannelsTipShown,helpdeskOverviewTipShown,issueListPageCompleted,issuePageCompleted,agileBoardPageCompleted,articlesPageCompleted,projectOverviewPageCompleted,projectSettingsPageCompleted,helpdeskProjectPageCompleted,onboardingTourState,onboardingTourAIBlockDismissed),appearance(firstDayOfWeek(),showTooltips,showRecentEntities,showSidebarResizerTip,showToolbar,showInlineEditorToolbar,useMarkdownEditor,useSummaryInIssueLinks,showQuickView,lastUsedColor,tableViewColumns,naturalCommentsOrder,showCommentsInActivityStream,showVcsChangesInActivityStream,showWorkItemsInActivityStream,showHistoryInActivityStream,expandChangesInActivityStream,useAbsoluteDates,showSimilarIssues,exceptionsExpanded,knowledgeBaseSidebarWidth,sivSidebarWidth,quickViewSidebarWidth,modalSidebarWidth,issueListSidebarWidth,quickViewWidth,showKnowledgeBaseSidebar,showSIVSidebar,attachmentsCollapsed,showLinksUnderDescription,openCwOnTyping,compactMode,recognizedTextSidebarExpanded,attachmentsSorting,hideEmbeddedAttachments,hideCommentAttachments,attachmentsListLayout,expandNavigation,issuesTableViewMode,sidebarQuickViewMode,dashboardHeaderCollapsed,knowledgeBaseSubarticlesCollapsed,aiLinkSuggestionsCollapsed,onboardingTourPanelWidth),issuesList(sortTextByRelevance,showSidebar,unresolvedIssuesOnly,queryMode,issueListView(treeView,treeViewCollapsed,detailLevel),projectsExpanded,savedSearchesExpanded,tagsExpanded),helpdesk(isAgent,isReporter,agentInProjects(id),reporterInProjects(id),tableViewTicketColumns),ai(disableChat,chatSidebarShow,chatsListShow,chatSidebarWidth,chatFloatingOffsetX,chatFloatingOffsetY,chatFloatingAnchor,chatFloatingWidth,chatFloatingHeight,chatMode),notifications(emailNotificationsEnabled,disabledDirect,disabledSubscription,disabledSystem,showSystem,showUnreadOnly,emailBlocked,emailBlockReason,usePlainTextEmails,notifyOnOwnChanges,mentionNotificationsEnabled,duplicateClusterNotificationsEnabled,mailboxIntegrationNotificationsEnabled,autoWatchOnComment,autoWatchOnCreate,autoWatchOnUpdate,autoWatchOnFieldSet,autoWatchOnVote)),featureFlags(id,enabled),widgets(id,key,appId,description,appName,appTitle,name,collapsed,configurable,indexPath,extensionPoint,iconPath,guard,appIconPath,appDarkIconPath,defaultHeight,defaultWidth,expectedHeight,expectedWidth,vendorName,vendorEmail,vendorUrl,marketplaceId,showHeader,borderless)%3B%40helpdeskContext%3Aid,name,issuesUrl,pinned,pinnedInHelpdesk,owner(%40permittedUsers),$type,query,isUpdatable,shortName%3B%40permittedUsers%3Aid,login,email,fullName,avatarUrl,userType(id,name),name,isEmailVerified,guest,online,banned,banBadge,canReadProfile,isLocked%3B%40permittedGroups%3Aid,name,$type(),auditTargetId,description,allUsersGroup,icon,teamForProject(id,name,icon),isUpdatable,isRemovable`

	tree := Parse(req1)
	if tree == nil {
		t.Fatalf("expected non-nil tree for request 1")
	}

	if !tree.Has("login") {
		t.Errorf("expected login")
	}
	if !tree.Has("profiles.appearance.compactMode") {
		t.Errorf("expected profiles.appearance.compactMode")
	}
	if !tree.Has("profiles.tips.onboardingTourState") {
		t.Errorf("expected profiles.tips.onboardingTourState")
	}
	if !tree.Has("profiles.ai.chatMode") {
		t.Errorf("expected profiles.ai.chatMode")
	}
	if !tree.Has("featureFlags.id") {
		t.Errorf("expected featureFlags.id")
	}
	if !tree.Has("widgets.name") {
		t.Errorf("expected widgets.name")
	}
}

func TestParseEmptyAndNil(t *testing.T) {
	tree := Parse("")
	if tree != nil {
		t.Errorf("expected nil tree for empty query")
	}
}

func TestParseMalformedFields(t *testing.T) {
	tree := Parse("id,userType(name,invalid(")
	if tree == nil {
		t.Fatalf("expected non-nil tree even on malformed input")
	}
	if !tree.Has("id") {
		t.Errorf("expected tree to still contain 'id'")
	}
}
