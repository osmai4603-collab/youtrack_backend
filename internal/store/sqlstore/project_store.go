package sqlstore

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// ProjectStore تطبيق عمليات المشاريع على PostgreSQL.
type ProjectStore struct {
	db *pgxpool.Pool
}

func scanProject(row interface{ Scan(...any) error }) (*model.Project, error) {
	p := &model.Project{}
	var projectTypeID string
	err := row.Scan(
		&p.ID, &p.Name, &p.ShortName, &projectTypeID, &p.Pinned, &p.Template, &p.Archived,
		&p.Restricted, &p.HasArticles, &p.IsDemo, &p.LeaderID, &p.OrganizationID, &p.Description,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.ProjectTypeID = projectTypeID
	p.ProjectType = &model.ProjectType{ID: projectTypeID}
	return p, nil
}

const projectSelect = `SELECT id, name, short_name, project_type_id, pinned, template, archived,
	restricted, has_articles, is_demo, leader_id, organization_id, description, created_at, updated_at
	FROM projects`

const projectDetailedSelect = `SELECT id, name, short_name, project_type_id, pinned, template, archived,
	restricted, has_articles, is_demo, leader_id, organization_id, description, created_at, updated_at,
	icon_url, source_template, fields_sorted, query, issues_url, creation_time, default_smtp, from_email,
	from_personal, reply_to_email, supports_email_delimiter, email_delimiter, use_email_delimiter,
	default_visibility_group_id, audit_target_id, historical_short_names, team_id
	FROM projects`

func (s *ProjectStore) GetByID(ctx context.Context, id string) (*model.Project, error) {
	row := s.db.QueryRow(ctx, projectSelect+` WHERE id = $1`, id)
	return scanProject(row)
}

func (s *ProjectStore) GetByShortName(ctx context.Context, shortName string) (*model.Project, error) {
	row := s.db.QueryRow(ctx, projectSelect+` WHERE short_name = $1`, shortName)
	return scanProject(row)
}

func (s *ProjectStore) All(ctx context.Context) ([]*model.Project, error) {
	rows, err := s.db.Query(ctx, projectSelect)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*model.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *ProjectStore) GetDetailed(ctx context.Context, id string, tree *fields.FieldTree) (*model.Project, error) {
	row := s.db.QueryRow(ctx, projectDetailedSelect+` WHERE id = $1 OR short_name = $1`, id)
	p := &model.Project{}
	var projectTypeID string
	var leaderID, organizationID, teamID, defaultVisibilityGroupID *string

	err := row.Scan(
		&p.ID, &p.Name, &p.ShortName, &projectTypeID, &p.Pinned, &p.Template, &p.Archived,
		&p.Restricted, &p.HasArticles, &p.IsDemo, &leaderID, &organizationID, &p.Description,
		&p.CreatedAt, &p.UpdatedAt, &p.IconURL, &p.SourceTemplate, &p.FieldsSorted, &p.Query,
		&p.IssuesURL, &p.CreationTime, &p.DefaultSmtp, &p.FromEmail, &p.FromPersonal, &p.ReplyToEmail,
		&p.SupportsEmailDelimiter, &p.EmailDelimiter, &p.UseEmailDelimiter, &defaultVisibilityGroupID,
		&p.AuditTargetID, &p.HistoricalShortNames, &teamID,
	)
	if err != nil {
		return nil, err
	}
	p.ProjectTypeID = projectTypeID
	p.ProjectType = &model.ProjectType{ID: projectTypeID}
	if leaderID != nil {
		p.LeaderID = *leaderID
	}
	if organizationID != nil {
		p.OrganizationID = *organizationID
	}

	// Leader
	if tree == nil || tree.Has("leader") {
		if p.LeaderID != "" {
			p.Leader, _ = s.getUser(ctx, p.LeaderID)
		}
	}

	// Organization
	if tree == nil || tree.Has("organization") {
		if p.OrganizationID != "" {
			p.Organization, _ = s.getOrganization(ctx, p.OrganizationID)
		}
	}

	// Team
	if tree == nil || tree.Has("team") {
		if teamID != nil {
			var child *fields.FieldTree
			if tree != nil {
				child = tree.Child("team")
			}
			p.Team, _ = s.getTeamDetailed(ctx, *teamID, child)
		}
	}

	// Widgets
	if tree == nil || tree.Has("widgets") {
		p.Widgets, _ = s.getProjectWidgets(ctx, p.ID)
	}

	// Plugins
	if tree == nil || tree.Has("plugins") {
		var child *fields.FieldTree
		if tree != nil {
			child = tree.Child("plugins")
		}
		p.Plugins, _ = s.getProjectPlugins(ctx, p.ID, child)
	}

	// Visibility Groups
	if tree == nil || tree.Has("defaultVisibilityGroup") {
		if defaultVisibilityGroupID != nil {
			var child *fields.FieldTree
			if tree != nil {
				child = tree.Child("defaultVisibilityGroup")
			}
			p.DefaultVisibilityGroup, _ = s.getUserGroup(ctx, *defaultVisibilityGroupID, child)
		}
	}
	if tree == nil || tree.Has("relevantVisibilityGroups") {
		var child *fields.FieldTree
		if tree != nil {
			child = tree.Child("relevantVisibilityGroups")
		}
		p.RelevantVisibilityGroups, _ = s.getRelevantVisibilityGroups(ctx, p.ID, child)
	}

	p.Normalize()
	return p, nil
}

// GetProjectTeamAndLeader يجلب فقط حقول القائد والفريق المطلوبة في الطلب #27 (Request #27)
// مع تقليم الأعمدة إلى ما طُلب في اجتياز حقول الـ fields. لا يلمس نقاط المشاريع الأخرى.
func (s *ProjectStore) GetProjectTeamAndLeader(ctx context.Context, id string, leaderTree, teamTree *fields.FieldTree) (*model.ProjectTeamAndLeader, error) {
	res := &model.ProjectTeamAndLeader{}

	// 1. المشروع: فقط leader_id و team_id (لا حاجة لبقية الأعمدة)
	var leaderID, teamID *string
	err := s.db.QueryRow(ctx,
		`SELECT leader_id, team_id FROM projects WHERE id = $1 OR short_name = $1`, id).
		Scan(&leaderID, &teamID)
	if err != nil {
		return nil, err
	}

	// 2. القائد: تجلب فقط الأعمدة المطلوبة داخل leader(...)
	if leaderID != nil && *leaderID != "" {
		leader := &model.User27{Type: "User"}
		// id يُطلب دائماً في الطلب #27 (leader(id))، ونضيف أي أعمدة إضافية مطلوبة.
		fields := leaderColumnsForTree(leaderTree)
		q := buildSelect("users", fields, "id = $1")
		row := s.db.QueryRow(ctx, q, *leaderID)
		if err := scanUser27(row, leader, fields); err != nil {
			return nil, err
		}
		leader.Normalize()
		res.Leader = leader
	}

	// 3. الفريق: name + الأعضاء المطلوبون داخل users(...)
	if teamID != nil && *teamID != "" {
		team := &model.Team27{Type: "ProjectTeam", Users: []*model.User27{}}
		// name قد يُطلب ضمن team(...)
		if teamTree == nil || teamTree.Has("name") {
			_ = s.db.QueryRow(ctx, `SELECT name FROM project_teams WHERE id = $1`, *teamID).Scan(&team.Name)
		}
		// users(...) -> الأعمدة المطلوبة لكل مستخدم عضو
		if teamTree == nil || teamTree.Has("users") {
			var userTree *fields.FieldTree
			if teamTree != nil {
				userTree = teamTree.Child("users")
			}
			fields := userColumnsForTree(userTree)
			q := buildSelect("u", fields, "u.id IN (SELECT user_id FROM project_team_members WHERE team_id = $1)")
			rows, err := s.db.Query(ctx, q, *teamID)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				u := &model.User27{Type: "User"}
				if err := scanUser27(rows, u, fields); err != nil {
					return nil, err
				}
				u.Normalize()
				team.Users = append(team.Users, u)
			}
			if rows.Err() != nil {
				return nil, rows.Err()
			}
		}
		team.Normalize()
		res.Team = team
	}

	res.Normalize()
	return res, nil
}

