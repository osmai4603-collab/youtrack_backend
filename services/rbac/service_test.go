package rbac_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"youtrack_backend/internal/services/rbac"
)

// mockRoleRepo implements rbac.RoleRepository for testing
type mockRoleRepo struct {
	mu    sync.RWMutex
	roles map[string]*rbac.Role
}

func newMockRoleRepo() *mockRoleRepo {
	return &mockRoleRepo{
		roles: make(map[string]*rbac.Role),
	}
}

func (m *mockRoleRepo) GetRole(ctx context.Context, id string) (*rbac.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.roles[id]
	if !ok {
		return nil, rbac.ErrRoleNotFound
	}
	return r, nil
}

func (m *mockRoleRepo) GetRolesWhereIDs(ctx context.Context, ids []string) ([]*rbac.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*rbac.Role
	for _, id := range ids {
		if r, ok := m.roles[id]; ok {
			res = append(res, r)
		}
	}
	return res, nil
}

func (m *mockRoleRepo) CreateRole(ctx context.Context, role *rbac.Role) (*rbac.Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.roles[role.ID]; ok {
		return nil, rbac.ErrRoleAlreadyExists
	}
	m.roles[role.ID] = role
	return role, nil
}

func (m *mockRoleRepo) DeleteRole(ctx context.Context, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, roleID)
	return nil
}

