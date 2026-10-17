package rbac

import (
	"context"
	"fmt"
	"log/slog"

	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// AuthorizerAdapter adapts IRBACService to satisfy the middlewares.AuthorizerService interface.
// It bridges request-scoped role assignments with the high-performance PDP Engine.
type AuthorizerAdapter struct {
	rbacSvc IRBACService
}

// NewAuthorizerAdapter creates an adapter for middlewares.AuthorizerService.
func NewAuthorizerAdapter(rbacSvc IRBACService) *AuthorizerAdapter {
	return &AuthorizerAdapter{
		rbacSvc: rbacSvc,
	}
}

// IsAllowed evaluates whether the user is permitted to perform the action under the given scope.
func (a *AuthorizerAdapter) IsAllowed(ctx context.Context, userID string, permission perms2.PermKey, scope perms2.ScopeContext) (bool, error) {
	if a.rbacSvc == nil {
		return false, fmt.Errorf("%w: rbac service is not configured", ErrInvalidInput)
	}

	// 4. Construct the standard AccessRequest for the PDP Engine.
	req := AccessRequest{
		SubjectID:  userID,
		Permission: permission,
		ScopeType:  string(scope.Type),
		Resource:   string(scope.ResourceType),
		ResourceID: scope.ScopeID,
	}

	// 5. Evaluate through Policy Decision Point (PDP).
	decision := a.rbacSvc.Evaluate(ctx, req)
	if !decision.Allowed {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelWarn, "rbac: authorization denied",
			slog.String("user_id", userID),
			slog.String("scope", string(scope.Type)),
			slog.String("reason", decision.Reason),
		)
		return false, nil
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelWarn, "access allowd", slog.String("reason", decision.Reason))

	return true, nil
}

// func toRBACRoleAssignment(a RoleAssignment) RoleAssignment {
// 	perms := make([]Permission, len(a.Permissions))
// 	for i, p := range a.Permissions {
// 		perms[i] = p.ID
// 	}
// 	return RoleAssignment{
// 		AssignmentID:   a.AssignmentID,
// 		RoleID:         a.RoleID,
// 		RoleName:       a.RoleName,
// 		ScopeType:      string(a.ScopeType),
// 		OrganizationID: a.OrganizationID,
// 		ProjectID:      a.ProjectID,
// 		Permissions:    a.Permissions,
// 	}
// }

// // ToRBACRoleAssignments converts perms2 role assignments to native rbac role assignments.
// func ToRBACRoleAssignments(assignments []RoleAssignment) []RoleAssignment {
// 	out := make([]RoleAssignment, len(assignments))
// 	for i, a := range assignments {
// 		out[i] = toRBACRoleAssignment(a)
// 	}
// 	return out
// }

// ScopeMatchesAssignment determines whether a role assignment applies to a requested scope.
func ScopeMatchesAssignment(a RoleAssignment, s perms2.ScopeContext) bool {
	// A GLOBAL role assignment grants access across all scopes (Global, Organization, Project).
	if a.ScopeType == string(perms2.ScopeGlobal) {
		return true
	}

	switch s.Type {
	case perms2.ScopeGlobal:
		// Only GLOBAL assignments apply at the global scope level.
		return a.ScopeType == string(perms2.ScopeGlobal)

	case perms2.ScopeOrganization:
		// ORGANIZATION assignment applies if the organization IDs match.
		if a.ScopeType == string(perms2.ScopeOrganization) {
			return s.ScopeID == "" || a.OrganizationID == s.ScopeID
		}
		return false

	case perms2.ScopeProject:
		// PROJECT assignment applies if project IDs match.
		if a.ScopeType == string(perms2.ScopeProject) {
			return s.ScopeID == "" || a.ProjectID == s.ScopeID
		}
		// ORGANIZATION assignment also applies to projects within that organization.
		if a.ScopeType == string(perms2.ScopeOrganization) {
			return s.ScopeID == "" || a.OrganizationID == s.ScopeID
		}
		return false

	default:
		return false
	}
}
