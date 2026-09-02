package app

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetRoles يعيد قائمة الأدوار.
func (a *YouTrackApp) GetRoles(c request.CTX) ([]*model.Role, *model.AppError) {
	service := NewPermissionService()
	if a == nil || a.channels == nil || a.channels.server == nil || a.channels.server.platform == nil || a.channels.server.platform.Store() == nil {
		return service.DefaultRoles(), nil
	}
	if c == nil {
		c = request.EmptyContext(nil)
	}
	ctx := c.Context()
	roles, err := service.LoadRolesFromStore(ctx, a.Store().Admin())
	if err != nil {
		return nil, model.NewInternalError("App.GetRoles", "failed to load roles", err)
	}
	if roles == nil {
		return service.DefaultRoles(), nil
	}
	return roles, nil
}

// GetWidgets يعيد قائمة ودجات لوحة التحكم.
func (a *YouTrackApp) GetWidgets(c request.CTX) ([]*model.DashboardWidget, *model.AppError) {
	ctx := c.Context()
	widgets, err := a.Store().Admin().Widgets(ctx)
	if err != nil {
		return nil, model.NewInternalError("App.GetWidgets", "failed to load widgets", err)
	}
	for _, w := range widgets {
		w.Type = "WidgetView"
	}
	return widgets, nil
}

// GetPermissions يعيد قائمة الصلاحيات.
func (a *YouTrackApp) GetPermissions(c request.CTX) ([]*model.Permission, *model.AppError) {
	service := NewPermissionService()
	if a == nil || a.channels == nil || a.channels.server == nil || a.channels.server.platform == nil || a.channels.server.platform.Store() == nil {
		return service.DefaultPermissions(), nil
	}
	if c == nil {
		c = request.EmptyContext(nil)
	}
	ctx := c.Context()
	perms, err := service.LoadPermissionsFromStore(ctx, a.Store().Admin())
	if err != nil {
		return nil, model.NewInternalError("App.GetPermissions", "failed to load permissions", err)
	}
	if perms == nil {
		return service.DefaultPermissions(), nil
	}
	return perms, nil
}

// GetRoleByName يعيد الدور حسب الاسم من مجموعة الأدوار الحالية أو الافتراضية.
func (a *YouTrackApp) GetRoleByName(c request.CTX, name string) (*model.Role, *model.AppError) {
	if name == "" {
		return nil, model.NewBadRequestError("App.GetRoleByName", "role name is required")
	}
	roles, err := a.GetRoles(c)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role != nil && strings.EqualFold(role.Name, name) {
			return role, nil
		}
	}
	return nil, model.NewNotFoundError("App.GetRoleByName", "role not found")
}

// GetPermissionByName يعيد الصلاحية حسب الاسم من مجموعة الصلاحيات الحالية أو الافتراضية.
func (a *YouTrackApp) GetPermissionByName(c request.CTX, name string) (*model.Permission, *model.AppError) {
	if name == "" {
		return nil, model.NewBadRequestError("App.GetPermissionByName", "permission name is required")
	}
	perms, err := a.GetPermissions(c)
	if err != nil {
		return nil, err
	}
	for _, perm := range perms {
		if perm != nil && strings.EqualFold(perm.Name, name) {
			return perm, nil
		}
	}
	return nil, model.NewNotFoundError("App.GetPermissionByName", "permission not found")
}

// GetPermissionsCache يعيد الصلاحيات المخبأة للمستخدم مع نطاقاتها.
func (a *YouTrackApp) GetPermissionsCache(c request.CTX, userID string, tree *fields.FieldTree) ([]*model.PermissionCacheEntry, *model.AppError) {
	if userID == "" {
		userID = c.UserID()
	}
	ctx := c.Context()
	cache, err := a.Store().Admin().PermissionsCache(ctx, userID, tree)
	if err != nil {
		return nil, model.NewInternalError("App.GetPermissionsCache", "failed to load permissions cache", err)
	}
	if cache == nil {
		cache = []*model.PermissionCacheEntry{}
	}
	return cache, nil
}

// GetProjectDashboard يعيد لوحة ودجات المشروع عبر المخطط المستقل ProjectDashboard.
func (a *YouTrackApp) GetProjectDashboard(c request.CTX, projectKey string, tree *fields.FieldTree) (*model.ProjectDashboard, *model.AppError) {
	ctx := c.Context()
	dashboard, err := a.Store().Admin().ProjectDashboard(ctx, projectKey, tree)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.NewNotFoundError("App.GetProjectDashboard", "project not found")
		}
		return nil, model.NewInternalError("App.GetProjectDashboard", "failed to load project dashboard", err)
	}
	dashboard.Normalize()
	return dashboard, nil
}

// GetOrganizations يعيد قائمة المنظمات مع الجلب الانتقائي حسب شجرة الحقول.
func (a *YouTrackApp) GetOrganizations(c request.CTX, tree *fields.FieldTree, top int, skip int, sorting string) ([]*model.Organization, *model.AppError) {
	ctx := c.Context()
	orgs, err := a.Store().Admin().Organizations(ctx, tree, top, skip, sorting)
	if err != nil {
		return nil, model.NewInternalError("App.GetOrganizations", "failed to load organizations", err)
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
