package rbac

import "errors"

var (
	// ErrRoleNotFound indicates that the requested role ID does not exist.
	ErrRoleNotFound = errors.New("rbac: role not found")

	// ErrRoleAlreadyExists indicates that a role with the given ID already exists.
	ErrRoleAlreadyExists = errors.New("rbac: role already exists")

	// ErrRoleImmutable indicates that the role is marked as immutable system role and cannot be modified or deleted.
	ErrRoleImmutable = errors.New("rbac: role is immutable and cannot be modified or deleted")

	// ErrSSDViolation indicates that assigning the role violates a Static Separation of Duties constraint.
	ErrSSDViolation = errors.New("rbac: static separation of duties (SSD) constraint violation")

	// ErrDSDViolation indicates that activating the role in the session violates a Dynamic Separation of Duties constraint.
	ErrDSDViolation = errors.New("rbac: dynamic separation of duties (DSD) constraint violation")

	// ErrSessionNotFound indicates that the requested session does not exist.
	ErrSessionNotFound = errors.New("rbac: session not found")

	// ErrSessionExpired indicates that the session has exceeded its validity window.
	ErrSessionExpired = errors.New("rbac: session has expired")

	// ErrRoleExpired indicates that a temporary Just-In-Time (JIT) role assignment has expired.
	ErrRoleExpired = errors.New("rbac: role assignment has expired")

	// ErrAccessDenied indicates that the subject lacks necessary permissions or fails contextual policy requirements.
	ErrAccessDenied = errors.New("rbac: access denied")

	// ErrInvalidInput indicates that an input argument, ID, or slice is empty or malformed.
	ErrInvalidInput = errors.New("rbac: invalid input parameter")

	// ErrConstraintNotFound indicates that the specified SoD constraint was not found.
	ErrConstraintNotFound = errors.New("rbac: sod constraint not found")

	// ErrConstraintAlreadyExists indicates that an SoD constraint with the same ID already exists.
	ErrConstraintAlreadyExists = errors.New("rbac: sod constraint already exists")
)