// user27ScalarColumns يحدد أعمدة جدول users المطلوبة بناءً على شجرة الحقول.
// tree في مستوى users(...) (مثلاً id,login,name,avatarUrl,email).
func userColumnsForTree(tree *fields.FieldTree) []string {
	// العمود id يُضمَّن دائماً (معرّف الكيان وسلامة الـ JSON).
	cols := []string{"id"}
	if tree == nil {
		cols = append(cols, "login", "name", "avatar_url", "email")
		return cols
	}
	if tree.Has("login") {
		cols = append(cols, "login")
	}
	if tree.Has("name") {
		cols = append(cols, "name")
	}
	if tree.Has("avatarUrl") {
		cols = append(cols, "avatar_url")
	}
	if tree.Has("email") {
		cols = append(cols, "email")
	}
	return cols
}

// leaderColumnsForTree هو نفسه userColumnsForTree لكن لمستوى leader(...) في الطلب #27.
func leaderColumnsForTree(tree *fields.FieldTree) []string {
	return userColumnsForTree(tree)
}

// buildSelect يبني عبارة SELECT عن أعمدة معينة مع جاهزية الاستعلام و WHERE.
func buildSelect(table string, cols []string, where string) string {
	q := "SELECT "
	for i, c := range cols {
		if i > 0 {
			q += ", "
		}
		q += c
	}
	q += " FROM " + table + " WHERE " + where
	return q
}

