package rbac

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// SymmetricAuditor provides bidirectional auditing, access recertification queries,
// and governance reporting pursuant to ANSI/INCITS 359 Symmetric RBAC.
type SymmetricAuditor struct {
	mu          sync.RWMutex
	graph       *RoleGraph
	sodMgr      *SoDManager
	assignments map[string][]UserAssignment // userID -> assignments
}

// NewSymmetricAuditor instantiates a new SymmetricAuditor.
func NewSymmetricAuditor(graph *RoleGraph, sodMgr *SoDManager) *SymmetricAuditor {
	if graph == nil {
		graph = NewRoleGraph(nil)
	}
	if sodMgr == nil {
		sodMgr = NewSoDManager()
	}
	return &SymmetricAuditor{
		graph:       graph,
		sodMgr:      sodMgr,
		assignments: make(map[string][]UserAssignment),
	}
}

// AssignRole assigns a role to a user, checking Static SoD (SSD) constraints.
func (a *SymmetricAuditor) AssignRole(ua UserAssignment) error {
	if ua.UserID == "" || len(ua.RoleIDs) == 0 {
		return fmt.Errorf("%w: user ID and at least one role ID are required", ErrInvalidInput)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Pre-validate Static Separation of Duties (SSD)
	existingRoles := a.getActiveRoleIDsForUserLocked(ua.UserID, time.Now())
	for _, newRole := range ua.RoleIDs {
		if a.sodMgr != nil {
			if err := a.sodMgr.ValidateUserAssignment(existingRoles, newRole); err != nil {
				return err
			}
		}
		existingRoles = append(existingRoles, newRole)
	}

	if ua.GrantedAt.IsZero() {
		ua.GrantedAt = time.Now()
	}

	a.assignments[ua.UserID] = append(a.assignments[ua.UserID], ua)
	return nil
}

// RevokeRole revokes an assignment matching the role and optional scope for a user.
func (a *SymmetricAuditor) RevokeRole(userID string, roleID RoleID, scope string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	userList, exists := a.assignments[userID]
	if !exists || len(userList) == 0 {
		return fmt.Errorf("%w: user %s has no assignments", ErrInvalidInput, userID)
	}

	var updated []UserAssignment
	for _, ua := range userList {
		// If scope is specified, match scope; otherwise revoke across all scopes
		if scope != "" && ua.Scope != scope {
			updated = append(updated, ua)
			continue
		}

		// Filter out the roleID
		var remainingRoles []RoleID
		for _, r := range ua.RoleIDs {
			if r != roleID {
				remainingRoles = append(remainingRoles, r)
			}
		}

		if len(remainingRoles) > 0 {
			ua.RoleIDs = remainingRoles
			updated = append(updated, ua)
		}
	}

	a.assignments[userID] = updated
	return nil
}

// GetUserAssignments returns all active, non-expired assignments for the given user.
func (a *SymmetricAuditor) GetUserAssignments(userID string) []UserAssignment {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	var res []UserAssignment
	for _, ua := range a.assignments[userID] {
		if !ua.IsExpired(now) {
			res = append(res, ua)
		}
	}
	return res
}

// GetAllAssignments returns all non-expired assignments across all users.
func (a *SymmetricAuditor) GetAllAssignments() []UserAssignment {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	var res []UserAssignment
	for _, list := range a.assignments {
		for _, ua := range list {
			if !ua.IsExpired(now) {
				res = append(res, ua)
			}
		}
	}
	return res
}

// GetUserEffectivePermissions queries from the user perspective:
// "What permissions are explicitly granted by this user's assigned roles?"
func (a *SymmetricAuditor) GetUserEffectivePermissions(userID string) ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	userRoles := a.getActiveRoleIDsForUserLocked(userID, now)
	if len(userRoles) == 0 {
		return []string{}, nil
	}

	permSet := make(map[string]struct{})
	for _, roleID := range userRoles {
		perms, err := a.graph.GetEffectivePermissions(roleID)
		if err != nil {
			continue
		}
		for _, p := range perms {
			permSet[p] = struct{}{}
		}
	}

	res := make([]string, 0, len(permSet))
	for p := range permSet {
		res = append(res, p)
	}
	sort.Strings(res)
	return res, nil
}

