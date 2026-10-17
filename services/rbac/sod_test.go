package rbac_test

import (
	"errors"
	"testing"
	"time"

	"youtrack_backend/internal/services/rbac"
)

func TestSoD_ConstraintManagement(t *testing.T) {
	sodMgr := rbac.NewSoDManager()

	// Valid constraint
	valid := rbac.SoDConstraint{
		ID:          "SOD_PO_PAYMENT",
		Name:        "PO Creation vs Payment Approval",
		Type:        rbac.StaticSoD,
		Roles:       []string{"PO_CREATOR", "PAYMENT_APPROVER"},
		MaxAllowed:  1,
		Description: "Prevent fraud by disallowing creation and approval by same user",
	}
	if err := sodMgr.AddConstraint(valid); err != nil {
		t.Fatalf("failed to add valid constraint: %v", err)
	}

	// Duplicate constraint
	if err := sodMgr.AddConstraint(valid); !errors.Is(err, rbac.ErrConstraintAlreadyExists) {
		t.Errorf("expected ErrConstraintAlreadyExists, got: %v", err)
	}

	// Invalid constraints
	invalidType := valid
	invalidType.ID = "INV_TYPE"
	invalidType.Type = "UNKNOWN"
	if err := sodMgr.AddConstraint(invalidType); !errors.Is(err, rbac.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid type, got: %v", err)
	}

	invalidRoles := valid
	invalidRoles.ID = "INV_ROLES"
	invalidRoles.Roles = []string{"ONLY_ONE"}
	if err := sodMgr.AddConstraint(invalidRoles); !errors.Is(err, rbac.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for single role, got: %v", err)
	}

	invalidMax := valid
	invalidMax.ID = "INV_MAX"
	invalidMax.MaxAllowed = 2 // >= len(Roles)
	if err := sodMgr.AddConstraint(invalidMax); !errors.Is(err, rbac.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for max_allowed >= len(Roles), got: %v", err)
	}

	// Retrieval and Listing
	c, err := sodMgr.GetConstraint("SOD_PO_PAYMENT")
	if err != nil || c.ID != "SOD_PO_PAYMENT" {
		t.Errorf("failed to retrieve constraint: %v", err)
	}

	list := sodMgr.ListConstraints()
	if len(list) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(list))
	}

	// Removal
	if err := sodMgr.RemoveConstraint("SOD_PO_PAYMENT"); err != nil {
		t.Errorf("failed to remove constraint: %v", err)
	}
	if _, err := sodMgr.GetConstraint("SOD_PO_PAYMENT"); !errors.Is(err, rbac.ErrConstraintNotFound) {
		t.Errorf("expected ErrConstraintNotFound after removal, got: %v", err)
	}
}

func TestSoD_StaticSeparationOfDuties_Direct(t *testing.T) {
	sodMgr := rbac.NewSoDManager()

	// Add Static SoD constraint
	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "SOD_BILLING",
		Name:       "Billing separation",
		Type:       rbac.StaticSoD,
		Roles:      []string{"INVOICE_CREATOR", "INVOICE_APPROVER"},
		MaxAllowed: 1,
	})

	// 1. User holds INVOICE_CREATOR, attempt to assign non-conflicting DEVELOPER -> Allow
	err := sodMgr.ValidateUserAssignment([]string{"INVOICE_CREATOR"}, "DEVELOPER")
	if err != nil {
		t.Errorf("non-conflicting assignment should succeed: %v", err)
	}

	// 2. User holds INVOICE_CREATOR, attempt to assign INVOICE_APPROVER -> Deny
	err = sodMgr.ValidateUserAssignment([]string{"INVOICE_CREATOR"}, "INVOICE_APPROVER")
	if !errors.Is(err, rbac.ErrSSDViolation) {
		t.Errorf("expected ErrSSDViolation when assigning conflicting role, got: %v", err)
	}
}

