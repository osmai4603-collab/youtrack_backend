package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"youtrack_backend/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	query := `
		SELECT id, login, email, name, full_name, avatar_url, online, banned, is_locked, is_email_verified, guest
		FROM users
		WHERE id = $1
	`
	user := &domain.User{}
	var email, fullName, avatarURL sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&email,
		&user.Name,
		&fullName,
		&avatarURL,
		&user.Online,
		&user.Banned,
		&user.IsLocked,
		&user.IsEmailVerified,
		&user.Guest,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if fullName.Valid {
		user.FullName = fullName.String
		user.LocalizedName = fullName.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	user.Type = "User"
	return user, nil
}

func (r *UserRepository) GetByLogin(login string) (*domain.User, error) {
	query := `
		SELECT id, login, email, name, full_name, avatar_url, online, banned, is_locked, is_email_verified, guest
		FROM users
		WHERE login = $1
	`
	user := &domain.User{}
	var email, fullName, avatarURL sql.NullString

	err := r.db.QueryRow(query, login).Scan(
		&user.ID,
		&user.Login,
		&email,
		&user.Name,
		&fullName,
		&avatarURL,
		&user.Online,
		&user.Banned,
		&user.IsLocked,
		&user.IsEmailVerified,
		&user.Guest,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if fullName.Valid {
		user.FullName = fullName.String
		user.LocalizedName = fullName.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	user.Type = "User"
	return user, nil
}

// GetCurrentUser يُرجع المستخدم الحالي (يتطلب id ثابت في هذا السياق).
func (r *UserRepository) GetCurrentUser() (*domain.User, error) {
	return r.GetByID("11-556284")
}

// GetCurrentUserDetail يجلب مخطط المستخدم الحالي من قاعدة البيانات وفقاً للأقسام
// المطلوبة فقط (sel nil = كل شيء). لا تُنفَّذ استعلامات الأقسام غير المطلوبة.
// (profiles, widgets, featureFlags, userType, issueRelatedGroup) دون بيانات افتراضية.
func (r *UserRepository) GetCurrentUserDetail(id string, sel *domain.UserFieldSelect) (*domain.CurrentUser, error) {
	cu := &domain.CurrentUser{}
	cu.Type = "Me"

	var email, fullName, avatarURL, banBadge sql.NullString
	var userTypeID sql.NullString
	var userTypeName sql.NullString

	// 1) أساسيات المستخدم من جدول users مع userType
	userQuery := `
		SELECT u.id, u.login, u.email, u.name, u.full_name, u.avatar_url, u.online,
		       u.banned, u.ban_badge, u.can_read_profile, u.is_locked, u.is_email_verified,
		       u.guest, u.user_type_id, ut.name AS user_type_name
		FROM users u
		LEFT JOIN user_types ut ON ut.id = u.user_type_id
		WHERE u.id = $1
	`
	user := &domain.User{}
	err := r.db.QueryRow(userQuery, id).Scan(
		&user.ID,
		&user.Login,
		&email,
		&user.Name,
		&fullName,
		&avatarURL,
		&user.Online,
		&user.Banned,
		&banBadge,
		&user.CanReadProfile,
		&user.IsLocked,
		&user.IsEmailVerified,
		&user.Guest,
		&userTypeID,
		&userTypeName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load current user: %w", err)
	}

	cu.Widgets = []*domain.Widget{}
	cu.FeatureFlags = []*domain.FeatureFlag{}
	cu.Guest = user.Guest
	cu.Banned = user.Banned
	cu.Login = user.Login
	cu.IsLocked = user.IsLocked
	cu.IsEmailVerified = user.IsEmailVerified
	cu.CanReadProfile = user.CanReadProfile
	cu.Online = user.Online
	cu.ID = user.ID
	cu.Name = user.Name
	cu.FullName = user.FullName
	if email.Valid {
		cu.Email = email.String
	}
	if fullName.Valid {
		cu.FullName = fullName.String
	}
	if avatarURL.Valid {
		cu.AvatarURL = avatarURL.String
	}
	if banBadge.Valid {
		bb := banBadge.String
		cu.BanBadge = &bb
	}

	// 2) userType
	if userTypeID.Valid {
		cu.UserType = &domain.UserType{
			ID:   userTypeID.String,
			Type: "UserType",
		}
		if userTypeName.Valid {
			cu.UserType.Name = userTypeName.String
		}
	}

	// 3) profiles — يجلب فقط الأقسام الفرعية المطلوبة من ملف التفضيلات
	if sel == nil || sel.Profiles {
		profiles, err := r.loadCurrentUserProfiles(id, sel)
		if err != nil {
			return nil, err
		}
		cu.Profiles = profiles
	}

	// 4) issueRelatedGroup (المجموعات المباشرة للمستخدم)
	if sel == nil || sel.IssueRelatedGroup {
		group, err := r.loadIssueRelatedGroup(id)
		if err != nil {
			return nil, err
		}
		cu.IssueRelatedGroup = group
	}

	// 5) featureFlags
	if sel == nil || sel.FeatureFlags {
		ffs, err := r.loadFeatureFlags()
		if err != nil {
			return nil, err
		}
		cu.FeatureFlags = ffs
	}

	// 6) widgets
	if sel == nil || sel.Widgets {
		widgets, err := r.loadWidgets()
		if err != nil {
			return nil, err
		}
		cu.Widgets = widgets
	}

	return cu, nil
}

func (r *UserRepository) loadCurrentUserProfiles(userID string, sel *domain.UserFieldSelect) (*domain.CurrentUserProfiles, error) {
	var timezoneID, localeID, datePattern, periodFormatID sql.NullString
	var emailNotif, mentionNotif, autoWatchCreate, autoWatchComment, compactMode,
		expandNavigation, naturalCommentsOrder, useMarkdown, showSidebar,
		unresolvedOnly, ttAvailable sql.NullBool

	// الاستعلام الأساسي عن user_profiles رخيص (صف واحد) ويُحمَّل دائماً عندما
	// تكون أي من الأقسام الفرعية مطلوبة، ويغذّي كافة الأقسام أدناه.
	err := r.db.QueryRow(`
		SELECT timezone_id, locale_id, date_pattern, period_format_id,
		       email_notifications_enabled, mention_notifications_enabled,
		       auto_watch_on_create, auto_watch_on_comment, compact_mode,
		       expand_navigation, natural_comments_order, use_markdown_editor,
		       show_sidebar, unresolved_issues_only, is_time_tracking_available
		FROM user_profiles
		WHERE user_id = $1
	`, userID).Scan(
		&timezoneID, &localeID, &datePattern, &periodFormatID,
		&emailNotif, &mentionNotif, &autoWatchCreate, &autoWatchComment, &compactMode,
		&expandNavigation, &naturalCommentsOrder, &useMarkdown, &showSidebar,
		&unresolvedOnly, &ttAvailable,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to load user profiles: %w", err)
	}

	// البدء من الشكل الكامل (ديفولت request1.txt) ثم دمج قيم قاعدة البيانات فوقه،
	// لضمان خروج كافة الأقسام بالبنية الكاملة حتى عندما لا تكون مخزنة في DB.
	profiles := domain.DefaultProfiles()

	// General
	if timezoneID.Valid {
		profiles.General.Timezone = &domain.TimeZoneDescriptor{ID: timezoneID.String, Type: "TimeZoneDescriptor"}
	}
	if localeID.Valid {
		profiles.General.Locale = &domain.LocaleDescriptor{ID: localeID.String, Language: localeID.String, Locale: localeID.String, Name: localeID.String, Type: "LocaleDescriptor"}
	}
	if datePattern.Valid {
		profiles.General.DateFieldFormat = &domain.DateFormatDescriptor{Pattern: datePattern.String, DatePattern: datePattern.String, Type: "DateFormatDescriptor"}
	}

	// TimeTracking
	if periodFormatID.Valid {
		profiles.TimeTracking.PeriodFormat = &domain.PeriodFieldFormat{ID: periodFormatID.String, Type: "PeriodFieldFormat"}
	}
	if ttAvailable.Valid {
		profiles.TimeTracking.IsTimeTrackingAvailable = ttAvailable.Bool
	}

	// Notifications
	if emailNotif.Valid {
		profiles.Notifications.EmailNotificationsEnabled = emailNotif.Bool
	}
	if mentionNotif.Valid {
		profiles.Notifications.MentionNotificationsEnabled = mentionNotif.Bool
	}
	if autoWatchCreate.Valid {
		profiles.Notifications.AutoWatchOnCreate = autoWatchCreate.Bool
	}
	if autoWatchComment.Valid {
		profiles.Notifications.AutoWatchOnComment = autoWatchComment.Bool
	}

	// Appearance
	if compactMode.Valid {
		profiles.Appearance.CompactMode = compactMode.Bool
	}
	if expandNavigation.Valid {
		profiles.Appearance.ExpandNavigation = expandNavigation.Bool
	}
	if naturalCommentsOrder.Valid {
		profiles.Appearance.NaturalCommentsOrder = naturalCommentsOrder.Bool
	}
	if useMarkdown.Valid {
		profiles.Appearance.UseMarkdownEditor = useMarkdown.Bool
	}

	// IssuesList
	if showSidebar.Valid {
		profiles.IssuesList.ShowSidebar = showSidebar.Bool
	}
	if unresolvedOnly.Valid {
		profiles.IssuesList.UnresolvedIssuesOnly = unresolvedOnly.Bool
	}

	// Articles — آخر مقالة تمت زيارتها (من recent_articles) إن وُجدت
	if sel == nil || sel.Articles {
		article, err := r.loadLastVisitedArticle(userID)
		if err != nil {
			return nil, err
		}
		profiles.Articles.LastVisitedArticle = article
	}

	// General — سياقات البحث والمساعدة (من saved_queries) إن وُجدت
	if sel == nil || sel.General {
		searchCtx, err := r.loadSearchContext(userID, false)
		if err != nil {
			return nil, err
		}
		helpdeskCtx, err := r.loadSearchContext(userID, true)
		if err != nil {
			return nil, err
		}
		profiles.General.SearchContext = searchCtx
		profiles.General.HelpdeskContext = helpdeskCtx
	}

	return profiles, nil
}

// loadLastVisitedArticle يجلب آخر مقالة زارها المستخدم من recent_articles
// (مع بيانات المشروع التابعة لها). يُرجع nil عندما لا توجد أي زيارة سابقة.
func (r *UserRepository) loadLastVisitedArticle(userID string) (*domain.Article, error) {
	row := r.db.QueryRow(`
		SELECT a.id, a.id_readable, a.summary, a.ordinal, a.has_unpublished_changes,
		       a.is_updatable, a.is_deletable, a.collaborative_draft_id,
		       a.has_children, a.has_star, a.updated,
		       p.id, p.name, p.short_name
		FROM recent_articles ra
		JOIN articles a ON a.id = ra.article_id
		LEFT JOIN projects p ON p.id = a.project_id
		WHERE ra.user_id = $1
		ORDER BY ra.date DESC
		LIMIT 1
	`, userID)

	var idReadable, summary, collabDraftID, pID, pName, pShort sql.NullString
	var ordinal, updated sql.NullInt64
	var hasUnpublished, isUpdatable, isDeletable, hasChildren, hasStar sql.NullBool

	a := &domain.Article{}
	a.Type = "Article"

	err := row.Scan(
		&a.ID, &idReadable, &summary, &ordinal, &hasUnpublished,
		&isUpdatable, &isDeletable, &collabDraftID,
		&hasChildren, &hasStar, &updated,
		&pID, &pName, &pShort,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load last visited article: %w", err)
	}

	if idReadable.Valid {
		a.IDReadable = idReadable.String
	}
	if summary.Valid {
		a.Summary = summary.String
	}
	if ordinal.Valid {
		a.Ordinal = int(ordinal.Int64)
	}
	if hasUnpublished.Valid {
		a.HasUnpublishedChanges = hasUnpublished.Bool
	}
	if isUpdatable.Valid {
		a.IsUpdatable = isUpdatable.Bool
	}
	if isDeletable.Valid {
		a.IsDeletable = isDeletable.Bool
	}
	if collabDraftID.Valid {
		a.CollaborativeDraftID = collabDraftID.String
	}
	if hasChildren.Valid {
		a.HasChildren = hasChildren.Bool
	}
	if hasStar.Valid {
		a.HasStar = hasStar.Bool
	}
	if updated.Valid {
		a.Updated = updated.Int64
	}
	if pID.Valid {
		a.Project = &domain.Project{
			ID:        pID.String,
			Name:      pName.String,
			ShortName: pShort.String,
			Type:      "Project",
		}
	}
	return a, nil
}

// loadSearchContext يجلب سياق البحث أو المساعدة (Helpdesk) الخاص بالمستخدم من
// saved_queries: searchContext = أول استعلام مثبّت (pinned)، و helpdeskContext =
// أول استعلام مثبّت داخل مركز المساعدة. يُرجع nil عندما لا يوجد سياق.
func (r *UserRepository) loadSearchContext(userID string, helpdesk bool) (*domain.HelpdeskContext, error) {
	query := `
		SELECT id, name, issues_url, pinned, pinned_in_helpdesk, "query", is_updatable
		FROM saved_queries
		WHERE owner_id = $1
	`
	if helpdesk {
		query += ` AND pinned_in_helpdesk = TRUE`
	} else {
		query += ` AND pinned = TRUE`
	}
	query += ` ORDER BY id LIMIT 1`

	var name, issuesURL, q sql.NullString
	var pinned, pinnedInHelpdesk, isUpdatable sql.NullBool

	hc := &domain.HelpdeskContext{}
	hc.Type = "HelpdeskContext"

	err := r.db.QueryRow(query, userID).Scan(
		&hc.ID, &name, &issuesURL, &pinned, &pinnedInHelpdesk, &q, &isUpdatable,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load search context: %w", err)
	}

	if name.Valid {
		hc.Name = name.String
	}
	if issuesURL.Valid {
		hc.IssuesURL = issuesURL.String
	}
	if pinned.Valid {
		hc.Pinned = pinned.Bool
	}
	if pinnedInHelpdesk.Valid {
		hc.PinnedInHelpdesk = pinnedInHelpdesk.Bool
	}
	if q.Valid {
		hc.Query = q.String
	}
	if isUpdatable.Valid {
		hc.IsUpdatable = isUpdatable.Bool
	}
	return hc, nil
}

func (r *UserRepository) loadIssueRelatedGroup(userID string) (*domain.UserGroup, error) {
	row := r.db.QueryRow(`
		SELECT g.id, g.name, g.description, g.all_users_group, g.icon, g.audit_target_id,
		       g.is_updatable, g.is_removable, g.group_type,
		       p.id, p.name, p.icon_url
		FROM user_group_members m
		JOIN user_groups g ON g.id = m.group_id
		LEFT JOIN projects p ON p.id = g.team_for_project_id
		WHERE m.user_id = $1
		LIMIT 1
	`, userID)

	var name, description, icon, auditTargetID, groupType sql.NullString
	var allUsersGroup, isUpdatable, isRemovable sql.NullBool
	var teamID, teamName, teamIcon sql.NullString
	g := &domain.UserGroup{}
	err := row.Scan(
		&g.ID, &name, &description, &allUsersGroup, &icon, &auditTargetID,
		&isUpdatable, &isRemovable, &groupType,
		&teamID, &teamName, &teamIcon,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load issue related group: %w", err)
	}
	g.Type = "UserGroup"
	if name.Valid {
		g.Name = name.String
	}
	if description.Valid {
		g.Description = description.String
	}
	if icon.Valid {
		g.Icon = icon.String
	}
	if auditTargetID.Valid {
		g.AuditTargetID = auditTargetID.String
	}
	if allUsersGroup.Valid {
		g.AllUsersGroup = allUsersGroup.Bool
	}
	if isUpdatable.Valid {
		g.IsUpdatable = isUpdatable.Bool
	}
	if isRemovable.Valid {
		g.IsRemovable = isRemovable.Bool
	}
	if teamID.Valid {
		g.TeamForProject = &domain.ProjectTeam{
			ID:   teamID.String,
			Name: teamName.String,
			Icon: teamIcon.String,
			Type: "ProjectTeam",
		}
	}
	return g, nil
}

func (r *UserRepository) loadFeatureFlags() ([]*domain.FeatureFlag, error) {
	rows, err := r.db.Query(`SELECT id, enabled FROM feature_flags ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to load feature flags: %w", err)
	}
	defer rows.Close()

	ffs := []*domain.FeatureFlag{}
	for rows.Next() {
		ff := &domain.FeatureFlag{}
		if err := rows.Scan(&ff.ID, &ff.Enabled); err != nil {
			return nil, err
		}
		ff.Type = "FeatureFlag"
		ffs = append(ffs, ff)
	}
	return ffs, rows.Err()
}

