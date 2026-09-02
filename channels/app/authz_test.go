package app

import (
	"context"
	"testing"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

func TestRolesGrantPermission(t *testing.T) {
	roles := []string{"system_admin"}
	if !RolesGrantPermission(roles, "project.read") {
		t.Fatal("system_admin should have project.read")
	}

	if !RolesGrantPermission([]string{"user"}, "issue.read") {
		t.Fatal("user should have issue.read")
	}

	if RolesGrantPermission([]string{"user"}, "project.delete") {
		t.Fatal("user should not have project.delete")
	}
}

func TestAppSessionHasPermission(t *testing.T) {
	ctx := request.NewContext(nil, "r1", "127.0.0.1", "/api/projects", "test-agent")
	ctx = ctx.WithUserID("u-1").(*request.Context)
	ctx = ctx.WithUserRoles([]string{"project_admin"}).(*request.Context)

	app := &YouTrackApp{}
	if !app.SessionHasPermission(ctx, "project.write") {
		t.Fatal("project_admin should have project.write")
	}

	if app.SessionHasPermission(ctx, "system.config") {
		t.Fatal("project_admin should not have system.config")
	}

	_ = model.User{ID: "u-1", Roles: []string{"user"}}
}

func TestPermissionRegistry(t *testing.T) {
	if !IsKnownPermission(PermissionIssueRead) {
		t.Fatal("issue.read should be registered")
	}
	if !IsKnownPermission(PermissionSystemAdmin) {
		t.Fatal("system.admin should be registered")
	}

	perms := RolePermissions("user")
	if len(perms) == 0 {
		t.Fatal("user role should have permissions")
	}
	if !containsPermission(perms, PermissionProjectRead) {
		t.Fatal("user role should include project.read")
	}

	if containsPermission(RolePermissions("system_admin"), PermissionSystemAdmin) == false {
		t.Fatal("system_admin should include system.admin")
	}
}

func TestDefaultRoleModels(t *testing.T) {
	roles := DefaultRoles()
	if len(roles) == 0 {
		t.Fatal("default roles should not be empty")
	}

	var foundUser bool
	for _, role := range roles {
		if role == nil {
			continue
		}
		if role.Name == "user" {
			foundUser = true
			if len(role.Permissions) == 0 {
				t.Fatal("user role should have model permissions")
			}
			if !containsModelPermission(role.Permissions, PermissionProjectRead) {
				t.Fatal("user role should include project.read in model permissions")
			}
		}
	}
	if !foundUser {
		t.Fatal("default roles should include user")
	}
}

func containsPermission(perms []string, permission string) bool {
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

func containsModelPermission(perms []*model.Permission, permission string) bool {
	for _, p := range perms {
		if p != nil && p.Name == permission {
			return true
		}
	}
	return false
}

func TestPermissionServiceRoleAndPermissionLookups(t *testing.T) {
	service := NewPermissionService()

	role := service.RoleByName("user")
	if role == nil {
		t.Fatal("user role should resolve from default registry")
	}
	if !containsModelPermission(role.Permissions, PermissionProjectRead) {
		t.Fatal("user role should include project.read")
	}

	perm := service.PermissionByName(PermissionSystemAdmin)
	if perm == nil {
		t.Fatal("system.admin permission should resolve from default registry")
	}
	if perm.Name != PermissionSystemAdmin {
		t.Fatal("resolved permission name should match requested permission")
	}
}

func TestPermissionServiceLoadsFromStoreOrFallbacks(t *testing.T) {
	service := NewPermissionService()

	roles, err := service.LoadRolesFromStore(t.Context(), nil)
	if err != nil {
		t.Fatalf("LoadRolesFromStore should fallback to defaults when admin store is nil: %v", err)
	}
	if len(roles) == 0 {
		t.Fatal("fallback roles should not be empty")
	}

	perms, err := service.LoadPermissionsFromStore(t.Context(), nil)
	if err != nil {
		t.Fatalf("LoadPermissionsFromStore should fallback to defaults when admin store is nil: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("fallback permissions should not be empty")
	}
}

func TestAppGetRolesAndPermissionsFallbackToDefaults(t *testing.T) {
	app := &YouTrackApp{}

	roles, err := app.GetRoles(nil)
	if err != nil {
		t.Fatalf("GetRoles should fallback to defaults without store, got error: %v", err)
	}
	if len(roles) == 0 {
		t.Fatal("GetRoles should return default roles when store is unavailable")
	}

	perms, err := app.GetPermissions(nil)
	if err != nil {
		t.Fatalf("GetPermissions should fallback to defaults without store, got error: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("GetPermissions should return default permissions when store is unavailable")
	}

	if !containsModelPermission(perms, PermissionProjectRead) {
		t.Fatal("default permissions should include project.read")
	}
}

type mockAdminRolePermissionStore struct {
	roles []*model.Role
	perms []*model.Permission
}

func (m *mockAdminRolePermissionStore) Roles(ctx context.Context) ([]*model.Role, error) {
	return m.roles, nil
}

func (m *mockAdminRolePermissionStore) Permissions(ctx context.Context) ([]*model.Permission, error) {
	return m.perms, nil
}

func (m *mockAdminRolePermissionStore) GlobalSettings(ctx context.Context) (*model.GlobalSettings, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) AdminGlobalSettings(ctx context.Context, tree *fields.FieldTree) (*model.AdminGlobalSettings, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) BannersConfig(ctx context.Context, tree *fields.FieldTree) (*model.BannersConfig, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) Widgets(ctx context.Context) ([]*model.DashboardWidget, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) CachedPermissions(ctx context.Context, userID string) ([]*model.CachedPermission, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) PermissionsCache(ctx context.Context, userID string, tree *fields.FieldTree) ([]*model.PermissionCacheEntry, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) ProjectDashboard(ctx context.Context, projectKey string, tree *fields.FieldTree) (*model.ProjectDashboard, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) Organizations(ctx context.Context, tree *fields.FieldTree, top int, skip int, sorting string) ([]*model.Organization, error) {
	return nil, nil
}

func (m *mockAdminRolePermissionStore) Services(ctx context.Context, tree *fields.FieldTree, top int, skip int) (*model.ServicesPage, error) {
	return nil, nil
}

func TestPermissionServiceLoadsStoreValuesWhenPresent(t *testing.T) {
	customRole := &model.Role{ID: "custom", Name: "custom", Permissions: []*model.Permission{{ID: "custom.permission", Name: "custom.permission"}}}
	customPerm := &model.Permission{ID: "custom.permission", Name: "custom.permission"}
	store := &mockAdminRolePermissionStore{roles: []*model.Role{customRole}, perms: []*model.Permission{customPerm}}

	service := NewPermissionService()
	roles, err := service.LoadRolesFromStore(context.Background(), store)
	if err != nil {
		t.Fatalf("LoadRolesFromStore returned error for populated store: %v", err)
	}
	if len(roles) != 1 || roles[0].Name != "custom" {
		t.Fatal("LoadRolesFromStore should return store data when present")
	}

	perms, err := service.LoadPermissionsFromStore(context.Background(), store)
	if err != nil {
		t.Fatalf("LoadPermissionsFromStore returned error for populated store: %v", err)
	}
	if len(perms) != 1 || perms[0].Name != "custom.permission" {
		t.Fatal("LoadPermissionsFromStore should return store data when present")
	}
}
