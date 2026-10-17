package rbac

import (
	"context"
	"fmt"
	"slices"
	"time"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// ImpliedPermissionResolver is an optional hook allowing the PDP to expand permissions
// with their implied permissions (e.g. YouTrack implied permissions hierarchy).
type ImpliedPermissionResolver interface {
	ResolveImplied(permissions []string) []string
}

// IPDPEngine defines the contract for an authoritative Policy Decision Point.
type IPDPEngine interface {
	// Evaluate assesses an incoming access request and renders an authoritative decision.
	// It guarantees Fail-Closed (Default Deny) semantics.
	Evaluate(ctx context.Context, req AccessRequest) Decision
}

// PDPEngine is the reference implementation of the Policy Decision Point (PDP).
type PDPEngine struct {
	graph           *RoleGraph
	impliedResolver ImpliedPermissionResolver
	contextMgr      *ContextManager
	permServ        perms2.IPermissionService
}

// NewPDPEngine creates a new Policy Decision Point engine backed by a role graph.
func NewPDPEngine(graph *RoleGraph, resolver ImpliedPermissionResolver, permServ perms2.IPermissionService) *PDPEngine {
	if graph == nil {
		graph = NewRoleGraph(nil)
	}
	return &PDPEngine{
		graph:           graph,
		impliedResolver: resolver,
		permServ:        permServ,
	}
}

// SetImpliedResolver updates or configures the implied permission resolver.
func (p *PDPEngine) SetImpliedResolver(resolver ImpliedPermissionResolver) {
	p.impliedResolver = resolver
}

// SetContextManager binds a contextual ABAC attribute evaluator manager to the PDP.
func (p *PDPEngine) SetContextManager(cm *ContextManager) {
	p.contextMgr = cm
}

// Evaluate evaluates an access request against the role registry and policies.
// Strict Fail-Closed (Default Deny) is enforced at all times.
func (p *PDPEngine) Evaluate(ctx context.Context, req AccessRequest) Decision {
	now := time.Now()

	assignments, hasAssignment := GetCurrentUserRoleAssignments(ctx)

	if !hasAssignment {
		return Decision{
			Allowed:     false,
			Reason:      "fail-closed: no assignments found",
			EvaluatedAt: now,
		}
	}

	if err := p.permServ.ContainsPermission(req.Permission); err != nil {
		return Decision{
			Allowed:     false,
			Reason:      err.Error(),
			EvaluatedAt: now,
		}
	}

	// 2. Guard check: At least one role or assignment must be provided
	if len(assignments) == 0 {
		return Decision{
			Allowed:     false,
			Reason:      "fail-closed: no active roles or assignments supplied in request",
			EvaluatedAt: now,
		}
	}

	// 3a. Extract permissions directly from Assignments sent with the request (no repository reload needed)
	for _, assignment := range assignments {

		if !slices.Contains(assignment.Permissions, req.Permission) {
			continue
		}

		// this condition is for administrators.
		if assignment.ScopeType == "GLOBAL" && req.ScopeType != "GLOBAL" {
			return Decision{
				Allowed:     true,
				Reason:      fmt.Sprintf("permission '%s' granted via role: %v", req.Permission, assignment.RoleName),
				EvaluatedAt: now,
			}
		}

		if assignment.ScopeType != req.ScopeType {
			continue
		}

		if req.ResourceID != "" && assignment.ScopeType == "PROJECT" && assignment.ProjectID != req.ResourceID {
			continue
		}

		if req.ResourceID != "" && assignment.ScopeType == "ORGANIZATION" && assignment.OrganizationID != req.ResourceID {
			continue
		}

		return Decision{
			Allowed:     true,
			Reason:      fmt.Sprintf("permission '%s' granted via role: %v", req.Permission, assignment.RoleName),
			EvaluatedAt: now,
		}

	}

	return Decision{
		Allowed:     false,
		Reason:      fmt.Sprintf("access denied: permission '%s' is not held by roles %v", req.Permission, getRolesNames(assignments)),
		EvaluatedAt: now,
	}

}

func getRolesNames(assginments []RoleAssignment) []string {
	var roles []string
	for _, assignment := range assginments {
		roles = append(roles, assignment.RoleName)
	}
	return roles
}