func TestSoD_StaticSeparationOfDuties_RolesAreIndependent(t *testing.T) {
	sodMgr := rbac.NewSoDManager()

	// Add Static Constraint between PO_CREATOR and PAYMENT_APPROVER
	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "SSD_PURCHASE",
		Name:       "Purchasing SSD",
		Type:       rbac.StaticSoD,
		Roles:      []string{"PO_CREATOR", "PAYMENT_APPROVER"},
		MaxAllowed: 1,
	})

	// A separate role does not imply PO_CREATOR, so this assignment is allowed.
	err := sodMgr.ValidateUserAssignment([]string{"PAYMENT_APPROVER"}, "SENIOR_PURCHASER")
	if err != nil {
		t.Errorf("independent role should not trigger an SSD conflict: %v", err)
	}
}

func TestSoD_DynamicSeparationOfDuties(t *testing.T) {
	sodMgr := rbac.NewSoDManager()

	// Add Dynamic SoD constraint (allowed in assignment, forbidden in same session)
	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "DSD_AUDIT",
		Name:       "Treasurer vs Auditor DSD",
		Type:       rbac.DynamicSoD,
		Roles:      []string{"TREASURER", "AUDITOR"},
		MaxAllowed: 1,
	})

	// 1. Static validation must ALLOW user to hold both roles
	err := sodMgr.ValidateUserAssignment([]string{"TREASURER"}, "AUDITOR")
	if err != nil {
		t.Errorf("Dynamic constraint must NOT block user assignment: %v", err)
	}

	// 2. In session with TREASURER active, activating a benign role DEVELOPER -> Allow
	err = sodMgr.ValidateSessionActivation([]string{"TREASURER"}, "DEVELOPER")
	if err != nil {
		t.Errorf("activating non-conflicting role in session should succeed: %v", err)
	}

	// 3. In session with TREASURER active, activating AUDITOR -> Deny with ErrDSDViolation
	err = sodMgr.ValidateSessionActivation([]string{"TREASURER"}, "AUDITOR")
	if !errors.Is(err, rbac.ErrDSDViolation) {
		t.Errorf("expected ErrDSDViolation when activating conflicting role in same session, got: %v", err)
	}
}

func TestSoD_AuditViolations(t *testing.T) {
	sodMgr := rbac.NewSoDManager()

	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "SSD_1",
		Name:       "Static 1",
		Type:       rbac.StaticSoD,
		Roles:      []string{"R1", "R2"},
		MaxAllowed: 1,
	})
	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "DSD_1",
		Name:       "Dynamic 1",
		Type:       rbac.DynamicSoD,
		Roles:      []string{"R3", "R4"},
		MaxAllowed: 1,
	})

	// Assignments: Alice violates SSD_1, Bob is clean
	assignments := []rbac.UserAssignment{
		{UserID: "alice", RoleIDs: []string{"R1", "R2"}},
		{UserID: "bob", RoleIDs: []string{"R1", "R5"}},
	}

	// Sessions: Session 101 violates DSD_1, Session 102 is clean
	sessions := []rbac.Session{
		{
			ID:     "sess_101",
			UserID: "charlie",
			ActiveRoles: map[string]time.Time{
				"R3": time.Now().Add(time.Hour),
				"R4": time.Now().Add(time.Hour),
			},
		},
		{
			ID:     "sess_102",
			UserID: "david",
			ActiveRoles: map[string]time.Time{
				"R3": time.Now().Add(time.Hour),
				"R5": time.Now().Add(time.Hour),
			},
		},
	}

	reports := sodMgr.AuditViolations(assignments, sessions)
	if len(reports) != 2 {
		t.Fatalf("expected 2 violation reports, got %d: %v", len(reports), reports)
	}

	// One SSD report for alice, one DSD report for sess_101
	var foundAliceSSD, foundCharlieDSD bool
	for _, r := range reports {
		if r.Type == rbac.StaticSoD && r.UserOrSessionID == "alice" {
			foundAliceSSD = true
		}
		if r.Type == rbac.DynamicSoD && r.UserOrSessionID == "sess_101" {
			foundCharlieDSD = true
		}
	}

	if !foundAliceSSD {
		t.Errorf("missing static violation report for alice")
	}
	if !foundCharlieDSD {
		t.Errorf("missing dynamic violation report for sess_101")
	}
}
