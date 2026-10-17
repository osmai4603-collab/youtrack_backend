package rbac

import (
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// PermissionCatalogAdapter bridges the YouTrack permission catalog (perms2.IPermissionService)
// with the RBAC engine's PermissionCatalog interface.
type PermissionCatalogAdapter struct {
	catalog perms2.IPermissionService
}

// NewPermissionCatalogAdapter constructs a new adapter for the given permission service.
func NewPermissionCatalogAdapter(catalog perms2.IPermissionService) *PermissionCatalogAdapter {
	return &PermissionCatalogAdapter{catalog: catalog}
}

// ResolveImplied expands the input permission keys by resolving all transitive implied permissions.
func (a *PermissionCatalogAdapter) ResolveImplied(permissionIDs []string) []string {
	if a.catalog == nil || len(permissionIDs) == 0 {
		return permissionIDs
	}
	return a.catalog.ResolveImplied(permissionIDs)
}

// Ensure interface compliance.
var _ PermissionCatalog = (*PermissionCatalogAdapter)(nil)
