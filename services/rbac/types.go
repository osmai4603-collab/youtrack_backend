package rbac

import (
	"time"
)

// RoleAssignment represents a role assignment with resolved permissions within the RBAC domain.
type RoleAssignment struct {
	AssignmentID   string   `json:"assignment_id"`
	RoleID         string   `json:"role_id"`
	RoleName       string   `json:"role_name"`
	ScopeType      string   `json:"scope_type"`
	OrganizationID string   `json:"organization_id,omitempty"`
	ProjectID      string   `json:"project_id,omitempty"`
	Permissions    []string `json:"permissions"`
}

// RoleID represents the unique identifier of a role.
type RoleID = string

// Role represents an independent collection of permissions and attributes.
type Role struct {
	// ID is the unique identifier of the role (e.g., "SYSTEM_ADMIN", "PROJECT_VIEWER").
	ID RoleID `json:"id"`

	// Name is the human-readable display name of the role.
	Name string `json:"name"`

	// Description provides detailed context about the role's responsibilities.
	Description string `json:"description"`

	// Permissions is the list of permission keys explicitly granted to this role.
	Permissions []string `json:"permissions"`

	// IsImmutable indicates whether this role is a system role that cannot be deleted or modified.
	IsImmutable bool `json:"immutable"`

	// Metadata holds arbitrary key-value pairs for organizational tagging.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UserAssignment captures the assignment of one or more roles to a user,
// supporting optional temporal bounds (JIT access) and organizational scoping.
type UserAssignment struct {
	// UserID identifies the user subject.
	UserID string `json:"user_id"`

	// RoleIDs lists all roles assigned to the user within this scope.
	RoleIDs []RoleID `json:"role_ids"`

	// Scope limits the assignment to a specific organizational entity (e.g., "PROJECT:123", or "" for global).
	Scope string `json:"scope,omitempty"`

	// GrantedAt is the timestamp when the assignment was authorized.
	GrantedAt time.Time `json:"granted_at"`

	// ExpiresAt defines the optional expiration for Just-In-Time (JIT) access.
	// If nil, the assignment has no temporal expiry.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// GrantedBy records the authorizer ID for auditability.
	GrantedBy string `json:"granted_by,omitempty"`
}

// IsExpired checks whether a user assignment has exceeded its time-to-live.
func (ua *UserAssignment) IsExpired(now time.Time) bool {
	if ua.ExpiresAt == nil {
		return false
	}
	return now.After(*ua.ExpiresAt)
}

// SoDType defines the category of Separation of Duties constraint.
type SoDType string

const (
	// StaticSoD (SSD) enforces mutual exclusion during role assignment.
	StaticSoD SoDType = "STATIC"

	// DynamicSoD (DSD) enforces mutual exclusion during runtime role activation in a session.
	DynamicSoD SoDType = "DYNAMIC"
)

// SoDConstraint specifies a Separation of Duties rule according to ANSI/INCITS 359 Constrained RBAC.
type SoDConstraint struct {
	// ID uniquely identifies the constraint.
	ID string `json:"id"`

	// Name provides a human-readable title (e.g., "PO Creator vs Payment Approver").
	Name string `json:"name"`

	// Type determines whether the constraint is static (SSD) or dynamic (DSD).
	Type SoDType `json:"type"`

	// Roles lists the conflicting roles governed by this constraint.
	Roles []RoleID `json:"roles"`

	// MaxAllowed is the maximum number of conflicting roles a user can hold (SSD)
	// or activate concurrently (DSD). Typically 1.
	MaxAllowed int `json:"max_allowed"`

	// Description explains the risk or compliance requirement behind this rule.
	Description string `json:"description"`
}

// Session represents an active runtime security context in which a user activates
// a subset of their assigned roles, supporting Dynamic SoD and JIT access.
type Session struct {
	// ID uniquely identifies the session token/state.
	ID string `json:"id"`

	// UserID identifies the authenticated subject.
	UserID string `json:"user_id"`

	// ActiveRoles maps each currently active role to its activation expiration time.
	// A zero time.Time indicates activation until session termination.
	ActiveRoles map[RoleID]time.Time `json:"active_roles"`

	// CreatedAt marks session inception.
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt marks absolute session expiration.
	ExpiresAt time.Time `json:"expires_at"`

	// Metadata stores contextual session claims (e.g., IP address, device health).
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IsActive checks if the session is currently valid and non-expired.
func (s *Session) IsActive(now time.Time) bool {
	return now.Before(s.ExpiresAt)
}

// AccessRequest encapsulates all parameters required by the Policy Decision Point (PDP)
// to make an authoritative access decision.
type AccessRequest struct {
	// SubjectID is the identifier of the requesting user or service principal.
	SubjectID string `json:"subject_id"`

	// RoleIDs lists the active roles claimed in the request or session (optional if Assignments are provided).
	//RoleIDs []RoleID `json:"role_ids,omitempty"`

	// Assignments contains pre-resolved role assignments with embedded permissions sent with the request,
	// avoiding redundant lookups or reloading of permissions from repositories.
	//Assignments []RoleAssignment `json:"assignments,omitempty"`

	// Permission is the target permission key to verify (e.g., "UPDATE_PROJECT").
	Permission string `json:"permission"`

	// Resource is the targeted object identifier (e.g., "issue:1001", "project:alpha").
	Resource string `json:"resource,omitempty"`

	// ScopeType is the organizational level context (e.g., "GLOBAL", "ORGANIZATION", "PROJECT").
	ScopeType string `json:"scope,omitempty"`

	ResourceID string `json:"resouce_id"`

	// Context contains dynamic subject, resource, or environmental attributes for hybrid ABAC evaluation.
	Context map[string]any `json:"context,omitempty"`
}

// Decision represents the outcome produced by the Policy Decision Point (PDP).
type Decision struct {
	// Allowed indicates whether access is permitted (true) or denied (false).
	Allowed bool `json:"allowed"`

	// Reason provides auditable diagnostic rationale explaining the decision.
	Reason string `json:"reason"`

	// EvaluatedAt marks the evaluation timestamp.
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// SoDViolationReport captures details of a detected Separation of Duties violation.
type SoDViolationReport struct {
	ConstraintID    string   `json:"constraint_id"`
	ConstraintName  string   `json:"constraint_name"`
	Type            SoDType  `json:"type"`
	ViolatingRoles  []RoleID `json:"violating_roles"`
	UserOrSessionID string   `json:"user_or_session_id"`
	Message         string   `json:"message"`
}

// ComplianceReport summarizes system-wide RBAC governance metrics for audit attestation.
type ComplianceReport struct {
	GeneratedAt       time.Time            `json:"generated_at"`
	TotalRoles        int                  `json:"total_roles"`
	TotalAssignments  int                  `json:"total_assignments"`
	ActiveConstraints int                  `json:"active_constraints"`
	Violations        []SoDViolationReport `json:"violations,omitempty"`
}
