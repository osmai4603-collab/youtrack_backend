package app

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetRoles يعيد قائمة الأدوار مع صلاحياتها.
func (a *App) GetRoles(ctx context.Context) ([]*model.Role, error) {
	roles, err := a.store.Admin().Roles(ctx)
	if err != nil {
		return nil, model.Internal("failed to load roles: %v", err)
	}
	for _, r := range roles {
		r.Type = "Role"
	}
	return roles, nil
}

// GetWidgets يعيد قائمة الودجات العامة (مطابق لـ request6.txt).
func (a *App) GetWidgets(ctx context.Context) ([]*model.DashboardWidget, error) {
	widgets, err := a.store.Admin().Widgets(ctx)
	if err != nil {
		return nil, model.Internal("failed to load widgets: %v", err)
	}
	for _, w := range widgets {
		w.Type = "WidgetView"
	}
	return widgets, nil
}

// GetPermissions يعيد قائمة الصلاحيات.
func (a *App) GetPermissions(ctx context.Context) ([]*model.Permission, error) {
	perms, err := a.store.Admin().Permissions(ctx)
	if err != nil {
		return nil, model.Internal("failed to load permissions: %v", err)
	}
	return perms, nil
}

// GetPermissionsCache يعيد الصلاحيات المخبأة للمستخدم مع نطاقاتها عبر المخطط
// الجديد المستقل PermissionCacheEntry مع الجلب الانتقائي حسب شجرة الحقول.
func (a *App) GetPermissionsCache(ctx context.Context, userID string, tree *fields.FieldTree) ([]*model.PermissionCacheEntry, error) {
	cache, err := a.store.Admin().PermissionsCache(ctx, userID, tree)
	if err != nil {
		return nil, model.Internal("failed to load permissions cache: %v", err)
	}
	if cache == nil {
		cache = []*model.PermissionCacheEntry{}
	}
	return cache, nil
}

// GetProjectDashboard يعيد لوحة ودجات المشروع عبر المخطط الجديد المستقل
// ProjectDashboard (مطابق لـ request13.txt و request23.txt)، مع ترجمة الأخطاء.
func (a *App) GetProjectDashboard(ctx context.Context, projectKey string, tree *fields.FieldTree) (*model.ProjectDashboard, error) {
	dashboard, err := a.store.Admin().ProjectDashboard(ctx, projectKey, tree)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.NotFound("project %q not found", projectKey)
		}
		return nil, model.Internal("failed to load project dashboard: %v", err)
	}
	dashboard.Normalize()
	return dashboard, nil
}

// GetOrganizations يعيد قائمة المنظمات مع الجلب الانتقائي حسب شجرة الحقول
// والفرز والحدود (مطابق لـ request22.txt و hh.json).
func (a *App) GetOrganizations(ctx context.Context, tree *fields.FieldTree, top int, skip int, sorting string) ([]*model.Organization, error) {
	orgs, err := a.store.Admin().Organizations(ctx, tree, top, skip, sorting)
	if err != nil {
		return nil, model.Internal("failed to load organizations: %v", err)
	}
	if orgs == nil {
		orgs = []*model.Organization{}
	}
	for _, o := range orgs {
		if o != nil {
			o.Normalize()
		}
	}
	return orgs, nil
}
