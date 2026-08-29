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
			p.Team, _ = s.getTeamDetailed(ctx, *teamID, tree.Child("team"))
		}
	}

	// Widgets
	if tree == nil || tree.Has("widgets") {
		p.Widgets, _ = s.getProjectWidgets(ctx, p.ID)
	}

	// Plugins
	if tree == nil || tree.Has("plugins") {
		p.Plugins, _ = s.getProjectPlugins(ctx, p.ID)
	}

	// Visibility Groups
	if tree == nil || tree.Has("defaultVisibilityGroup") {
		if defaultVisibilityGroupID != nil {
			p.DefaultVisibilityGroup, _ = s.getUserGroup(ctx, *defaultVisibilityGroupID)
		}
	}
	if tree == nil || tree.Has("relevantVisibilityGroups") {
		p.RelevantVisibilityGroups, _ = s.getRelevantVisibilityGroups(ctx, p.ID)
	}

	p.Normalize()
	return p, nil
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
	err := row.Scan(&o.ID, &o.Key, &o.Name, &o.IconURL, &o.ProjectsCount)
	if err != nil {
		return nil, err
	}
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

func (s *ProjectStore) getProjectPlugins(ctx context.Context, projectID string) (*model.ProjectPlugins, error) {
	plugins := &model.ProjectPlugins{Type: "ProjectPlugins"}

	// TimeTracking
	var ttEnabled bool
	var ttID string
	err := s.db.QueryRow(ctx, `SELECT id, enabled FROM project_time_tracking_settings WHERE project_id = $1`, projectID).Scan(&ttID, &ttEnabled)
	if err == nil {
		plugins.TimeTrackingSettings = &model.TimeTrackingSettings{ID: ttID, Enabled: ttEnabled, Type: "TimeTrackingSettings"}
	}

	// HelpDesk
	// نفترض وجود جدول للإعدادات أو استخراجها من app_configurations
	plugins.HelpDeskSettings = &model.HelpDeskSettings{ID: "helpdesk", Type: "HelpDeskSettings"}

	// VCS
	var hasVcs bool
	_ = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vcs_hosting_servers) LIMIT 1`).Scan(&hasVcs)
	plugins.VcsIntegrationSettings = &model.VcsIntegrationSettings{HasVcsIntegrations: hasVcs, Type: "VcsIntegrationSettings"}

	// Grazie
	plugins.Grazie = &model.GrazieSettings{Disabled: false, Type: "GrazieSettings"}

	return plugins, nil
}

func (s *ProjectStore) getUserGroup(ctx context.Context, groupID string) (*model.UserGroup, error) {
	row := s.db.QueryRow(ctx, `SELECT id, name, COALESCE(group_type,''), all_users_group, COALESCE(icon,''), COALESCE(description,''), COALESCE(audit_target_id,''), is_updatable, is_removable FROM user_groups WHERE id = $1`, groupID)
	g := &model.UserGroup{Type: "UserGroup"}
	err := row.Scan(&g.ID, &g.Name, &g.GroupType, &g.AllUsersGroup, &g.Icon, &g.Description, &g.AuditTargetID, &g.IsUpdatable, &g.IsRemovable)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *ProjectStore) getRelevantVisibilityGroups(ctx context.Context, projectID string) ([]*model.UserGroup, error) {
	// نفترض أن المجموعات ذات الصلة هي تلك التي لها أدوار في المشروع
	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.name, COALESCE(g.group_type,''), g.all_users_group, COALESCE(g.icon,''), COALESCE(g.description,''), COALESCE(g.audit_target_id,''), g.is_updatable, g.is_removable
		FROM user_groups g
		WHERE g.team_for_project_id = $1 OR g.all_users_group = true`, projectID)
	if err != nil {
		return []*model.UserGroup{}, nil
	}
	defer rows.Close()

	groups := []*model.UserGroup{}
	for rows.Next() {
		g := &model.UserGroup{Type: "UserGroup"}
		err := rows.Scan(&g.ID, &g.Name, &g.GroupType, &g.AllUsersGroup, &g.Icon, &g.Description, &g.AuditTargetID, &g.IsUpdatable, &g.IsRemovable)
		if err == nil {
			groups = append(groups, g)
		}
	}
	return groups, nil
}
