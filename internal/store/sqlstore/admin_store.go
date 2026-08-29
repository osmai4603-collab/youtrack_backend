package sqlstore

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
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
		g.RestSettings = &model.RestCorsSettings{
			AllowAllOrigins: allowAllOrigins,
			AllowedOrigins:  []string{},
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
