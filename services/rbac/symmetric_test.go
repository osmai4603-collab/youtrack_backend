package rbac_test

import (
	"errors"
	"testing"

	"youtrack_backend/internal/services/rbac"
)

func TestSymmetric_BidirectionalAuditing(t *testing.T) {
	graph := rbac.NewRoleGraph(nil)

	// Each role receives only its explicitly listed permissions.
	_ = graph.AddRole(&rbac.Role{ID: "VIEWER", Permissions: []string{"READ_DOCS"}})
	_ = graph.AddRole(&rbac.Role{ID: "EDITOR", Permissions: []string{"READ_DOCS", "UPDATE_DOCS"}})
	_ = graph.AddRole(&rbac.Role{ID: "ADMIN", Permissions: []string{"READ_DOCS", "UPDATE_DOCS", "DELETE_DOCS"}})
	_ = graph.AddRole(&rbac.Role{ID: "GUEST", Permissions: []string{"READ_PUBLIC"}})

	auditor := rbac.NewSymmetricAuditor(graph, nil)

	// User Alice: ADMIN
	_ = auditor.AssignRole(rbac.UserAssignment{
		UserID:  "alice",
		RoleIDs: []string{"ADMIN"},
	})

	// User Bob: EDITOR
	_ = auditor.AssignRole(rbac.UserAssignment{
		UserID:  "bob",
		RoleIDs: []string{"EDITOR"},
	})

	// User Charlie: VIEWER
	_ = auditor.AssignRole(rbac.UserAssignment{
		UserID:  "charlie",
		RoleIDs: []string{"VIEWER"},
	})

	// 1. From User perspective: Alice effective permissions
	alicePerms, err := auditor.GetUserEffectivePermissions("alice")
	if err != nil {
		t.Fatalf("failed to get alice permissions: %v", err)
	}
	expectedAlice := []string{"DELETE_DOCS", "READ_DOCS", "UPDATE_DOCS"}
	if len(alicePerms) != len(expectedAlice) {
		t.Fatalf("expected alice to have %d permissions, got %d: %v", len(expectedAlice), len(alicePerms), alicePerms)
	}
	for i, exp := range expectedAlice {
		if alicePerms[i] != exp {
			t.Errorf("expected %s at index %d, got %s", exp, i, alicePerms[i])
		}
	}

	// 2. From Permission perspective (Reverse Lookup): "Which roles have READ_DOCS?"
	// All roles hold it because it is explicitly granted to each.
	rolesWithRead, err := auditor.GetRolesWithPermission("READ_DOCS")
	if err != nil {
		t.Fatal(err)
	}
	expectedRoles := []string{"ADMIN", "EDITOR", "VIEWER"}
	if len(rolesWithRead) != len(expectedRoles) {
		t.Fatalf("expected %d roles with READ_DOCS, got %d: %v", len(expectedRoles), len(rolesWithRead), rolesWithRead)
	}
	for i, exp := range expectedRoles {
		if rolesWithRead[i] != exp {
			t.Errorf("expected role %s, got %s", exp, rolesWithRead[i])
		}
	}

	// 3. From Permission perspective (Reverse Lookup to Users): "Who can DELETE_DOCS?"
	// Only Alice (ADMIN) holds DELETE_DOCS.
	deleteUsers, err := auditor.GetUsersWithPermission("DELETE_DOCS")
	if err != nil {
		t.Fatal(err)
	}
	if len(deleteUsers) != 1 || deleteUsers[0] != "alice" {
		t.Errorf("expected only alice to have DELETE_DOCS, got: %v", deleteUsers)
	}

	// "Who can READ_DOCS?" -> All three have a role explicitly granting it.
	readUsers, err := auditor.GetUsersWithPermission("READ_DOCS")
	if err != nil {
		t.Fatal(err)
	}
	expectedUsers := []string{"alice", "bob", "charlie"}
	if len(readUsers) != len(expectedUsers) {
		t.Fatalf("expected %d users with READ_DOCS, got %d: %v", len(expectedUsers), len(readUsers), readUsers)
	}
	for i, exp := range expectedUsers {
		if readUsers[i] != exp {
			t.Errorf("expected user %s, got %s", exp, readUsers[i])
		}
	}
}