// scanUser27 يمسح صفاً إلى User27 وفق الأعمدة المطلوبة.
// يفترض أن الأعمدة المطلوبة ضمن المنظومة المعروفة؛ الأعمدة غير المطلوبة تُترك فارغة.
func scanUser27(row interface{ Scan(...any) error }, u *model.User27, cols []string) error {
	// نمرر مؤشرات للأعمدة المطلوبة فقط
	var login *string
	var name *string
	var avatarURL, email *string
	args := make([]any, 0, len(cols))
	id := &u.ID
	for _, c := range cols {
		switch c {
		case "id":
			args = append(args, id)
		case "login":
			login = new(string)
			args = append(args, login)
		case "name":
			name = new(string)
			args = append(args, name)
		case "avatar_url":
			avatarURL = new(string)
			args = append(args, avatarURL)
		case "email":
			email = new(string)
			args = append(args, email)
		default:
			// عمود غير معروف: نتجاهله في الفحص
			args = append(args, new(any))
		}
	}
	if err := row.Scan(args...); err != nil {
		return err
	}
	if login != nil {
		// login يُطلب ويُعالج كـ NOT NULL
		u.Login = *login
	}
	if name != nil {
		if *name != "" {
			u.Name = *name
		} else {
			u.Name = ""
		}
	}
	if avatarURL != nil {
		// قد تكون NULL في الجدول؛ نحتفظ بـ pointer لإخراج null
		if *avatarURL == "" {
			u.AvatarURL = nil
		} else {
			u.AvatarURL = avatarURL
		}
	}
	if email != nil {
		if *email == "" {
			u.Email = nil
		} else {
			u.Email = email
		}
	}
	return nil
}

func (s *ProjectStore) getUser(ctx context.Context, userID string) (*model.User, error) {
	row := s.db.QueryRow(ctx, `SELECT id, login, email, full_name, name, avatar_url, user_type_id, is_email_verified, guest, online, banned, ban_badge, can_read_profile, is_locked FROM users WHERE id = $1`, userID)
	u := &model.User{}
	var userTypeID, banBadge *string
	err := row.Scan(&u.ID, &u.Login, &u.Email, &u.FullName, &u.Name, &u.AvatarURL, &userTypeID, &u.IsEmailVerified, &u.Guest, &u.Online, &u.Banned, &banBadge, &u.CanReadProfile, &u.IsLocked)
	if err != nil {
		return nil, err
	}
	u.BanBadge = banBadge
	if userTypeID != nil {
		u.UserType = &model.UserType{ID: *userTypeID, Name: "Standard user"}
	}
	u.Normalize()
	return u, nil
}

func (s *ProjectStore) getOrganization(ctx context.Context, orgID string) (*model.Organization, error) {
	row := s.db.QueryRow(ctx, `SELECT id, key, name, icon_url, projects_count FROM organizations WHERE id = $1`, orgID)
	o := &model.Organization{Type: "Organization"}
	var iconURL *string
	err := row.Scan(&o.ID, &o.Key, &o.Name, &iconURL, &o.ProjectsCount)
	if err != nil {
		return nil, err
	}
	o.IconURL = iconURL
	return o, nil
}

