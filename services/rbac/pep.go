package rbac

import (
	"context"
	"fmt"
	"strings"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// Enforcer serves as the Policy Enforcement Point (PEP), intercepting access requests
// and enforcing decisions rendered by the PDP.
type Enforcer struct {
	pdp IPDPEngine
}

// NewEnforcer instantiates a new Policy Enforcement Point bound to a PDP engine.
func NewEnforcer(pdp IPDPEngine, permService perms2.IPermissionService) *Enforcer {
	if pdp == nil {
		pdp = NewPDPEngine(nil, nil, permService)
	}
	return &Enforcer{
		pdp: pdp,
	}
}

// Authorize evaluates the access request and returns nil if permitted,
// or an error wrapping ErrAccessDenied if refused.
func (e *Enforcer) Authorize(ctx context.Context, req AccessRequest) error {
	decision := e.pdp.Evaluate(ctx, req)
	if !decision.Allowed {
		return fmt.Errorf("%w: %s", ErrAccessDenied, decision.Reason)
	}
	return nil
}

// AuthorizePermission is a convenience method to verify a single required permission.
func (e *Enforcer) AuthorizePermission(ctx context.Context, subjectID string, permission string) error {
	req := AccessRequest{
		SubjectID:  subjectID,
		Permission: permission,
	}
	return e.Authorize(ctx, req)
}

// AuthorizeAny verifies that the subject possesses at least one of the listed permissions.
// If none are permitted, ErrAccessDenied is returned.
func (e *Enforcer) AuthorizeAny(ctx context.Context, subjectID string, permissions ...string) error {
	if len(permissions) == 0 {
		return fmt.Errorf("%w: no permissions specified to authorize", ErrInvalidInput)
	}

	var reasons []string
	for _, perm := range permissions {
		req := AccessRequest{
			SubjectID:  subjectID,
			Permission: perm,
		}
		decision := e.pdp.Evaluate(ctx, req)
		if decision.Allowed {
			return nil // At least one matched
		}
		reasons = append(reasons, decision.Reason)
	}

	return fmt.Errorf("%w: none of required permissions %v were granted (%s)",
		ErrAccessDenied, permissions, strings.Join(reasons, "; "))
}

// AuthorizeAll verifies that the subject possesses all of the listed permissions.
// If any single permission is missing, authorization fails immediately (Fail-Closed).
func (e *Enforcer) AuthorizeAll(ctx context.Context, subjectID string, permissions ...string) error {
	if len(permissions) == 0 {
		return fmt.Errorf("%w: no permissions specified to authorize", ErrInvalidInput)
	}

	for _, perm := range permissions {
		req := AccessRequest{
			SubjectID:  subjectID,
			Permission: perm,
		}
		decision := e.pdp.Evaluate(ctx, req)
		if !decision.Allowed {
			return fmt.Errorf("%w: permission '%s' not granted: %s",
				ErrAccessDenied, perm, decision.Reason)
		}
	}

	return nil
}

// GetPDP returns the underlying Policy Decision Point.
func (e *Enforcer) GetPDP() IPDPEngine {
	return e.pdp
}