func TestSymmetric_OrphanedRolesAndRevocation(t *testing.T) {
	graph := rbac.NewRoleGraph(nil)

	// Immutable system role
	_ = graph.AddRole(&rbac.Role{
		ID:          "SYS_CORE",
		IsImmutable: true,
	})

	// Assigned role
	_ = graph.AddRole(&rbac.Role{
		ID: "ACTIVE_ROLE",
	})

	// Orphaned custom role
	_ = graph.AddRole(&rbac.Role{
		ID: "UNUSED_PROJECT_ROLE",
	})

	auditor := rbac.NewSymmetricAuditor(graph, nil)

	_ = auditor.AssignRole(rbac.UserAssignment{
		UserID:  "user_1",
		RoleIDs: []string{"ACTIVE_ROLE"},
	})

	// 1. Find Orphaned Roles: UNUSED_PROJECT_ROLE should be detected, SYS_CORE must be excluded
	orphans := auditor.FindOrphanedRoles()
	if len(orphans) != 1 || orphans[0] != "UNUSED_PROJECT_ROLE" {
		t.Errorf("expected exactly [UNUSED_PROJECT_ROLE] orphaned, got: %v", orphans)
	}

	// 2. Revoke ACTIVE_ROLE from user_1
	err := auditor.RevokeRole("user_1", "ACTIVE_ROLE", "")
	if err != nil {
		t.Fatalf("failed to revoke role: %v", err)
	}

	// Now ACTIVE_ROLE should ALSO become orphaned!
	orphansAfter := auditor.FindOrphanedRoles()
	expectedOrphans := []string{"ACTIVE_ROLE", "UNUSED_PROJECT_ROLE"}
	if len(orphansAfter) != len(expectedOrphans) {
		t.Fatalf("expected %d orphaned roles after revocation, got %d: %v",
			len(expectedOrphans), len(orphansAfter), orphansAfter)
	}
}

func TestSymmetric_ComplianceReport(t *testing.T) {
	graph := rbac.NewRoleGraph(nil)
	sodMgr := rbac.NewSoDManager()

	_ = graph.AddRole(&rbac.Role{ID: "R1"})
	_ = graph.AddRole(&rbac.Role{ID: "R2"})

	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "STATIC_RULE",
		Name:       "Static Conflict",
		Type:       rbac.StaticSoD,
		Roles:      []string{"R1", "R2"},
		MaxAllowed: 1,
	})

	auditor := rbac.NewSymmetricAuditor(graph, sodMgr)

	// Assigning R1 succeeds
	_ = auditor.AssignRole(rbac.UserAssignment{
		UserID:  "alice",
		RoleIDs: []string{"R1"},
	})

	// Attempting to assign R2 to alice should be blocked by SSD validation
	err := auditor.AssignRole(rbac.UserAssignment{
		UserID:  "alice",
		RoleIDs: []string{"R2"},
	})
	if !errors.Is(err, rbac.ErrSSDViolation) {
		t.Errorf("expected ErrSSDViolation when assigning conflicting role, got: %v", err)
	}

	// Generate report
	report := auditor.GenerateComplianceReport(nil)
	if report.TotalRoles != 2 {
		t.Errorf("expected 2 roles in report, got %d", report.TotalRoles)
	}
	if report.TotalAssignments != 1 {
		t.Errorf("expected 1 assignment in report, got %d", report.TotalAssignments)
	}
	if report.ActiveConstraints != 1 {
		t.Errorf("expected 1 constraint in report, got %d", report.ActiveConstraints)
	}
	if len(report.Violations) != 0 {
		t.Errorf("expected 0 violations since SSD blocked the bad assignment, got: %v", report.Violations)
	}
}