func (s *ProjectStore) getTeamDetailed(ctx context.Context, teamID string, tree *fields.FieldTree) (*model.ProjectTeamDetailed, error) {
	row := s.db.QueryRow(ctx, `SELECT id, name, COALESCE(description,''), COALESCE(icon,''), COALESCE(audit_target_id,''), all_users_group, is_updatable, is_removable, project_id FROM project_teams WHERE id = $1`, teamID)
	t := &model.ProjectTeamDetailed{Type: "ProjectTeam"}
	var projectID string
	err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Icon, &t.AuditTargetID, &t.AllUsersGroup, &t.IsUpdatable, &t.IsRemovable, &projectID)
	if err != nil {
		return nil, err
	}

	if tree == nil || tree.Has("teamForProject") {
		var pName, pIcon string
		_ = s.db.QueryRow(ctx, `SELECT name, icon_url FROM projects WHERE id = $1`, projectID).Scan(&pName, &pIcon)
		t.TeamForProject = &model.ProjectRef{ID: projectID, Name: pName, Icon: pIcon, Type: "Project"}
	}

	if tree == nil || tree.Has("users") {
		rows, err := s.db.Query(ctx, `
			SELECT u.id, u.login, u.email, u.full_name, u.name, u.avatar_url, u.user_type_id, u.is_email_verified, u.guest, u.online, u.banned, u.ban_badge, u.can_read_profile, u.is_locked
			FROM users u
			JOIN project_team_members m ON u.id = m.user_id
			WHERE m.team_id = $1`, teamID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				u := &model.User{}
				var utid, bb *string
				_ = rows.Scan(&u.ID, &u.Login, &u.Email, &u.FullName, &u.Name, &u.AvatarURL, &utid, &u.IsEmailVerified, &u.Guest, &u.Online, &u.Banned, &bb, &u.CanReadProfile, &u.IsLocked)
				u.BanBadge = bb
				u.Normalize()
				t.Users = append(t.Users, u)
			}
		}
	}

	return t, nil
}

func (s *ProjectStore) getProjectWidgets(ctx context.Context, projectID string) ([]*model.DashboardWidget, error) {
	// في المرحلة الحالية، نستخدم الودجات العامة أو نربطها بالمشروع إذا كان هناك جدول ربط
	// حسب init_schema يوجد جدول project_dashboard_widgets
	rows, err := s.db.Query(ctx, `
		SELECT w.id, COALESCE(w.key,''), COALESCE(w.app_id,''), COALESCE(w.description,''), COALESCE(w.app_name,''),
		       COALESCE(w.app_title,''), COALESCE(w.name,''), w.collapsed, w.configurable, COALESCE(w.index_path,''),
		       COALESCE(w.extension_point,''), COALESCE(w.icon_path,''), COALESCE(w.guard,''), COALESCE(w.app_icon_path,''),
		       COALESCE(w.app_dark_icon_path,''), w.default_height, w.default_width, w.expected_height, w.expected_width,
		       COALESCE(w.vendor_name,''), COALESCE(w.vendor_email,''), COALESCE(w.vendor_url,''), w.marketplace_id,
		       w.show_header, w.borderless
		FROM dashboard_widgets w
		JOIN project_dashboard_widgets pw ON w.id = pw.widget_id
		WHERE pw.project_id = $1`, projectID)
	if err != nil {
		return []*model.DashboardWidget{}, nil
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
		if err == nil {
			widgets = append(widgets, w)
		}
	}
	return widgets, nil
}

