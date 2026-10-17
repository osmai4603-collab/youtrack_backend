package rbac

import "context"

type roleAssignmentsContextKey struct{}

// WithCurrentUserRoleAssignments attaches the user's role assignments to the context.
func WithCurrentUserRoleAssignments(ctx context.Context, assignments []RoleAssignment) context.Context {
	return context.WithValue(ctx, roleAssignmentsContextKey{}, assignments)
}

// GetCurrentUserRoleAssignments retrieves the user's role assignments from the context.
func GetCurrentUserRoleAssignments(ctx context.Context) ([]RoleAssignment, bool) {
	assignments, ok := ctx.Value(roleAssignmentsContextKey{}).([]RoleAssignment)
	return assignments, ok
}