// GetRolesWithPermission queries from the permission perspective:
// "Which roles have this target permission explicitly granted?"
func (a *SymmetricAuditor) GetRolesWithPermission(permission string) ([]RoleID, error) {
	allRoles := a.graph.GetAllRoles()

	var matchingRoles []RoleID
	for _, r := range allRoles {
		hasPerm, err := a.graph.HasPermission(r.ID, permission)
		if err == nil && hasPerm {
			matchingRoles = append(matchingRoles, r.ID)
		}
	}
	sort.Strings(matchingRoles)
	return matchingRoles, nil
}

// GetUsersWithPermission queries from the PERMISSION perspective:
// "Who are all the users that currently possess this target permission?"
// Essential for audit reviews (e.g. "Who can delete projects?").
func (a *SymmetricAuditor) GetUsersWithPermission(permission string) ([]string, error) {
	qualifyingRoles, err := a.GetRolesWithPermission(permission)
	if err != nil || len(qualifyingRoles) == 0 {
		return []string{}, nil
	}

	roleSet := make(map[RoleID]struct{}, len(qualifyingRoles))
	for _, r := range qualifyingRoles {
		roleSet[r] = struct{}{}
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	userSet := make(map[string]struct{})

	for userID, uList := range a.assignments {
		for _, ua := range uList {
			if ua.IsExpired(now) {
				continue
			}
			for _, r := range ua.RoleIDs {
				if _, matches := roleSet[r]; matches {
					userSet[userID] = struct{}{}
					break
				}
			}
		}
	}

	result := make([]string, 0, len(userSet))
	for u := range userSet {
		result = append(result, u)
	}
	sort.Strings(result)
	return result, nil
}

// FindOrphanedRoles identifies stale roles that have no active users assigned,
// excluding immutable system roles. Essential for preventing role creep / privilege bloat.
func (a *SymmetricAuditor) FindOrphanedRoles() []RoleID {
	a.mu.RLock()
	defer a.mu.RUnlock()

	allRoles := a.graph.GetAllRoles()
	now := time.Now()

	// Map of assigned roles currently in use
	assignedRoleSet := make(map[RoleID]struct{})
	for _, list := range a.assignments {
		for _, ua := range list {
			if !ua.IsExpired(now) {
				for _, r := range ua.RoleIDs {
					assignedRoleSet[r] = struct{}{}
				}
			}
		}
	}

	var orphaned []RoleID
	for _, r := range allRoles {
		if r.IsImmutable {
			continue // System roles are protected and not orphaned
		}
		if _, inUse := assignedRoleSet[r.ID]; !inUse {
			orphaned = append(orphaned, r.ID)
		}
	}
	sort.Strings(orphaned)
	return orphaned
}

// GenerateComplianceReport compiles an executive audit attestation report
// detailing role counts, assignments, SoD constraints, and any active policy violations.
func (a *SymmetricAuditor) GenerateComplianceReport(sessions []Session) *ComplianceReport {
	allRoles := a.graph.GetAllRoles()
	assignments := a.GetAllAssignments()

	var constraints []SoDConstraint
	var violations []SoDViolationReport
	if a.sodMgr != nil {
		constraints = a.sodMgr.ListConstraints()
		violations = a.sodMgr.AuditViolations(assignments, sessions)
	}

	return &ComplianceReport{
		GeneratedAt:       time.Now(),
		TotalRoles:        len(allRoles),
		TotalAssignments:  len(assignments),
		ActiveConstraints: len(constraints),
		Violations:        violations,
	}
}

// getActiveRoleIDsForUserLocked returns all non-expired role IDs for a user.
// Assumes lock is held.
func (a *SymmetricAuditor) getActiveRoleIDsForUserLocked(userID string, now time.Time) []RoleID {
	roleSet := make(map[RoleID]struct{})
	for _, ua := range a.assignments[userID] {
		if !ua.IsExpired(now) {
			for _, r := range ua.RoleIDs {
				roleSet[r] = struct{}{}
			}
		}
	}
	var res []RoleID
	for r := range roleSet {
		res = append(res, r)
	}
	sort.Strings(res)
	return res
}