func (s *ProjectStore) getProjectPlugins(ctx context.Context, projectID string, tree *fields.FieldTree) (*model.ProjectPlugins, error) {
	if tree != nil && tree.IsEmpty() {
		return &model.ProjectPlugins{Type: "ProjectPlugins"}, nil
	}
	plugins := &model.ProjectPlugins{Type: "ProjectPlugins"}

	// TimeTracking
	if tree == nil || tree.Has("timeTrackingSettings") {
		var ttEnabled bool
		var ttID string
		err := s.db.QueryRow(ctx, `SELECT id, enabled FROM project_time_tracking_settings WHERE project_id = $1`, projectID).Scan(&ttID, &ttEnabled)
		if err == nil {
			plugins.TimeTrackingSettings = &model.TimeTrackingSettings{ID: ttID, Enabled: ttEnabled, Type: "ProjectTimeTrackingSettings"}
		}
	}

	// HelpDesk
	if tree == nil || tree.Has("helpDeskSettings") {
		g := &model.HelpDeskSettings{Type: "ProjectHelpDeskSettings"}
		var uuid, title string
		err := s.db.QueryRow(ctx, `SELECT id, COALESCE(default_form_uuid,''), COALESCE(default_form_title,'') FROM project_helpdesk_settings WHERE project_id = $1`, projectID).Scan(&g.ID, &uuid, &title)
		if err == nil {
			if uuid != "" || title != "" {
				g.DefaultForm = &model.HelpDeskForm{UUID: uuid, Title: title}
			}
		}
		plugins.HelpDeskSettings = g
	}

	// VCS
	if tree == nil || tree.Has("vcsIntegrationSettings") {
		var hasVcs bool
		_ = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vcs_hosting_servers) LIMIT 1`).Scan(&hasVcs)
		plugins.VcsIntegrationSettings = &model.VcsIntegrationSettings{HasVcsIntegrations: hasVcs, Type: "ProjectVcsIntegrationSettings"}
	}

	// Grazie
	if tree == nil || tree.Has("grazie") {
		var disabled bool
		_ = s.db.QueryRow(ctx, `SELECT disabled FROM project_grazie_settings WHERE project_id = $1`, projectID).Scan(&disabled)
		plugins.Grazie = &model.GrazieSettings{Disabled: disabled, Type: "ProjectGraziePlugin"}
	}

	return plugins, nil
}

func (s *ProjectStore) getUserGroup(ctx context.Context, groupID string, tree *fields.FieldTree) (*model.UserGroup, error) {
	row := s.db.QueryRow(ctx, `SELECT id, name, COALESCE(group_type,''), all_users_group, COALESCE(icon,''), COALESCE(description,''), COALESCE(audit_target_id,''), is_updatable, is_removable, COALESCE(team_for_project_id,'') FROM user_groups WHERE id = $1`, groupID)
	g := &model.UserGroup{Type: "UserGroup"}
	var teamForProjectID string
	err := row.Scan(&g.ID, &g.Name, &g.GroupType, &g.AllUsersGroup, &g.Icon, &g.Description, &g.AuditTargetID, &g.IsUpdatable, &g.IsRemovable, &teamForProjectID)
	if err != nil {
		return nil, err
	}
	if tree == nil || tree.Has("teamForProject") {
		if teamForProjectID != "" {
			var pName, pIcon string
			_ = s.db.QueryRow(ctx, `SELECT COALESCE(name,''), COALESCE(icon_url,'') FROM projects WHERE id = $1`, teamForProjectID).Scan(&pName, &pIcon)
			g.TeamForProject = &model.ProjectRef{ID: teamForProjectID, Name: pName, Icon: pIcon, Type: "Project"}
		}
	}
	return g, nil
}

func (s *ProjectStore) getRelevantVisibilityGroups(ctx context.Context, projectID string, tree *fields.FieldTree) ([]*model.UserGroup, error) {
	// نستعلم أولاً من جدول الربط الجديد project_visibility_groups
	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.name, COALESCE(g.group_type,''), g.all_users_group, COALESCE(g.icon,''), COALESCE(g.description,''), COALESCE(g.audit_target_id,''), g.is_updatable, g.is_removable, COALESCE(g.team_for_project_id,'')
		FROM project_visibility_groups pvg
		JOIN user_groups g ON g.id = pvg.group_id
		WHERE pvg.project_id = $1`, projectID)
	if err != nil {
		// احتياط: قواعد ضمنية قديمة
		rows, err = s.db.Query(ctx, `
			SELECT g.id, g.name, COALESCE(g.group_type,''), g.all_users_group, COALESCE(g.icon,''), COALESCE(g.description,''), COALESCE(g.audit_target_id,''), g.is_updatable, g.is_removable, COALESCE(g.team_for_project_id,'')
			FROM user_groups g
			WHERE g.team_for_project_id = $1 OR g.all_users_group = true`, projectID)
		if err != nil {
			return []*model.UserGroup{}, nil
		}
	}
	defer rows.Close()

	fetchTeam := tree == nil || tree.Has("teamForProject")

	groups := []*model.UserGroup{}
	for rows.Next() {
		g := &model.UserGroup{Type: "UserGroup"}
		var teamForProjectID string
		err := rows.Scan(&g.ID, &g.Name, &g.GroupType, &g.AllUsersGroup, &g.Icon, &g.Description, &g.AuditTargetID, &g.IsUpdatable, &g.IsRemovable, &teamForProjectID)
		if err == nil {
			if fetchTeam && teamForProjectID != "" {
				var pName, pIcon string
				_ = s.db.QueryRow(ctx, `SELECT COALESCE(name,''), COALESCE(icon_url,'') FROM projects WHERE id = $1`, teamForProjectID).Scan(&pName, &pIcon)
				g.TeamForProject = &model.ProjectRef{ID: teamForProjectID, Name: pName, Icon: pIcon, Type: "Project"}
			}
			groups = append(groups, g)
		}
	}
	return groups, nil
}