func (r *UserRepository) loadWidgets() ([]*domain.Widget, error) {
	rows, err := r.db.Query(`
		SELECT id, name, key, app_id, app_name, app_title, description, extension_point,
		       icon_path, index_path, configurable, collapsed, borderless, show_header,
		       default_height, default_width, vendor_name, vendor_email, vendor_url,
		       marketplace_id, app_icon_path, app_dark_icon_path, guard
		FROM dashboard_widgets
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to load widgets: %w", err)
	}
	defer rows.Close()

	widgets := []*domain.Widget{}
	for rows.Next() {
		w := &domain.Widget{}
		var name, key, appID, appName, appTitle, description, extPoint, iconPath, indexPath, defHeight, defWidth, vendorName, vendorEmail, vendorURL, appIconPath, appDarkIconPath, guard sql.NullString
		var configurable, collapsed, borderless, showHeader sql.NullBool
		var marketplaceID sql.NullInt64
		if err := rows.Scan(
			&w.ID, &name, &key, &appID, &appName, &appTitle, &description, &extPoint,
			&iconPath, &indexPath, &configurable, &collapsed, &borderless, &showHeader,
			&defHeight, &defWidth, &vendorName, &vendorEmail, &vendorURL,
			&marketplaceID, &appIconPath, &appDarkIconPath, &guard,
		); err != nil {
			return nil, err
		}
		w.Type = "Widget"
		if name.Valid {
			w.Name = name.String
		}
		if key.Valid {
			w.Key = key.String
		}
		if appID.Valid {
			w.AppID = appID.String
		}
		if appName.Valid {
			w.AppName = appName.String
		}
		if appTitle.Valid {
			w.AppTitle = appTitle.String
		}
		if description.Valid {
			w.Description = description.String
		}
		if extPoint.Valid {
			w.ExtensionPoint = extPoint.String
		}
		if iconPath.Valid {
			w.IconPath = iconPath.String
		}
		if indexPath.Valid {
			w.IndexPath = indexPath.String
		}
		if configurable.Valid {
			w.Configurable = configurable.Bool
		}
		if collapsed.Valid {
			w.Collapsed = collapsed.Bool
		}
		if borderless.Valid {
			w.Borderless = borderless.Bool
		}
		if showHeader.Valid {
			w.ShowHeader = showHeader.Bool
		}
		if defHeight.Valid {
			w.DefaultHeight = defHeight.String
		}
		if defWidth.Valid {
			w.DefaultWidth = defWidth.String
		}
		if vendorName.Valid {
			w.VendorName = vendorName.String
		}
		if vendorEmail.Valid {
			w.VendorEmail = vendorEmail.String
		}
		if vendorURL.Valid {
			w.VendorURL = vendorURL.String
		}
		if marketplaceID.Valid {
			w.MarketplaceID = int(marketplaceID.Int64)
		}
		if appIconPath.Valid {
			w.AppIconPath = appIconPath.String
		}
		if appDarkIconPath.Valid {
			w.AppDarkIconPath = appDarkIconPath.String
		}
		if guard.Valid {
			w.Guard = guard.String
		}
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, login, email, name, full_name, avatar_url, online, banned, is_locked, is_email_verified, guest)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE 
		SET login = EXCLUDED.login,
		    email = EXCLUDED.email,
		    name = EXCLUDED.name,
		    full_name = EXCLUDED.full_name,
		    avatar_url = EXCLUDED.avatar_url,
		    online = EXCLUDED.online,
		    updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(query,
		user.ID,
		user.Login,
		user.Email,
		user.Name,
		user.FullName,
		user.AvatarURL,
		user.Online,
		user.Banned,
		user.IsLocked,
		user.IsEmailVerified,
		user.Guest,
	)
	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(user *domain.User) error {
	return r.Create(user)
}
