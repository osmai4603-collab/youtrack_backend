package sqlstore

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// UserStore تطبيق عمليات المستخدمين على PostgreSQL.
type UserStore struct {
	db *pgxpool.Pool
}

// scanUser يقرأ صفًا واحدًا من users إلى model.User.
func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	u := &model.User{}
	var userTypeID *string
	var banBadge *string
	err := row.Scan(
		&u.ID, &u.Login, &u.Email, &u.FullName, &u.Name, &u.AvatarURL, &userTypeID,
		&u.IsEmailVerified, &u.Guest, &u.Online, &u.Banned, &banBadge, &u.CanReadProfile, &u.IsLocked, &u.RingID, &u.PasswordHash,
	)
	if err != nil {
		return nil, err
	}
	u.BanBadge = banBadge
	if userTypeID != nil {
		u.UserTypeID = *userTypeID
		u.UserType = &model.UserType{ID: *userTypeID, Name: "Standard user", Type: "UserType"}
	}
	return u, nil
}

const userSelect = `SELECT id, login, email, full_name, name, avatar_url, user_type_id,
	is_email_verified, guest, online, banned, ban_badge, can_read_profile, is_locked, ring_id, password_hash FROM users`

func (s *UserStore) GetByID(ctx context.Context, id string) (*model.User, error) {
	row := s.db.QueryRow(ctx, userSelect+` WHERE id = $1`, id)
	return scanUser(row)
}

func (s *UserStore) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	row := s.db.QueryRow(ctx, userSelect+` WHERE login = $1`, login)
	return scanUser(row)
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := s.db.QueryRow(ctx, userSelect+` WHERE email = $1`, email)
	return scanUser(row)
}

func (s *UserStore) Create(ctx context.Context, u *model.User) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO users (id, login, email, full_name, name, avatar_url, user_type_id, is_email_verified, guest, online, banned, ban_badge, can_read_profile, is_locked, ring_id, password_hash)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET
			login = EXCLUDED.login,
			name = EXCLUDED.name,
			password_hash = COALESCE(EXCLUDED.password_hash, users.password_hash)`,
		u.ID, u.Login, u.Email, u.FullName, u.Name, u.AvatarURL, u.UserTypeID,
		u.IsEmailVerified, u.Guest, u.Online, u.Banned, u.BanBadge, u.CanReadProfile, u.IsLocked, u.RingID, u.PasswordHash)
	return err
}

func (s *UserStore) All(ctx context.Context) ([]*model.User, error) {
	rows, err := s.db.Query(ctx, userSelect)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *UserStore) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	p := &model.UserProfile{UserID: userID}
	err := s.db.QueryRow(ctx, `
		SELECT timezone_id, locale_id, email_notifications_enabled, mention_notifications_enabled,
		       compact_mode, expand_navigation, is_time_tracking_available
		FROM user_profiles WHERE user_id = $1`, userID).
		Scan(&p.TimezoneID, &p.LocaleID, &p.EmailNotifications, &p.MentionNotifications,
			&p.CompactMode, &p.ExpandNavigation, &p.IsTimeTrackingAvailable)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *UserStore) CreateProfile(ctx context.Context, p *model.UserProfile) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, timezone_id, locale_id, email_notifications_enabled, mention_notifications_enabled, compact_mode, expand_navigation, is_time_tracking_available)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (user_id) DO UPDATE SET
			timezone_id = EXCLUDED.timezone_id,
			locale_id = EXCLUDED.locale_id,
			compact_mode = EXCLUDED.compact_mode,
			expand_navigation = EXCLUDED.expand_navigation`,
		p.UserID, p.TimezoneID, p.LocaleID, p.EmailNotifications, p.MentionNotifications,
		p.CompactMode, p.ExpandNavigation, p.IsTimeTrackingAvailable)
	return err
}

