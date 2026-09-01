package sqlstore

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// newID يولّد معرّفًا عشوائيًا فريدًا (للكيانات الجديدة).
func newID() string {
	return uuid.NewString()
}

// AdminStore تطبيق عمليات الإدارة والميتاداتا على PostgreSQL.
type AdminStore struct {
	db *pgxpool.Pool
}

func (s *AdminStore) Roles(ctx context.Context) ([]*model.Role, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, description, is_updatable, immutable FROM roles`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []*model.Role{}
	for rows.Next() {
		r := &model.Role{}
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsUpdatable, &r.Immutable); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// تحميل الصلاحيات لكل دور
	for _, r := range roles {
		perms, err := s.rolePermissions(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		r.Permissions = perms
	}
	return roles, nil
}

func (s *AdminStore) rolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.name, p.description, p.permission_entity_type, p.localized_permission_entity_type, p.operation, p.is_global
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*model.Permission{}
	for rows.Next() {
		p := &model.Permission{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PermissionEntityType,
			&p.LocalizedPermissionEntityType, &p.Operation, &p.IsGlobal); err != nil {
			return nil, err
		}
		p.Type = "Permission"
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (s *AdminStore) Permissions(ctx context.Context) ([]*model.Permission, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, description, permission_entity_type, localized_permission_entity_type, operation, is_global
		FROM permissions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*model.Permission{}
	for rows.Next() {
		p := &model.Permission{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PermissionEntityType,
			&p.LocalizedPermissionEntityType, &p.Operation, &p.IsGlobal); err != nil {
			return nil, err
		}
		p.Type = "Permission"
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

// Widgets يعيد قائمة الودجات العامة (مطابق لـ request6.txt).
func (s *AdminStore) Widgets(ctx context.Context) ([]*model.DashboardWidget, error) {
	return loadDashboardWidgets(ctx, s.db)
}

func (s *AdminStore) GlobalSettings(ctx context.Context) (*model.GlobalSettings, error) {
	g := &model.GlobalSettings{}
	err := s.db.QueryRow(ctx, `
		SELECT id, version, build, read_only, helpdesk_enabled, max_upload_file_size, max_export_items
		FROM global_settings ORDER BY id LIMIT 1`).
		Scan(&g.ID, &g.Version, &g.Build, &g.ReadOnly, &g.HelpdeskEnabled, &g.MaxUploadFileSize, &g.MaxExportItems)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// AdminGlobalSettings يعيد إعدادات GET /api/admin/globalSettings من جدول global_settings
// عبر المخطط الجديد المستقل AdminGlobalSettings، مع الجلب الانتقائي الدقيق حسب شجرة
// الحقول FieldTree (مطابق لـ request5.txt و request19.txt و request46.txt).
func (s *AdminStore) AdminGlobalSettings(ctx context.Context, tree *fields.FieldTree) (*model.AdminGlobalSettings, error) {
	var id int
	var allowAllOrigins bool
	var itrEnabled bool
	var ocrSupported string
	var emailEnabled bool
	var emailIsDefault bool
	err := s.db.QueryRow(ctx, `
		SELECT id, allow_all_origins, image_text_recognition_enabled, ocr_supported,
		       email_settings_enabled, email_settings_is_default
		FROM global_settings ORDER BY id LIMIT 1`).
		Scan(&id, &allowAllOrigins, &itrEnabled, &ocrSupported, &emailEnabled, &emailIsDefault)
	if err != nil {
		return nil, err
	}

	g := &model.AdminGlobalSettings{Type: "GlobalSettings"}
	if tree == nil || tree.Has("restSettings") {
		allowedOrigins, err := s.globalSettingsAllowedOrigins(ctx, id)
		if err != nil {
			return nil, err
		}
		g.RestSettings = &model.RestCorsSettings{
			AllowAllOrigins: allowAllOrigins,
			AllowedOrigins:  allowedOrigins,
			Type:            "RestCorsSettings",
		}
	}
	if tree == nil || tree.Has("imageTextRecognitionSettings") {
		g.ImageTextRecognitionSettings = &model.ImageTextRecognitionSettings{
			Enabled: itrEnabled,
			Type:    "ImageTextRecognitionSettings",
		}
	}
	if tree == nil || tree.Has("systemSettings") {
		g.SystemSettings = &model.SystemSettings{
			OcrSupported: ocrSupported,
			Type:         "SystemSettings",
		}
	}
	if tree == nil || tree.Has("notificationSettings") {
		g.NotificationSettings = &model.NotificationSettings{
			EmailSettings: &model.EmailSettings{
				IsEnabled: emailEnabled,
				IsDefault: emailIsDefault,
				Type:      "EmailSettings",
			},
			Type: "NotificationSettings",
		}
	}
	return g, nil
}

// globalSettingsAllowedOrigins يعيد قائمة النطاقات المسموح بها (allowedOrigins)
// من جدول global_settings_allowed_origins للصف المحدد، أو مصفوفة فارغة إن لم توجد
// (مطابقة لقيمة YouTrack الافتراضية []).
func (s *AdminStore) globalSettingsAllowedOrigins(ctx context.Context, settingsID int) ([]string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT origin FROM global_settings_allowed_origins
		WHERE settings_id = $1
		ORDER BY id`, settingsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	origins := []string{}
	for rows.Next() {
		var origin string
		if err := rows.Scan(&origin); err != nil {
			return nil, err
		}
		origins = append(origins, origin)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return origins, nil
}

// BannersConfig يعيد إعدادات اللافتات الخاص بـ GET /api/config عبر المخطط الجديد
// المستقل BannersConfig، مع الجلب الانتقائي الدقيق حسب شجرة الحقول FieldTree
// (مطابق لـ request11.txt). يُستعلم عن جدول النظام_الفرعي system_events_banners
// فقط عند طلب الحقل systemEventsBanners، وتُخيَّر أعمدة جدول banners_config
// حسب الحقول المطلوبة حصراً (مثل طلب جلب بيانات المستخدم الحالي).
func (s *AdminStore) BannersConfig(ctx context.Context, tree *fields.FieldTree) (*model.BannersConfig, error) {
	b := &model.BannersConfig{Type: "BannersConfig"}

	// شجرة الحقول الفرعية الخاصة بكائن banners
	var bannerTree *fields.FieldTree
	if tree != nil {
		bannerTree = tree.Child("banners")
	}

	wantGlobal := bannerTree == nil || bannerTree.IsEmpty() || bannerTree.Has("globalBanner")
	wantEnabled := bannerTree == nil || bannerTree.IsEmpty() || bannerTree.Has("globalBannerEnabled")
	wantSystemEvents := bannerTree == nil || bannerTree.IsEmpty() || bannerTree.Has("systemEventsBanners")

	// 1. قراءة صف banners_config بأعمدة ديناميكية تطابق الحقول المطلوبة فقط
	if wantGlobal || wantEnabled {
		if err := s.loadBannersConfigRow(ctx, b, wantGlobal, wantEnabled); err != nil {
			return nil, err
		}
	}

	// 2. قراءة جدول system_events_banners فقط إذا طُلب الحقل
	if wantSystemEvents {
		events, err := s.systemEventsBanners(ctx)
		if err != nil {
			return nil, err
		}
		b.SystemEventsBanners = events
	}

	return b, nil
}

// loadBannersConfigRow يقرأ صف banners_config مع الأعمدة المطلوبة فقط
// (SELECT ديناميكي)، لضمان تطابق الأعمدة الجلوبة مع حقول معامل fields.
func (s *AdminStore) loadBannersConfigRow(ctx context.Context, b *model.BannersConfig, wantGlobal, wantEnabled bool) error {
	cols := make([]string, 0, 2)
	dests := make([]any, 0, 2)

	var globalBanner string
	var globalBannerEnabled bool
	if wantGlobal {
		cols = append(cols, "global_banner")
		dests = append(dests, &globalBanner)
	}
	if wantEnabled {
		cols = append(cols, "global_banner_enabled")
		dests = append(dests, &globalBannerEnabled)
	}

	err := s.db.QueryRow(ctx,
		`SELECT `+strings.Join(cols, ", ")+` FROM banners_config ORDER BY id LIMIT 1`).
		Scan(dests...)
	if err != nil {
		return err
	}

	if wantGlobal {
		b.GlobalBanner = globalBanner
	}
	if wantEnabled {
		b.GlobalBannerEnabled = globalBannerEnabled
	}
	return nil
}

// systemEventsBanners يعيد لافتات أحداث النظام المخزنة (فارغة إن لم توجد).
func (s *AdminStore) systemEventsBanners(ctx context.Context) ([]*model.SystemEventsBanner, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, COALESCE(text, '')
		FROM system_events_banners
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*model.SystemEventsBanner{}
	for rows.Next() {
		e := &model.SystemEventsBanner{Type: "SystemEventsBanner"}
		if err := rows.Scan(&e.ID, &e.Text); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// ProjectDashboard يعيد لوحة ودجات المشروع عبر المخطط الجديد المستقل ProjectDashboard
// مع الجلب الانتقائي الدقيق حسب شجرة الحقول FieldTree (مطابق لـ request13.txt و request23.txt).
// يُحَل المشروع بواسطة id أو short_name؛ وإن لم يوجد يُعهد خطأ pgx.ErrNoRows.
func (s *AdminStore) ProjectDashboard(ctx context.Context, projectKey string, tree *fields.FieldTree) (*model.ProjectDashboard, error) {
	var projectID string
	if err := s.db.QueryRow(ctx,
		`SELECT id FROM projects WHERE id = $1 OR short_name = $1`, projectKey).Scan(&projectID); err != nil {
		return nil, err
	}

	dashboard := &model.ProjectDashboard{
		Widgets: []*model.ProjectDashboardWidget{},
		Type:    "ProjectDashboard",
	}

	// عندما لا تُطلب widgets في معامل fields نعرض اللوحة فارغة دون أي استعلام.
	var widgetTree *fields.FieldTree
	if tree != nil && !tree.IsEmpty() {
		widgetTree = tree.Child("widgets")
		if widgetTree == nil {
			return dashboard, nil
		}
	}

	// أعمدة ديناميكية تطابق الحقول المطلوبة فقط (مثل جلب المستخدم الحالي).
	fieldIDs, fieldExprs := projectDashboardWidgetFields(widgetTree)
	query := `SELECT ` + strings.Join(fieldExprs, ", ") + `
		FROM project_dashboard_widgets pdw
		LEFT JOIN dashboard_widgets dw ON dw.id = pdw.widget_id
		WHERE pdw.project_id = $1
		ORDER BY pdw.y, pdw.x, pdw.id`

	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		w := &model.ProjectDashboardWidget{Type: "ProjectDashboardWidget"}
		dests := make([]any, 0, len(fieldIDs))
		for _, fid := range fieldIDs {
			switch fid {
			case "id":
				dests = append(dests, &w.ID)
			case "key":
				dests = append(dests, &w.Key)
			case "x":
				dests = append(dests, &w.X)
			case "y":
				dests = append(dests, &w.Y)
			case "width":
				dests = append(dests, &w.Width)
			case "height":
				dests = append(dests, &w.Height)
			case "settings":
				dests = append(dests, &w.Settings)
			case "widget_id":
				if w.Widget == nil {
					w.Widget = &model.DashboardWidget{Type: "WidgetView"}
				}
				dests = append(dests, &w.Widget.ID)
			}
		}
		if err := rows.Scan(dests...); err != nil {
			return nil, err
		}
		// بدون ودجت مرتبطة (LEFT JOIN فارغ) لا نضمّن الكائن المتداخل.
		if w.Widget != nil && w.Widget.ID == "" {
			w.Widget = nil
		}
		dashboard.Widgets = append(dashboard.Widgets, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dashboard, nil
}

// projectDashboardWidgetFields يختار أعمدة جدول project_dashboard_widgets المطلوبة
// حسب شجرة الحقول؛ يُرجع معرّفات الأعمدة مع تعبيرات SQL المناظرة.
// عندما تكون الشجرة فارغة (أو لم تُحدد) تُرجع كل الحقول.
func projectDashboardWidgetFields(tree *fields.FieldTree) (fieldIDs, fieldExprs []string) {
	allIDs := []string{"id", "key", "x", "y", "width", "height", "settings", "widget_id"}
	allExprs := []string{
		"pdw.id",
		"pdw.key",
		"pdw.x",
		"pdw.y",
		"pdw.width",
		"pdw.height",
		"COALESCE(pdw.settings, '')",
		"COALESCE(dw.id, '')",
	}

	if tree == nil || tree.IsEmpty() {
		return allIDs, allExprs
	}

	for i, fid := range allIDs {
		if fid == "widget_id" {
			if tree.Has("widget") {
				fieldIDs = append(fieldIDs, fid)
				fieldExprs = append(fieldExprs, allExprs[i])
			}
			continue
		}
		if tree.Has(fid) {
			fieldIDs = append(fieldIDs, fid)
			fieldExprs = append(fieldExprs, allExprs[i])
		}
	}
	if len(fieldIDs) == 0 {
		return allIDs, allExprs
	}
	return fieldIDs, fieldExprs
}

// Organizations يعيد قائمة المنظمات مع دعم الفرز (sorting) والحدود ($top و $skip)
// والجلب الانتقائي الدقيق حسب شجرة الحقول FieldTree (مطابق لـ request22.txt و hh.json).
// عند طلب الحقل projects يتم جلب المشاريع المرتبطة بكل منظمة مع دعم الحقول الفرعية
// (projectType, team, ...).
func (s *AdminStore) Organizations(ctx context.Context, tree *fields.FieldTree, top int, skip int, sorting string) ([]*model.Organization, error) {
	order := "name"
	switch sorting {
	case "asc", "natural":
		order = "name"
	case "desc":
		order = "name DESC"
	}

	query := `SELECT id, COALESCE(key,''), COALESCE(name,''), icon_url, projects_count,
		COALESCE(audit_target_id,''), COALESCE(description,'')
		FROM organizations
		ORDER BY ` + order
	if top > 0 {
		query += " LIMIT $1 OFFSET $2"
	}

	orgs := []*model.Organization{}
	var err error
	var rows pgx.Rows
	if top > 0 {
		rows, err = s.db.Query(ctx, query, top, skip)
	} else {
		rows, err = s.db.Query(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		o := &model.Organization{Type: "Organization"}
		var iconURL *string
		if err := rows.Scan(&o.ID, &o.Key, &o.Name, &iconURL, &o.ProjectsCount, &o.AuditTargetID, &o.Description); err != nil {
			return nil, err
		}
		o.IconURL = iconURL
		orgs = append(orgs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// جلب المشاريع المرتبطة فقط عند طلب الحقل projects في شجرة الحقول.
	if tree != nil && !tree.IsEmpty() && tree.Has("projects") {
		projectTree := tree.Child("projects")
		for _, o := range orgs {
			projects, err := s.organizationProjects(ctx, o.ID, projectTree)
			if err != nil {
				return nil, err
			}
			o.Projects = projects
		}
	}

	return orgs, nil
}

// organizationProjects يعيد مشاريع منظمة محددة مع دعم الحقول الفرعية حسب شجرة
// الحقول (projectType, team, ...) وتطبيع $type على كل مستوى.
func (s *AdminStore) organizationProjects(ctx context.Context, orgID string, tree *fields.FieldTree) ([]*model.Project, error) {
	wantProjectType := tree == nil || tree.IsEmpty() || tree.Has("projectType") || tree.Has("projectType.id")
	wantTeam := tree == nil || tree.IsEmpty() || tree.Has("team") || tree.Has("team.id")

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.name, p.short_name, p.pinned, COALESCE(p.icon_url, ''), p.template, p.archived,
		       p.restricted, p.has_articles, COALESCE(pt.id, 'DEFAULT'), COALESCE(p.team_id, '')
		FROM projects p
		LEFT JOIN project_types pt ON pt.id = p.project_type_id
		WHERE p.organization_id = $1
		ORDER BY p.short_name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*model.Project{}
	for rows.Next() {
		p := &model.Project{}
		var projectTypeID, teamID string
		if err := rows.Scan(&p.ID, &p.Name, &p.ShortName, &p.Pinned, &p.IconURL, &p.Template,
			&p.Archived, &p.Restricted, &p.HasArticles, &projectTypeID, &teamID); err != nil {
			return nil, err
		}
		p.Type = "Project"
		if wantProjectType {
			p.ProjectType = &model.ProjectType{ID: projectTypeID, Type: "ProjectType"}
		}
		if wantTeam && teamID != "" {
			p.Team = &model.ProjectTeamDetailed{ID: teamID, Type: "ProjectTeam"}
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// CachedPermissions يعيد الصلاحيات المخبأة للمستخدم مع مشاريعها.
func (s *AdminStore) CachedPermissions(ctx context.Context, userID string) ([]*model.CachedPermission, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, is_global
		FROM cached_permissions
		WHERE user_id = $1
		ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]*model.CachedPermission{}
	var ids []string
	for rows.Next() {
		cp := &model.CachedPermission{Type: "CachedPermission"}
		if err := rows.Scan(&cp.ID, &cp.Global); err != nil {
			return nil, err
		}
		// الصلاحيات العامة بلا نطاقات، وغير العامة بمصفوفات نطاقات (كما في request7).
		if cp.Global {
			cp.Projects = nil
			cp.Organizations = nil
		} else {
			cp.Projects = []*model.CachedPermissionProject{}
			cp.Organizations = []*model.Organization{}
		}
		byID[cp.ID] = cp
		ids = append(ids, cp.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		pRows, err := s.db.Query(ctx, `
			SELECT cpp.permission_id, p.id, COALESCE(pt.id, 'DEFAULT')
			FROM cached_permission_projects cpp
			JOIN cached_permissions c ON c.id = cpp.permission_id AND c.user_id = $2
			JOIN projects p ON p.id = cpp.project_id
			LEFT JOIN project_types pt ON pt.id = p.project_type_id
			WHERE cpp.permission_id = ANY($1)
			ORDER BY cpp.permission_id, p.id`, ids, userID)
		if err != nil {
			return nil, err
		}
		defer pRows.Close()

		for pRows.Next() {
			var permID, projectID, projectTypeID string
			if err := pRows.Scan(&permID, &projectID, &projectTypeID); err != nil {
				return nil, err
			}
			if cp, ok := byID[permID]; ok {
				cp.Projects = append(cp.Projects, &model.CachedPermissionProject{
					ID:          projectID,
					ProjectType: &model.ProjectType{ID: projectTypeID, Type: "ProjectType"},
					Type:        "Project",
				})
			}
		}
		if err := pRows.Err(); err != nil {
			return nil, err
		}
	}

	result := make([]*model.CachedPermission, 0, len(ids))
	for _, id := range ids {
		result = append(result, byID[id])
	}
	return result, nil
}

// PermissionsCache يعيد الصلاحيات المخبأة للمستخدم عبر المخطط الجديد المستقل
// PermissionCacheEntry مع الجلب الانتقائي الدقيق حسب شجرة الحقول FieldTree
// (مطابق لـ request20.txt). يُستعلم عن جدول cached_permission_projects وجداول
// projects و project_types فقط عند طلب الحقل projects (ومشروعه الفرعي المشترك).
func (s *AdminStore) PermissionsCache(ctx context.Context, userID string, tree *fields.FieldTree) ([]*model.PermissionCacheEntry, error) {
	wantPermission := tree != nil && tree.Has("permission")
	var rows pgx.Rows
	var err error

	if wantPermission {
		rows, err = s.db.Query(ctx, `
			SELECT cp.id, cp.is_global, p.name
			FROM cached_permissions cp
			LEFT JOIN permissions p ON p.id = cp.id
			WHERE cp.user_id = $1
			ORDER BY cp.id`, userID)
	} else {
		rows, err = s.db.Query(ctx, `
			SELECT id, is_global
			FROM cached_permissions
			WHERE user_id = $1
			ORDER BY id`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]*model.PermissionCacheEntry{}
	var ids []string
	for rows.Next() {
		cp := &model.PermissionCacheEntry{Type: "CachedPermission"}
		var global bool
		if wantPermission {
			var name *string
			if err := rows.Scan(&cp.ID, &global, &name); err != nil {
				return nil, err
			}
			if name != nil {
				cp.PermissionName = *name
			}
		} else {
			if err := rows.Scan(&cp.ID, &global); err != nil {
				return nil, err
			}
		}
		cp.Global = &global
		// الصلاحيات العامة بلا نطاقات، وغير العامة بمصفوفات نطاقات (كما في request20).
		if global {
			cp.Projects = nil
			cp.Organizations = nil
		} else {
			cp.Projects = []*model.PermissionCacheProject{}
			cp.Organizations = []*model.PermissionCacheOrganization{}
		}
		byID[cp.ID] = cp
		ids = append(ids, cp.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) > 0 && (tree == nil || tree.IsEmpty() || tree.Has("projects")) {
		projectTree := tree.Child("projects")
		wantProjectType := projectTree == nil || projectTree.IsEmpty() || projectTree.Has("projectType")
		wantID := projectTree == nil || projectTree.IsEmpty() || projectTree.Has("id")

		var projectExprs []string
		if wantID {
			projectExprs = append(projectExprs, "p.id")
		}
		if wantProjectType {
			projectExprs = append(projectExprs, "COALESCE(pt.id, 'DEFAULT')")
		}

		var query string
		if wantProjectType {
			query = `SELECT cpp.permission_id, ` + strings.Join(projectExprs, ", ") + `
				FROM cached_permission_projects cpp
				JOIN cached_permissions c ON c.id = cpp.permission_id AND c.user_id = $2
				JOIN projects p ON p.id = cpp.project_id
				LEFT JOIN project_types pt ON pt.id = p.project_type_id
				WHERE cpp.permission_id = ANY($1)
				ORDER BY cpp.permission_id, p.id`
		} else {
			query = `SELECT cpp.permission_id, ` + strings.Join(projectExprs, ", ") + `
				FROM cached_permission_projects cpp
				JOIN cached_permissions c ON c.id = cpp.permission_id AND c.user_id = $2
				JOIN projects p ON p.id = cpp.project_id
				WHERE cpp.permission_id = ANY($1)
				ORDER BY cpp.permission_id, p.id`
		}

		pRows, err := s.db.Query(ctx, query, ids, userID)
		if err != nil {
			return nil, err
		}
		defer pRows.Close()

		for pRows.Next() {
			var permID string
			var projectID string
			var projectTypeID string
			dests := []any{&permID}
			if wantID {
				dests = append(dests, &projectID)
			}
			if wantProjectType {
				dests = append(dests, &projectTypeID)
			}
			if err := pRows.Scan(dests...); err != nil {
				return nil, err
			}
			cp, ok := byID[permID]
			if !ok {
				continue
			}
			proj := &model.PermissionCacheProject{Type: "Project"}
			if wantID {
				proj.ID = projectID
			}
			if wantProjectType {
				proj.ProjectType = &model.ProjectType{ID: projectTypeID, Type: "ProjectType"}
			}
			cp.Projects = append(cp.Projects, proj)
		}
		if err := pRows.Err(); err != nil {
			return nil, err
		}
	}

	result := make([]*model.PermissionCacheEntry, 0, len(ids))
	for _, id := range ids {
		result = append(result, byID[id])
	}
	return result, nil
}

// serviceColumns يختار أعمدة جدول services المطلوبة حسب شجرة الحقول؛ يُرجع
// معرّفات الأعمدة مع تعبيرات SQL المناظرة (COALESCE للأعمدة النصية القابلة
// للنول). عندما تكون الشجرة فارغة (أو لم تُحدد) تُرجع كل الحقول.
func serviceColumns(tree *fields.FieldTree) (fieldIDs, fieldExprs []string) {
	all := []struct {
		id   string
		expr string
	}{
		{"id", "id"},
		{"name", "COALESCE(name, '')"},
		{"key", "COALESCE(key, '')"},
		{"home_url", "COALESCE(home_url, '')"},
		{"application_name", "COALESCE(application_name, '')"},
		{"vendor", "COALESCE(vendor, '')"},
		{"version", "COALESCE(version, '')"},
		{"trusted", "trusted"},
		{"icon_url", "COALESCE(icon_url, '')"},
		{"user_uri_pattern", "COALESCE(user_uri_pattern, '')"},
		{"group_uri_pattern", "COALESCE(group_uri_pattern, '')"},
		{"audience", "COALESCE(audience, '')"},
		{"immutable", "immutable"},
		{"client_credentials_flow_enabled", "client_credentials_flow_enabled"},
		{"auth_code_flow_enabled", "auth_code_flow_enabled"},
		{"implicit_flow_enabled", "implicit_flow_enabled"},
	}

	if tree == nil || tree.IsEmpty() {
		for _, c := range all {
			fieldIDs = append(fieldIDs, c.id)
			fieldExprs = append(fieldExprs, c.expr)
		}
		return fieldIDs, fieldExprs
	}

	jsonToColumn := []struct {
		json string
		id   string
		expr string
	}{
		{"id", "id", "id"},
		{"name", "name", "COALESCE(name, '')"},
		{"key", "key", "COALESCE(key, '')"},
		{"homeUrl", "home_url", "COALESCE(home_url, '')"},
		{"applicationName", "application_name", "COALESCE(application_name, '')"},
		{"vendor", "vendor", "COALESCE(vendor, '')"},
		{"version", "version", "COALESCE(version, '')"},
		{"trusted", "trusted", "trusted"},
		{"iconUrl", "icon_url", "COALESCE(icon_url, '')"},
		{"userUriPattern", "user_uri_pattern", "COALESCE(user_uri_pattern, '')"},
		{"groupUriPattern", "group_uri_pattern", "COALESCE(group_uri_pattern, '')"},
		{"audience", "audience", "COALESCE(audience, '')"},
		{"immutable", "immutable", "immutable"},
		{"clientCredentialsFlowEnabled", "client_credentials_flow_enabled", "client_credentials_flow_enabled"},
		{"authCodeFlowEnabled", "auth_code_flow_enabled", "auth_code_flow_enabled"},
		{"implicitFlowEnabled", "implicit_flow_enabled", "implicit_flow_enabled"},
	}

	for _, c := range jsonToColumn {
		if tree.Has(c.json) {
			fieldIDs = append(fieldIDs, c.id)
			fieldExprs = append(fieldExprs, c.expr)
		}
	}
	// معامل fields لا يطلب أي حقل معروف: نرجع كل الحقول الافتراضية.
	if len(fieldIDs) == 0 {
		return serviceColumns(nil)
	}
	return fieldIDs, fieldExprs
}

// Services يعيد صفحة خدمات Hub مع الجلب الانتقائي الدقيق حسب شجرة الحقول
// FieldTree ودعم $top و $skip (مطابق لـ request24.txt و hh.json و hh2.json).
// يتم حساب total عبر COUNT(*) وترتيب النتائج حسب المعرّف لثبات الترتيب.
func (s *AdminStore) Services(ctx context.Context, tree *fields.FieldTree, top int, skip int) (*model.ServicesPage, error) {
	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM services`).Scan(&total); err != nil {
		return nil, err
	}

	fieldIDs, fieldExprs := serviceColumns(tree)
	query := `SELECT ` + strings.Join(fieldExprs, ", ") + `
		FROM services
		ORDER BY id`
	if top > 0 {
		query += " LIMIT $1 OFFSET $2"
	}

	page := &model.ServicesPage{
		Type:     "ServicesPage",
		Skip:     skip,
		Top:      top,
		Total:    total,
		Services: []*model.HubService{},
	}

	var rows pgx.Rows
	var err error
	if top > 0 {
		rows, err = s.db.Query(ctx, query, top, skip)
	} else {
		rows, err = s.db.Query(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	selected := make(map[string]bool, len(fieldIDs))
	for _, fid := range fieldIDs {
		selected[fid] = true
	}

	for rows.Next() {
		svc := &model.HubService{Type: "service"}
		dests := make([]any, 0, len(fieldIDs))
		for _, fid := range fieldIDs {
			switch fid {
			case "id":
				dests = append(dests, &svc.ID)
			case "name":
				dests = append(dests, &svc.Name)
			case "key":
				dests = append(dests, &svc.Key)
			case "home_url":
				dests = append(dests, &svc.HomeURL)
			case "application_name":
				dests = append(dests, &svc.ApplicationName)
			case "vendor":
				dests = append(dests, &svc.Vendor)
			case "version":
				dests = append(dests, &svc.Version)
			case "trusted":
				svc.Trusted = new(bool)
				dests = append(dests, &svc.Trusted)
			case "icon_url":
				dests = append(dests, &svc.IconURL)
			case "user_uri_pattern":
				dests = append(dests, &svc.UserUriPattern)
			case "group_uri_pattern":
				dests = append(dests, &svc.GroupUriPattern)
			case "audience":
				dests = append(dests, &svc.Audience)
			case "immutable":
				svc.Immutable = new(bool)
				dests = append(dests, &svc.Immutable)
			case "client_credentials_flow_enabled":
				svc.ClientCredentialsFlowEnabled = new(bool)
				dests = append(dests, &svc.ClientCredentialsFlowEnabled)
			case "auth_code_flow_enabled":
				svc.AuthCodeFlowEnabled = new(bool)
				dests = append(dests, &svc.AuthCodeFlowEnabled)
			case "implicit_flow_enabled":
				svc.ImplicitFlowEnabled = new(bool)
				dests = append(dests, &svc.ImplicitFlowEnabled)
			}
		}
		if err := rows.Scan(dests...); err != nil {
			return nil, err
		}
		if !selected["trusted"] {
			svc.Trusted = nil
		}
		if !selected["immutable"] {
			svc.Immutable = nil
		}
		if !selected["client_credentials_flow_enabled"] {
			svc.ClientCredentialsFlowEnabled = nil
		}
		if !selected["auth_code_flow_enabled"] {
			svc.AuthCodeFlowEnabled = nil
		}
		if !selected["implicit_flow_enabled"] {
			svc.ImplicitFlowEnabled = nil
		}
		page.Services = append(page.Services, svc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return page, nil
}
