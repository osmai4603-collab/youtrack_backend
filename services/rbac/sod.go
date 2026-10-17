package rbac

import (
	"fmt"
	"sort"
	"sync"
)

// SoDManager manages and enforces Separation of Duties constraints (SSD and DSD)
// in accordance with the ANSI/INCITS 359 Constrained RBAC specification.
type SoDManager struct {
	mu          sync.RWMutex
	constraints map[string]SoDConstraint
}

// NewSoDManager creates a new SoD constraint manager.
func NewSoDManager() *SoDManager {
	return &SoDManager{
		constraints: make(map[string]SoDConstraint),
	}
}

// AddConstraint registers a new Separation of Duties constraint.
func (m *SoDManager) AddConstraint(c SoDConstraint) error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("%w: constraint ID and name must not be empty", ErrInvalidInput)
	}
	if c.Type != StaticSoD && c.Type != DynamicSoD {
		return fmt.Errorf("%w: invalid constraint type %s (must be STATIC or DYNAMIC)", ErrInvalidInput, c.Type)
	}
	if len(c.Roles) < 2 {
		return fmt.Errorf("%w: constraint must involve at least 2 roles", ErrInvalidInput)
	}
	if c.MaxAllowed < 1 || c.MaxAllowed >= len(c.Roles) {
		return fmt.Errorf("%w: max_allowed must be between 1 and %d", ErrInvalidInput, len(c.Roles)-1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.constraints[c.ID]; exists {
		return fmt.Errorf("%w: %s", ErrConstraintAlreadyExists, c.ID)
	}

	// Defensive copy
	rolesCopy := make([]RoleID, len(c.Roles))
	copy(rolesCopy, c.Roles)
	c.Roles = rolesCopy

	m.constraints[c.ID] = c
	return nil
}

// RemoveConstraint deletes a constraint by its unique ID.
func (m *SoDManager) RemoveConstraint(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.constraints[id]; !exists {
		return fmt.Errorf("%w: %s", ErrConstraintNotFound, id)
	}
	delete(m.constraints, id)
	return nil
}

// GetConstraint retrieves a constraint definition by ID.
func (m *SoDManager) GetConstraint(id string) (SoDConstraint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, exists := m.constraints[id]
	if !exists {
		return SoDConstraint{}, fmt.Errorf("%w: %s", ErrConstraintNotFound, id)
	}
	return copyConstraint(c), nil
}

// ListConstraints returns all registered constraints sorted by ID.
func (m *SoDManager) ListConstraints() []SoDConstraint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]SoDConstraint, 0, len(m.constraints))
	for _, c := range m.constraints {
		res = append(res, copyConstraint(c))
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

// ListConstraintsByType returns all constraints of a specific type (STATIC or DYNAMIC).
func (m *SoDManager) ListConstraintsByType(t SoDType) []SoDConstraint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []SoDConstraint
	for _, c := range m.constraints {
		if c.Type == t {
			res = append(res, copyConstraint(c))
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

// ValidateUserAssignment enforces Static Separation of Duties (SSD).
// It verifies whether assigning `newRole` to a user holding `existingRoles`
// would breach any static constraint.
func (m *SoDManager) ValidateUserAssignment(existingRoles []RoleID, newRole RoleID) error {
	if newRole == "" {
		return fmt.Errorf("%w: new role ID cannot be empty", ErrInvalidInput)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Gather candidate roles
	candidateRoles := append([]RoleID{}, existingRoles...)
	candidateRoles = append(candidateRoles, newRole)

	candidateRoleSet := roleSet(candidateRoles)

	// Check each Static constraint
	for _, c := range m.constraints {
		if c.Type != StaticSoD {
			continue
		}

		violating := findIntersectingRoles(c.Roles, candidateRoleSet)
		if len(violating) > c.MaxAllowed {
			return fmt.Errorf("%w: constraint '%s' (max allowed: %d, present: %v)",
				ErrSSDViolation, c.Name, c.MaxAllowed, violating)
		}
	}

	return nil
}

// ValidateSessionActivation enforces Dynamic Separation of Duties (DSD).
// It verifies whether activating `roleToActivate` within a session currently running
// `activeRoles` would breach any dynamic constraint.
func (m *SoDManager) ValidateSessionActivation(activeRoles []RoleID, roleToActivate RoleID) error {
	if roleToActivate == "" {
		return fmt.Errorf("%w: role to activate cannot be empty", ErrInvalidInput)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Gather candidate session roles
	candidateRoles := append([]RoleID{}, activeRoles...)
	candidateRoles = append(candidateRoles, roleToActivate)

	candidateRoleSet := roleSet(candidateRoles)

	// Check each Dynamic constraint
	for _, c := range m.constraints {
		if c.Type != DynamicSoD {
			continue
		}

		violating := findIntersectingRoles(c.Roles, candidateRoleSet)
		if len(violating) > c.MaxAllowed {
			return fmt.Errorf("%w: constraint '%s' (max allowed: %d, present in session: %v)",
				ErrDSDViolation, c.Name, c.MaxAllowed, violating)
		}
	}

	return nil
}

// AuditViolations scans user assignments and active sessions across the system,
// producing detailed compliance reports for any Static or Dynamic SoD breaches.
func (m *SoDManager) AuditViolations(assignments []UserAssignment, sessions []Session) []SoDViolationReport {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var reports []SoDViolationReport

	// 1. Audit Static SoD on User Assignments
	for _, ua := range assignments {
		assignedRoleSet := roleSet(ua.RoleIDs)

		for _, c := range m.constraints {
			if c.Type != StaticSoD {
				continue
			}

			violating := findIntersectingRoles(c.Roles, assignedRoleSet)
			if len(violating) > c.MaxAllowed {
				reports = append(reports, SoDViolationReport{
					ConstraintID:    c.ID,
					ConstraintName:  c.Name,
					Type:            c.Type,
					ViolatingRoles:  violating,
					UserOrSessionID: ua.UserID,
					Message: fmt.Sprintf("User %s holds conflicting roles %v violating static constraint %s (max %d)",
						ua.UserID, violating, c.Name, c.MaxAllowed),
				})
			}
		}
	}

	// 2. Audit Dynamic SoD on Active Sessions
	for _, sess := range sessions {
		var activeRoleList []RoleID
		for r := range sess.ActiveRoles {
			activeRoleList = append(activeRoleList, r)
		}
		activeRoleSet := roleSet(activeRoleList)

		for _, c := range m.constraints {
			if c.Type != DynamicSoD {
				continue
			}

			violating := findIntersectingRoles(c.Roles, activeRoleSet)
			if len(violating) > c.MaxAllowed {
				reports = append(reports, SoDViolationReport{
					ConstraintID:    c.ID,
					ConstraintName:  c.Name,
					Type:            c.Type,
					ViolatingRoles:  violating,
					UserOrSessionID: sess.ID,
					Message: fmt.Sprintf("Session %s of user %s has active conflicting roles %v violating dynamic constraint %s (max %d)",
						sess.ID, sess.UserID, violating, c.Name, c.MaxAllowed),
				})
			}
		}
	}

	return reports
}

func roleSet(roles []RoleID) map[RoleID]struct{} {
	set := make(map[RoleID]struct{}, len(roles))
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return set
}

// findIntersectingRoles returns the subset of constraint roles that are present in the target set.
func findIntersectingRoles(constraintRoles []RoleID, targetSet map[RoleID]struct{}) []RoleID {
	var matched []RoleID
	for _, r := range constraintRoles {
		if _, exists := targetSet[r]; exists {
			matched = append(matched, r)
		}
	}
	sort.Strings(matched)
	return matched
}

// copyConstraint produces a deep copy of an SoDConstraint.
func copyConstraint(c SoDConstraint) SoDConstraint {
	cp := c
	if c.Roles != nil {
		cp.Roles = make([]RoleID, len(c.Roles))
		copy(cp.Roles, c.Roles)
	}
	return cp
}