// GetMe يجلب المستخدم الحالي مع الجلب الانتقائي الدقيق لكل جدول حسب شجرة الحقول FieldTree.
func (s *UserStore) GetMe(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error) {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. جلب اسم نوع المستخدم الفعلي فقط إذا طُلب
	if tree == nil || tree.Has("userType") {
		if u.UserTypeID != "" {
			var typeName string
			if err := s.db.QueryRow(ctx, `SELECT name FROM user_types WHERE id = $1`, u.UserTypeID).Scan(&typeName); err == nil && typeName != "" {
				u.UserType = &model.UserType{ID: u.UserTypeID, Name: typeName, Type: "UserType"}
			}
		}
	} else {
		u.UserType = nil
	}

	// 2. فحص وجلب featureFlags فقط عند الطلب
	if tree == nil || tree.Has("featureFlags") {
		flags, err := s.GetFeatureFlags(ctx)
		if err == nil {
			u.FeatureFlags = flags
		} else {
			u.FeatureFlags = []*model.FeatureFlag{}
		}
	} else {
		u.FeatureFlags = nil
	}

	// 3. فحص وجلب widgets فقط عند الطلب
	if tree == nil || tree.Has("widgets") {
		widgets, err := s.getDashboardWidgets(ctx)
		if err == nil {
			u.Widgets = widgets
		} else {
			u.Widgets = []*model.DashboardWidget{}
		}
	} else {
		u.Widgets = nil
	}

	// 4. فحص وجلب issueRelatedGroup فقط عند الطلب
	if tree == nil || tree.Has("issueRelatedGroup") {
		groups, err := s.getUserGroups(ctx, userID)
		if err == nil && len(groups) > 0 {
			u.IssueRelatedGroup = &model.IssueRelatedGroup{
				PermittedGroups: groups,
				Type:            "IssueRelatedGroup",
			}
		} else {
			u.IssueRelatedGroup = nil
		}
	} else {
		u.IssueRelatedGroup = nil
	}

	// 5. فحص وبناء كائن profiles وجلب جداوله الفرعية فقط إذا كان مطلوباً
	if tree == nil || tree.Has("profiles") {
		profiles := model.DefaultUserProfiles()

		// قراءة جدول user_profiles الأساسي إذا طُلب general أو appearance أو notifications أو timetracking
		if tree == nil || tree.Has("profiles.general") || tree.Has("profiles.appearance") || tree.Has("profiles.notifications") || tree.Has("profiles.timetracking") {
			if p, err := s.GetProfile(ctx, userID); err == nil && p != nil {
				if p.TimezoneID != "" {
					profiles.General.Timezone.ID = p.TimezoneID
				}
				if p.LocaleID != "" {
					profiles.General.Locale.ID = p.LocaleID
				}
				profiles.Appearance.CompactMode = p.CompactMode
				profiles.Appearance.ExpandNavigation = p.ExpandNavigation
				profiles.Notifications.EmailNotificationsEnabled = p.EmailNotifications
				profiles.Notifications.MentionNotificationsEnabled = p.MentionNotifications
				profiles.Timetracking.IsTimeTrackingAvailable = p.IsTimeTrackingAvailable
			}
		}

		// قراءة الجداول الموسعة بشكل انتقائي حصري بناءً على ما طُلب في الشجرة
		s.populateExtendedProfiles(ctx, userID, profiles, tree)

		// إزالة الكائنات الفرعية التي لم يتم طلبها صراحة عند تخصيص الـ fields
		if tree != nil && tree.Child("profiles") != nil {
			profTree := tree.Child("profiles")
			if !profTree.Has("ai") {
				profiles.AI = nil
			}
			if !profTree.Has("tips") {
				profiles.Tips = nil
			}
			if !profTree.Has("helpdesk") {
				profiles.Helpdesk = nil
			}
			if !profTree.Has("timetracking") {
				profiles.Timetracking = nil
			}
			if !profTree.Has("general") {
				profiles.General = nil
			}
			if !profTree.Has("appearance") {
				profiles.Appearance = nil
			}
			if !profTree.Has("issuesList") {
				profiles.IssuesList = nil
			}
			if !profTree.Has("articles") {
				profiles.Articles = nil
			}
			if !profTree.Has("notifications") {
				profiles.Notifications = nil
			}
		}

		u.Profiles = profiles
	} else {
		u.Profiles = nil
	}

	u.NormalizeMe()
	return u, nil
}