func TestRBACService_EndToEndIntegration(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	catalog := &mockImpliedResolver{
		impliedMap: map[string][]string{
			"WRITE_ISSUE": {"READ_ISSUE"},
		},
	}

	svc := rbac.NewService(repo, catalog, nil)

	// 1. Policy Administration Point (PAP): Role creation
	viewerRole := &rbac.Role{
		ID:          "VIEWER",
		Name:        "Issue Viewer",
		Permissions: []string{"READ_ISSUE"},
	}
	createdViewer, err := svc.CreateRole(ctx, viewerRole)
	if err != nil {
		t.Fatalf("failed to create viewer role: %v", err)
	}
	if createdViewer.ID != "VIEWER" {
		t.Errorf("expected role ID VIEWER, got %s", createdViewer.ID)
	}

	editorRole := &rbac.Role{
		ID:          "EDITOR",
		Name:        "Issue Editor",
		Permissions: []string{"READ_ISSUE", "WRITE_ISSUE"},
	}
	_, err = svc.CreateRole(ctx, editorRole)
	if err != nil {
		t.Fatalf("failed to create editor role: %v", err)
	}

	// 2. Separation of Duties (SoD) registration
	ssdRule := rbac.SoDConstraint{
		ID:         "SSD_CREATOR_APPROVER",
		Name:       "Creator vs Approver",
		Type:       rbac.StaticSoD,
		Roles:      []string{"CREATOR", "APPROVER"},
		MaxAllowed: 1,
	}
	if err := svc.AddSoDConstraint(ctx, ssdRule); err != nil {
		t.Fatalf("failed to add SSD constraint: %v", err)
	}

	_, _ = svc.CreateRole(ctx, &rbac.Role{ID: "CREATOR"})
	_, _ = svc.CreateRole(ctx, &rbac.Role{ID: "APPROVER"})

	// 3. User Provisioning & SSD Enforcement
	// Assign CREATOR to Alice -> Succeeds
	err = svc.AssignRoleToUser(ctx, rbac.UserAssignment{
		UserID:  "alice",
		RoleIDs: []string{"CREATOR"},
	})
	if err != nil {
		t.Fatalf("failed to assign role to alice: %v", err)
	}

	// Attempt to assign APPROVER to Alice -> Must fail with ErrSSDViolation
	err = svc.AssignRoleToUser(ctx, rbac.UserAssignment{
		UserID:  "alice",
		RoleIDs: []string{"APPROVER"},
	})
	if !errors.Is(err, rbac.ErrSSDViolation) {
		t.Errorf("expected ErrSSDViolation, got: %v", err)
	}

	// Assign EDITOR to Bob
	err = svc.AssignRoleToUser(ctx, rbac.UserAssignment{
		UserID:  "bob",
		RoleIDs: []string{"EDITOR"},
	})
	if err != nil {
		t.Fatalf("failed to assign editor to bob: %v", err)
	}

	// 4. Session Management & JIT Access
	sess, err := svc.CreateSession(ctx, "bob", []string{"EDITOR"}, 2*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if sess.UserID != "bob" {
		t.Errorf("expected session user bob, got %s", sess.UserID)
	}

	// 5. Enforcement (PEP / PDP)
	// Bob has EDITOR, which explicitly grants READ_ISSUE and WRITE_ISSUE.
	err = svc.Authorize(ctx, rbac.AccessRequest{
		SubjectID: "bob",
		//RoleIDs:    []string{"EDITOR"},
		Permission: "WRITE_ISSUE",
	})
	if err != nil {
		t.Errorf("bob should be authorized for WRITE_ISSUE: %v", err)
	}

	err = svc.Authorize(ctx, rbac.AccessRequest{
		SubjectID: "bob",
		//RoleIDs:    []string{"EDITOR"},
		Permission: "READ_ISSUE",
	})
	if err != nil {
		t.Errorf("bob should be authorized for explicitly granted READ_ISSUE: %v", err)
	}

	// Unauthorized action
	err = svc.Authorize(ctx, rbac.AccessRequest{
		SubjectID: "bob",
		//RoleIDs:    []string{"EDITOR"},
		Permission: "DELETE_PROJECT",
	})
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied for unassigned action, got: %v", err)
	}

	// 6. Symmetric RBAC & Auditing
	// "Who can READ_ISSUE?" -> Bob holds a role explicitly granting READ_ISSUE.
	usersWithRead, err := svc.GetUsersWithPermission(ctx, "READ_ISSUE")
	if err != nil {
		t.Fatal(err)
	}
	if len(usersWithRead) != 1 || usersWithRead[0] != "bob" {
		t.Errorf("expected bob in users with READ_ISSUE, got: %v", usersWithRead)
	}

	// Compliance Report
	report, err := svc.GenerateComplianceReport(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if report.ActiveConstraints != 1 {
		t.Errorf("expected 1 active constraint, got %d", report.ActiveConstraints)
	}
}

func TestRBACService_Concurrency(t *testing.T) {
	ctx := context.Background()
	svc := rbac.NewService(nil, nil, nil)

	_, _ = svc.CreateRole(ctx, &rbac.Role{
		ID:          "COMMON_ROLE",
		Permissions: []string{"PERM_1", "PERM_2"},
	})

	var wg sync.WaitGroup
	const routines = 25

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			uID := fmt.Sprintf("concurrent_user_%d", id)
			_ = svc.AssignRoleToUser(ctx, rbac.UserAssignment{
				UserID:  uID,
				RoleIDs: []string{"COMMON_ROLE"},
			})
			_ = svc.AuthorizePermission(ctx, uID, "PERM_1")
			sess, err := svc.CreateSession(ctx, uID, []string{"COMMON_ROLE"}, time.Minute)
			if err == nil {
				_ = svc.TerminateSession(ctx, sess.ID)
			}
			_, _ = svc.GetUserEffectivePermissions(ctx, uID)
		}(i)
	}

	wg.Wait()
}

func TestRBACService_ComprehensiveFacadeCoverage(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := rbac.NewService(repo, nil, nil)

	// Context Manager
	cm := svc.GetContextManager()
	if cm == nil {
		t.Fatal("expected non-nil context manager from svc")
	}

	// Create Role
	roleR1 := &rbac.Role{ID: "R1", Permissions: []string{"PERM_R1"}}
	_, err := svc.CreateRole(ctx, roleR1)
	if err != nil {
		t.Fatal(err)
	}

	roleR2 := &rbac.Role{ID: "R2", Permissions: []string{"PERM_R2"}}
	_, err = svc.CreateRole(ctx, roleR2)
	if err != nil {
		t.Fatal(err)
	}

	// GetRole
	gotR1, err := svc.GetRole(ctx, "R1")
	if err != nil || gotR1.ID != "R1" {
		t.Fatalf("failed GetRole: %v", err)
	}

	// GetAllRoles
	allRoles, err := svc.GetAllRoles(ctx)
	if err != nil || len(allRoles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(allRoles))
	}

	// UpdateRole
	gotR1.Description = "Updated R1"
	_, err = svc.UpdateRole(ctx, gotR1)
	if err != nil {
		t.Fatalf("failed UpdateRole: %v", err)
	}

	effPerms, err := svc.GetEffectivePermissions(ctx, "R2")
	if err != nil || len(effPerms) != 1 || effPerms[0] != "PERM_R2" {
		t.Fatalf("expected only R2's explicitly granted permission, got %v (err=%v)", effPerms, err)
	}

	// SoD Constraints
	sodRule := rbac.SoDConstraint{
		ID:         "RULE_X",
		Name:       "Rule X",
		Type:       rbac.StaticSoD,
		Roles:      []string{"R1", "R2"},
		MaxAllowed: 1,
	}
	_ = svc.AddSoDConstraint(ctx, sodRule)
	sodGot, err := svc.GetSoDConstraint(ctx, "RULE_X")
	if err != nil || sodGot.ID != "RULE_X" {
		t.Fatalf("failed GetSoDConstraint: %v", err)
	}
	constraints, err := svc.ListSoDConstraints(ctx)
	if err != nil || len(constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(constraints))
	}
	_ = svc.RemoveSoDConstraint(ctx, "RULE_X")

	// User Assignment & Revocation
	ua := rbac.UserAssignment{UserID: "u100", RoleIDs: []string{"R1"}}
	_ = svc.AssignRoleToUser(ctx, ua)
	userAssignments, err := svc.GetUserAssignments(ctx, "u100")
	if err != nil || len(userAssignments) != 1 {
		t.Fatalf("expected 1 user assignment, got %d", len(userAssignments))
	}
	_ = svc.RevokeRoleFromUser(ctx, "u100", "R1", "")

	// Session management
	sess, err := svc.CreateSession(ctx, "u200", []string{"R1"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_ = svc.ActivateRoleInSession(ctx, sess.ID, "R2", time.Minute)
	_ = svc.DeactivateRoleInSession(ctx, sess.ID, "R2")
	gotSess, err := svc.GetSession(ctx, sess.ID)
	if err != nil || gotSess.ID != sess.ID {
		t.Fatalf("failed GetSession: %v", err)
	}

	// AuthorizeAll & AuthorizeAny & Evaluate
	_ = svc.AuthorizeAny(ctx, "u200", "PERM_R1", "NON_EXISTENT")
	_ = svc.AuthorizeAll(ctx, "u200", "PERM_R1")
	dec := svc.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "u200",
		//RoleIDs:    []string{"R1"},
		Permission: "PERM_R1",
	})
	if !dec.Allowed {
		t.Errorf("expected Evaluate to allow")
	}

	// Symmetric queries
	rolesWithP, err := svc.GetRolesWithPermission(ctx, "PERM_R1")
	if err != nil || len(rolesWithP) != 1 {
		t.Fatalf("expected 1 role with PERM_R1, got %v", rolesWithP)
	}
	orphans, err := svc.FindOrphanedRoles(ctx)
	if err != nil || len(orphans) != 2 {
		t.Fatalf("expected 2 orphaned roles, got %v", orphans)
	}

	// DeleteRole
	if err := svc.DeleteRole(ctx, "R1"); err != nil {
		t.Fatalf("failed DeleteRole: %v", err)
	}
}