// GetFeatureFlags يجلب جميع ميزات النظام المفعلة من جدول feature_flags.
func (s *UserStore) GetFeatureFlags(ctx context.Context) ([]*model.FeatureFlag, error) {
	rows, err := s.db.Query(ctx, `SELECT id, enabled FROM feature_flags ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flags := []*model.FeatureFlag{}
	for rows.Next() {
		f := &model.FeatureFlag{Type: "FeatureFlag"}
		if err := rows.Scan(&f.ID, &f.Enabled); err != nil {
			return nil, err
		}
		flags = append(flags, f)
	}
	return flags, rows.Err()
}

// getDashboardWidgets يجلب الودجات العامة من جدول dashboard_widgets.
func (s *UserStore) getDashboardWidgets(ctx context.Context) ([]*model.DashboardWidget, error) {
	widgets, err := loadDashboardWidgets(ctx, s.db)
	if err != nil {
		return []*model.DashboardWidget{}, nil
	}
	return widgets, nil
}

// loadDashboardWidgets يجلب الودجات العامة من جدول dashboard_widgets.
func loadDashboardWidgets(ctx context.Context, db *pgxpool.Pool) ([]*model.DashboardWidget, error) {
	rows, err := db.Query(ctx, `
		SELECT id, COALESCE(key,''), COALESCE(app_id,''), COALESCE(description,''), COALESCE(app_name,''),
		       COALESCE(app_title,''), COALESCE(name,''), collapsed, configurable, COALESCE(index_path,''),
		       COALESCE(extension_point,''), COALESCE(icon_path,''), COALESCE(guard,''), COALESCE(app_icon_path,''),
		       COALESCE(app_dark_icon_path,''), default_height, default_width, expected_height, expected_width,
		       COALESCE(vendor_name,''), COALESCE(vendor_email,''), COALESCE(vendor_url,''), marketplace_id,
		       show_header, borderless
		FROM dashboard_widgets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	widgets := []*model.DashboardWidget{}
	for rows.Next() {
		w := &model.DashboardWidget{Type: "DashboardWidget"}
		err := rows.Scan(
			&w.ID, &w.Key, &w.AppID, &w.Description, &w.AppName,
			&w.AppTitle, &w.Name, &w.Collapsed, &w.Configurable, &w.IndexPath,
			&w.ExtensionPoint, &w.IconPath, &w.Guard, &w.AppIconPath,
			&w.AppDarkIconPath, &w.DefaultHeight, &w.DefaultWidth, &w.ExpectedHeight, &w.ExpectedWidth,
			&w.VendorName, &w.VendorEmail, &w.VendorURL, &w.MarketplaceID,
			&w.ShowHeader, &w.Borderless,
		)
		if err != nil {
			return nil, err
		}
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// getUserGroups يجلب مجموعات المستخدم من user_groups و user_group_members.
func (s *UserStore) getUserGroups(ctx context.Context, userID string) ([]*model.UserGroup, error) {
	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.name, COALESCE(g.group_type,''), g.all_users_group, COALESCE(g.icon,''),
		       COALESCE(g.description,''), COALESCE(g.audit_target_id,''), g.is_updatable, g.is_removable
		FROM user_groups g
		JOIN user_group_members m ON g.id = m.group_id
		WHERE m.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*model.UserGroup{}
	for rows.Next() {
		g := &model.UserGroup{Type: "UserGroup"}
		err := rows.Scan(&g.ID, &g.Name, &g.GroupType, &g.AllUsersGroup, &g.Icon, &g.Description, &g.AuditTargetID, &g.IsUpdatable, &g.IsRemovable)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// populateExtendedProfiles يملأ البيانات من الجداول الموسعة بشكل انتقائي حصري بناءً على ما طُلب في FieldTree.
func (s *UserStore) populateExtendedProfiles(ctx context.Context, userID string, p *model.UserProfiles, tree *fields.FieldTree) {
	// 1. جدول user_profile_appearance: يتم الاستعلام عنه فقط إذا طُلب profiles.appearance
	if tree == nil || tree.Has("profiles.appearance") {
		_ = s.db.QueryRow(ctx, `
			SELECT open_cw_on_typing, siv_sidebar_width, show_toolbar, show_quick_view, show_similar_issues,
			       show_knowledge_base_sidebar, show_siv_sidebar, dashboard_header_collapsed, knowledge_base_subarticles_collapsed,
			       ai_link_suggestions_collapsed, show_comments_in_activity_stream, knowledge_base_sidebar_width, last_used_color,
			       attachments_list_layout, show_tooltips, expand_navigation, hide_comment_attachments, sidebar_quick_view_mode,
			       expand_changes_in_activity_stream, natural_comments_order, use_absolute_dates, use_markdown_editor,
			       show_inline_editor_toolbar, show_vcs_changes_in_activity_stream, issue_list_sidebar_width, attachments_collapsed,
			       use_summary_in_issue_links, show_recent_entities, exceptions_expanded, onboarding_tour_panel_width,
			       show_links_under_description, recognized_text_sidebar_expanded, quick_view_sidebar_width, modal_sidebar_width,
			       show_sidebar_resizer_tip, issues_table_view_mode, attachments_sorting, hide_embedded_attachments,
			       compact_mode, quick_view_width, show_history_in_activity_stream, show_work_items_in_activity_stream, first_day_of_week
			FROM user_profile_appearance WHERE user_id = $1`, userID).
			Scan(
				&p.Appearance.OpenCwOnTyping, &p.Appearance.SivSidebarWidth, &p.Appearance.ShowToolbar, &p.Appearance.ShowQuickView, &p.Appearance.ShowSimilarIssues,
				&p.Appearance.ShowKnowledgeBaseSidebar, &p.Appearance.ShowSIVSidebar, &p.Appearance.DashboardHeaderCollapsed, &p.Appearance.KnowledgeBaseSubarticlesCollapsed,
				&p.Appearance.AiLinkSuggestionsCollapsed, &p.Appearance.ShowCommentsInActivityStream, &p.Appearance.KnowledgeBaseSidebarWidth, &p.Appearance.LastUsedColor,
				&p.Appearance.AttachmentsListLayout, &p.Appearance.ShowTooltips, &p.Appearance.ExpandNavigation, &p.Appearance.HideCommentAttachments, &p.Appearance.SidebarQuickViewMode,
				&p.Appearance.ExpandChangesInActivityStream, &p.Appearance.NaturalCommentsOrder, &p.Appearance.UseAbsoluteDates, &p.Appearance.UseMarkdownEditor,
				&p.Appearance.ShowInlineEditorToolbar, &p.Appearance.ShowVcsChangesInActivityStream, &p.Appearance.IssueListSidebarWidth, &p.Appearance.AttachmentsCollapsed,
				&p.Appearance.UseSummaryInIssueLinks, &p.Appearance.ShowRecentEntities, &p.Appearance.ExceptionsExpanded, &p.Appearance.OnboardingTourPanelWidth,
				&p.Appearance.ShowLinksUnderDescription, &p.Appearance.RecognizedTextSidebarExpanded, &p.Appearance.QuickViewSidebarWidth, &p.Appearance.ModalSidebarWidth,
				&p.Appearance.ShowSidebarResizerTip, &p.Appearance.IssuesTableViewMode, &p.Appearance.AttachmentsSorting, &p.Appearance.HideEmbeddedAttachments,
				&p.Appearance.CompactMode, &p.Appearance.QuickViewWidth, &p.Appearance.ShowHistoryInActivityStream, &p.Appearance.ShowWorkItemsInActivityStream, &p.Appearance.FirstDayOfWeek,
			)
	}

	// 2. جدول user_profile_ai: يتم الاستعلام عنه فقط إذا طُلب profiles.ai
	if tree == nil || tree.Has("profiles.ai") {
		_ = s.db.QueryRow(ctx, `
			SELECT chats_list_show, chat_floating_width, chat_mode, chat_sidebar_show, chat_floating_offset_y,
			       chat_floating_offset_x, chat_floating_anchor, chat_floating_height, chat_sidebar_width, disable_chat
			FROM user_profile_ai WHERE user_id = $1`, userID).
			Scan(
				&p.AI.ChatsListShow, &p.AI.ChatFloatingWidth, &p.AI.ChatMode, &p.AI.ChatSidebarShow, &p.AI.ChatFloatingOffsetY,
				&p.AI.ChatFloatingOffsetX, &p.AI.ChatFloatingAnchor, &p.AI.ChatFloatingHeight, &p.AI.ChatSidebarWidth, &p.AI.DisableChat,
			)
	}

	// 3. جدول user_profile_tips: يتم الاستعلام عنه فقط إذا طُلب profiles.tips
	if tree == nil || tree.Has("profiles.tips") {
		_ = s.db.QueryRow(ctx, `
			SELECT issue_list_page_completed, issue_page_completed, helpdesk_team_tip_shown, project_overview_tip_shown,
			       project_settings_tip_shown, issue_ai_actions_tip_shown, project_content_tip_shown, agile_board_cards_tip_shown,
			       apps_project_tab_tip_shown, change_rule_type_tip_shown, agile_board_visibility_tip_shown, agile_board_swimlanes_tip_shown,
			       agile_board_columns_tip_shown, article_ai_assistant_tip_shown, article_visibility_tip_shown, article_inline_comments_tip_shown,
			       project_settings_people_tip_shown, project_settings_fields_tip_shown, project_settings_vcs_tip_shown, helpdesk_pinned_comments_tip_shown,
			       pricing_admin_popup_shown, text_completion_promo_shown, text_recognition_tips_shown, article_sidebar_tip_shown,
			       article_comments_tip_shown, activity_types_tip_shown, issue_fields_tip_shown, commands_tip_shown,
			       time_tracking_tip_shown, visible_fields_tip_shown, survey_shown, votes_tip_shown,
			       ai_promo_shown, collapsible_sidebar_tip_shown, pmf_shown, ai_writing_assistant_promo_shown,
			       inline_comment_promo_shown, project_settings_time_tracking_tip_shown, project_settings_teamcity_tip_shown, project_settings_workflow_tip_shown,
			       project_settings_apps_tip_shown, issue_text_recognition_tip_shown, helpdesk_channels_tip_shown, helpdesk_overview_tip_shown,
			       project_overview_page_completed, project_settings_page_completed, delayed_demo_modal_shown, saved_searches_tip_shown,
			       search_options_tip_shown, visibility_restrictions_tip_shown, agile_board_backlog_tip_shown, text_recognition_promo_shown,
			       ai_tips_shown, agile_board_page_completed, articles_page_completed, helpdesk_sla_tip_shown,
			       onboarding_tour_state, helpdesk_project_page_completed, onboarding_tour_ai_block_dismissed
			FROM user_profile_tips WHERE user_id = $1`, userID).
			Scan(
				&p.Tips.IssueListPageCompleted, &p.Tips.IssuePageCompleted, &p.Tips.HelpdeskTeamTipShown, &p.Tips.ProjectOverviewTipShown,
				&p.Tips.ProjectSettingsTipShown, &p.Tips.IssueAiActionsTipShown, &p.Tips.ProjectContentTipShown, &p.Tips.AgileBoardCardsTipShown,
				&p.Tips.AppsProjectTabTipShown, &p.Tips.ChangeRuleTypeTipShown, &p.Tips.AgileBoardVisibilityTipShown, &p.Tips.AgileBoardSwimlanesTipShown,
				&p.Tips.AgileBoardColumnsTipShown, &p.Tips.ArticleAiAssistantTipShown, &p.Tips.ArticleVisibilityTipShown, &p.Tips.ArticleInlineCommentsTipShown,
				&p.Tips.ProjectSettingsPeopleTipShown, &p.Tips.ProjectSettingsFieldsTipShown, &p.Tips.ProjectSettingsVcsTipShown, &p.Tips.HelpdeskPinnedCommentsTipShown,
				&p.Tips.PricingAdminPopupShown, &p.Tips.TextCompletionPromoShown, &p.Tips.TextRecognitionTipsShown, &p.Tips.ArticleSidebarTipShown,
				&p.Tips.ArticleCommentsTipShown, &p.Tips.ActivityTypesTipShown, &p.Tips.IssueFieldsTipShown, &p.Tips.CommandsTipShown,
				&p.Tips.TimeTrackingTipShown, &p.Tips.VisibleFieldsTipShown, &p.Tips.SurveyShown, &p.Tips.VotesTipShown,
				&p.Tips.AiPromoShown, &p.Tips.CollapsibleSidebarTipShown, &p.Tips.PmfShown, &p.Tips.AiWritingAssistantPromoShown,
				&p.Tips.InlineCommentPromoShown, &p.Tips.ProjectSettingsTimeTrackingTipShown, &p.Tips.ProjectSettingsTeamcityTipShown, &p.Tips.ProjectSettingsWorkflowTipShown,
				&p.Tips.ProjectSettingsAppsTipShown, &p.Tips.IssueTextRecognitionTipShown, &p.Tips.HelpdeskChannelsTipShown, &p.Tips.HelpdeskOverviewTipShown,
				&p.Tips.ProjectOverviewPageCompleted, &p.Tips.ProjectSettingsPageCompleted, &p.Tips.DelayedDemoModalShown, &p.Tips.SavedSearchesTipShown,
				&p.Tips.SearchOptionsTipShown, &p.Tips.VisibilityRestrictionsTipShown, &p.Tips.AgileBoardBacklogTipShown, &p.Tips.TextRecognitionPromoShown,
				&p.Tips.AiTipsShown, &p.Tips.AgileBoardPageCompleted, &p.Tips.ArticlesPageCompleted, &p.Tips.HelpdeskSlaTipShown,
				&p.Tips.OnboardingTourState, &p.Tips.HelpdeskProjectPageCompleted, &p.Tips.OnboardingTourAIBlockDismissed,
			)
	}

	// 4. جدول user_profile_notifications: يتم الاستعلام عنه فقط إذا طُلب profiles.notifications
	if tree == nil || tree.Has("profiles.notifications") {
		_ = s.db.QueryRow(ctx, `
			SELECT show_unread_only, mention_notifications_enabled, duplicate_cluster_notifications_enabled,
			       show_system, notify_on_own_changes, auto_watch_on_field_set, auto_watch_on_create,
			       auto_watch_on_comment, auto_watch_on_update, auto_watch_on_vote, email_blocked,
			       email_block_reason, email_notifications_enabled, use_plain_text_emails, disabled_direct,
			       disabled_subscription, disabled_system, mailbox_integration_notifications_enabled
			FROM user_profile_notifications WHERE user_id = $1`, userID).
			Scan(
				&p.Notifications.ShowUnreadOnly, &p.Notifications.MentionNotificationsEnabled, &p.Notifications.DuplicateClusterNotificationsEnabled,
				&p.Notifications.ShowSystem, &p.Notifications.NotifyOnOwnChanges, &p.Notifications.AutoWatchOnFieldSet, &p.Notifications.AutoWatchOnCreate,
				&p.Notifications.AutoWatchOnComment, &p.Notifications.AutoWatchOnUpdate, &p.Notifications.AutoWatchOnVote, &p.Notifications.EmailBlocked,
				&p.Notifications.EmailBlockReason, &p.Notifications.EmailNotificationsEnabled, &p.Notifications.UsePlainTextEmails, &p.Notifications.DisabledDirect,
				&p.Notifications.DisabledSubscription, &p.Notifications.DisabledSystem, &p.Notifications.MailboxIntegrationNotificationsEnabled,
			)
	}

	// 5. جدول user_profile_helpdesk و user_profile_helpdesk_projects: يتم الاستعلام عنهما فقط إذا طُلب profiles.helpdesk
	if tree == nil || tree.Has("profiles.helpdesk") {
		_ = s.db.QueryRow(ctx, `SELECT is_reporter, is_agent FROM user_profile_helpdesk WHERE user_id = $1`, userID).
			Scan(&p.Helpdesk.IsReporter, &p.Helpdesk.IsAgent)
	}
}

// GetRecentIssues يجلب المشاكل المشاهدة مؤخرًا من recent_issues و issues.
func (s *UserStore) GetRecentIssues(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentIssue, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.pinned, r.date, i.id, i.id_readable, i.summary, i.resolved, i.project_id
		FROM recent_issues r
		LEFT JOIN issues i ON r.issue_id = i.id
		WHERE r.user_id = $1
		ORDER BY r.date DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return []*model.RecentIssue{}, nil
	}
	defer rows.Close()

	recents := []*model.RecentIssue{}
	for rows.Next() {
		ri := &model.RecentIssue{Type: "RecentIssue"}
		var iss model.Issue
		var resolved *int64
		err := rows.Scan(&ri.ID, &ri.Pinned, &ri.Date, &iss.ID, &iss.IDReadable, &iss.Summary, &resolved, &iss.ProjectID)
		if err == nil {
			if resolved != nil {
				iss.Resolved = *resolved
			}
			iss.Type = "Issue"
			ri.Issue = &iss
		}
		recents = append(recents, ri)
	}
	return recents, nil
}

// GetRecentArticles يجلب المقالات المشاهدة مؤخرًا من recent_articles و articles.
func (s *UserStore) GetRecentArticles(ctx context.Context, userID string, limit int, offset int) ([]*model.RecentArticle, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.pinned, r.date, a.id, a.id_readable, a.summary, a.project_id
		FROM recent_articles r
		LEFT JOIN articles a ON r.article_id = a.id
		WHERE r.user_id = $1
		ORDER BY r.date DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return []*model.RecentArticle{}, nil
	}
	defer rows.Close()

	recents := []*model.RecentArticle{}
	for rows.Next() {
		ra := &model.RecentArticle{Type: "RecentArticle"}
		var art struct {
			ID         string `json:"id"`
			IDReadable string `json:"idReadable"`
			Summary    string `json:"summary"`
			ProjectID  string `json:"projectId"`
			Type       string `json:"$type"`
		}
		err := rows.Scan(&ra.ID, &ra.Pinned, &ra.Date, &art.ID, &art.IDReadable, &art.Summary, &art.ProjectID)
		if err == nil {
			art.Type = "Article"
			ra.Article = art
		}
		recents = append(recents, ra)
	}
	return recents, nil
}

// GetGrazieProfile يجلب إعدادات التدقيق اللغوي والذكاء الاصطناعي Grazie من user_profile_grazie.
func (s *UserStore) GetGrazieProfile(ctx context.Context, userID string) (*model.GrazieUserProfile, error) {
	g := &model.GrazieUserProfile{
		ExcludedIssueTypes:            "",
		EnableSpellChecker:            true,
		HasMoreTokens:                 true,
		EnableTextCompletion:          true,
		CycleRestart:                  1788271206218,
		FreeLicense:                   false,
		SpellCheckerEnabledInSystem:   true,
		TextCompletionEnabledInSystem: true,
		Enabled:                       false,
		Type:                          "GrazieUserProfile",
	}
	_ = s.db.QueryRow(ctx, `
		SELECT excluded_issue_types, enable_spell_checker, has_more_tokens, enable_text_completion,
		       cycle_restart, free_license, spell_checker_enabled_in_system, text_completion_enabled_in_system, enabled
		FROM user_profile_grazie WHERE user_id = $1`, userID).
		Scan(
			&g.ExcludedIssueTypes, &g.EnableSpellChecker, &g.HasMoreTokens, &g.EnableTextCompletion,
			&g.CycleRestart, &g.FreeLicense, &g.SpellCheckerEnabledInSystem, &g.TextCompletionEnabledInSystem, &g.Enabled,
		)
	return g, nil
}

// GetGeneralProfile يجلب إعدادات التاريخ والمنطقة الزمنية واللغة من user_profiles.
func (s *UserStore) GetGeneralProfile(ctx context.Context, userID string) (*model.GeneralUserProfile, error) {
	gp := &model.GeneralUserProfile{
		ID: "generalProfile",
		Timezone: &model.TimeZoneDescriptor{
			ID:   "Europe/Prague",
			Type: "TimeZoneDescriptor",
		},
		DateFormat: &model.DateFormatDescriptor{
			Pattern:     "d MMM yyyy HH:mm",
			DatePattern: "d MMM yyyy",
			Type:        "DateFormatDescriptor",
		},
		Locale: &model.LocaleDescriptor{
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
	}
	if p, err := s.GetProfile(ctx, userID); err == nil && p != nil {
		if p.TimezoneID != "" {
			gp.Timezone.ID = p.TimezoneID
		}
		if p.LocaleID != "" {
			gp.Locale.ID = p.LocaleID
		}
	}
	return gp, nil
}

// GetQuestionnaireProfile يجلب إعدادات الاستبيانات من user_profile_questionnaire.
func (s *UserStore) GetQuestionnaireProfile(ctx context.Context, userID string) (*model.QuestionnaireUserProfile, error) {
	qp := &model.QuestionnaireUserProfile{
		ShowSurvey:               false,
		ShowPmfSurvey:            false,
		DemoEligibilityTimestamp: nil,
		Type:                     "QuestionnaireUserProfile",
	}
	_ = s.db.QueryRow(ctx, `
		SELECT show_survey, show_pmf_survey, demo_eligibility_timestamp
		FROM user_profile_questionnaire WHERE user_id = $1`, userID).
		Scan(&qp.ShowSurvey, &qp.ShowPmfSurvey, &qp.DemoEligibilityTimestamp)
	return qp, nil
}

// GetHubMe يجلب ملف المستخدم في خدمة Hub.
func (s *UserStore) GetHubMe(ctx context.Context, userID string) (*model.HubUser, error) {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	hubID := u.RingID
	if hubID == "" {
		hubID = u.ID
	}
	return &model.HubUser{
		Guest: u.Guest,
		ID:    hubID,
		Name:  u.Name,
		Login: u.Login,
		Profile: &model.HubProfile{
			Email: u.Email,
			Avatar: &model.HubAvatar{
				URL:  u.AvatarURL,
				Type: "defaultAvatar",
			},
		},
		RequiredTwoFactorAuthentication: false,
		TwoFactorAuthentication:         &model.Hub2FA{Enabled: false},
		WebauthnDevice:                  &model.HubWebauthn{Enabled: false},
		Type:                            "User",
	}, nil
}

// InboxFolders يجلب مجلدات صندوق الوارد للمستخدم من inbox_folders (مطابق لـ request8.txt).
func (s *UserStore) InboxFolders(ctx context.Context, userID string) ([]*model.InboxFolder, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, last_notified, last_seen, enabled
		FROM inbox_folders
		WHERE user_id = $1
		ORDER BY id`, userID)
	if err != nil {
		return []*model.InboxFolder{}, nil
	}
	defer rows.Close()

	folders := []*model.InboxFolder{}
	for rows.Next() {
		f := &model.InboxFolder{Type: "InboxFolder"}
		if err := rows.Scan(&f.ID, &f.LastNotified, &f.LastSeen, &f.Enabled); err != nil {
			return []*model.InboxFolder{}, nil
		}
		folders = append(folders, f)
	}
	return folders, rows.Err()
}
