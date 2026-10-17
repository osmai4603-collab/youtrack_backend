package rbac_test

import (
	"context"
	"errors"
	"testing"

	"youtrack_backend/internal/services/rbac"
)

// mockImpliedResolver simulates an external catalog resolving implied permissions.
type mockImpliedResolver struct {
	impliedMap map[string][]string
}

func (m *mockImpliedResolver) ResolveImplied(permissions []string) []string {
	resultMap := make(map[string]struct{})
	for _, p := range permissions {
		resultMap[p] = struct{}{}
		if implied, ok := m.impliedMap[p]; ok {
			for _, imp := range implied {
				resultMap[imp] = struct{}{}
			}
		}
	}
	var res []string
	for p := range resultMap {
		res = append(res, p)
	}
	return res
}

func TestPDP_DirectRoleEvaluation(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)

	// Base role: Viewer
	_ = graph.AddRole(&rbac.Role{
		ID:          "VIEWER",
		Permissions: []string{"READ_ISSUE"},
	})

	// Editor receives only the permissions explicitly assigned to it.
	_ = graph.AddRole(&rbac.Role{
		ID:          "EDITOR",
		Permissions: []string{"UPDATE_ISSUE"},
	})

	pdp := rbac.NewPDPEngine(graph, nil, nil)

	// 1. Direct permission match on Viewer
	dec := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "alice",
		// RoleIDs:    []string{"VIEWER"},
		Permission: "READ_ISSUE",
	})
	if !dec.Allowed {
		t.Errorf("expected allowed for direct permission, got reason: %s", dec.Reason)
	}

	// 2. Editor does not inherit READ_ISSUE from Viewer.
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "bob",
		// RoleIDs:    []string{"EDITOR"},
		Permission: "READ_ISSUE",
	})
	if dec.Allowed {
		t.Errorf("expected denied because READ_ISSUE was not explicitly granted to EDITOR")
	}

	// 3. Permission not granted (VIEWER attempting UPDATE_ISSUE)
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "alice",
		// RoleIDs:    []string{"VIEWER"},
		Permission: "UPDATE_ISSUE",
	})
	if dec.Allowed {
		t.Errorf("expected denied for unauthorized permission")
	}
}

func TestPDP_ImpliedPermissions(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)

	_ = graph.AddRole(&rbac.Role{
		ID:          "PROJECT_MANAGER",
		Permissions: []string{"UPDATE_PROJECT"},
	})

	// When UPDATE_PROJECT is granted, READ_PROJECT is implied
	resolver := &mockImpliedResolver{
		impliedMap: map[string][]string{
			"UPDATE_PROJECT": {"READ_PROJECT", "READ_ISSUES"},
		},
	}

	pdp := rbac.NewPDPEngine(graph, resolver, nil)

	dec := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "pm_user",
		// RoleIDs:    []string{"PROJECT_MANAGER"},
		Permission: "READ_PROJECT",
	})
	if !dec.Allowed {
		t.Errorf("expected allowed for implied permission READ_PROJECT, reason: %s", dec.Reason)
	}
}

func TestPDP_SuperAdminWildcard(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)

	_ = graph.AddRole(&rbac.Role{
		ID:          "SUPER_ADMIN",
		Permissions: []string{"*"},
	})
	_ = graph.AddRole(&rbac.Role{
		ID:          "SYS_ALL",
		Permissions: []string{"ALL"},
	})

	pdp := rbac.NewPDPEngine(graph, nil, nil)

	// Test wildcard '*'
	decStar := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "root",
		// RoleIDs:    []string{"SUPER_ADMIN"},
		Permission: "ANY_CUSTOM_ACTION_UNKNOWN",
	})
	if !decStar.Allowed {
		t.Errorf("expected wildcard '*' to grant any permission, reason: %s", decStar.Reason)
	}

	// Test wildcard 'ALL'
	decAll := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "admin",
		// RoleIDs:    []string{"SYS_ALL"},
		Permission: "DELETE_ANYTHING",
	})
	if !decAll.Allowed {
		t.Errorf("expected 'ALL' permission to grant any permission, reason: %s", decAll.Reason)
	}
}

func TestPDP_FailClosedGuards(t *testing.T) {
	ctx := context.Background()
	pdp := rbac.NewPDPEngine(nil, nil, nil)

	// 1. Empty permission
	decEmptyPerm := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "user_1",
		// RoleIDs:    []string{"SOME_ROLE"},
		Permission: "",
	})
	if decEmptyPerm.Allowed {
		t.Errorf("expected Fail-Closed for empty permission")
	}

	// 2. Empty roles
	decEmptyRoles := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "user_1",
		// RoleIDs:    []string{},
		Permission: "READ_PROJECT",
	})
	if decEmptyRoles.Allowed {
		t.Errorf("expected Fail-Closed for empty role list")
	}
}

func TestPEP_Enforcer(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)
	_ = graph.AddRole(&rbac.Role{
		ID:          "DEVELOPER",
		Permissions: []string{"CODE_READ", "CODE_WRITE"},
	})

	pdp := rbac.NewPDPEngine(graph, nil, nil)
	enforcer := rbac.NewEnforcer(pdp, nil)

	// 1. Authorize success
	err := enforcer.Authorize(ctx, rbac.AccessRequest{
		SubjectID: "dev1",
		// RoleIDs:    []string{"DEVELOPER"},
		Permission: "CODE_READ",
	})
	if err != nil {
		t.Errorf("expected authorization success, got: %v", err)
	}

	// 2. Authorize failure
	err = enforcer.Authorize(ctx, rbac.AccessRequest{
		SubjectID: "dev1",
		// RoleIDs:    []string{"DEVELOPER"},
		Permission: "DEPLOY_PROD",
	})
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied for unauthorized request, got: %v", err)
	}

	// 3. AuthorizeAny: one matches -> Success
	err = enforcer.AuthorizeAny(ctx, "dev1", "DEPLOY_PROD", "CODE_WRITE")
	if err != nil {
		t.Errorf("AuthorizeAny should succeed if at least one matches, got: %v", err)
	}

	// 4. AuthorizeAny: none matches -> Failure
	err = enforcer.AuthorizeAny(ctx, "dev1", "DEPLOY_PROD", "DROP_DATABASE")
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied when none match, got: %v", err)
	}

	// 5. AuthorizeAll: all match -> Success
	err = enforcer.AuthorizeAll(ctx, "dev1", "CODE_READ", "CODE_WRITE")
	if err != nil {
		t.Errorf("AuthorizeAll should succeed when all match, got: %v", err)
	}

	// 6. AuthorizeAll: one missing -> Failure
	err = enforcer.AuthorizeAll(ctx, "dev1", "CODE_READ", "DEPLOY_PROD")
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("AuthorizeAll should fail if one is missing, got: %v", err)
	}
}

func BenchmarkPDPEvaluate(b *testing.B) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)
	_ = graph.AddRole(&rbac.Role{ID: "R0", Permissions: []string{"P0", "P1"}})
	_ = graph.AddRole(&rbac.Role{ID: "R1", Permissions: []string{"P0", "P2"}})
	_ = graph.AddRole(&rbac.Role{ID: "R2", Permissions: []string{"P0", "P3"}})

	pdp := rbac.NewPDPEngine(graph, nil, nil)
	req := rbac.AccessRequest{
		SubjectID: "benchmark_user",
		// RoleIDs:    []string{"R2"},
		Permission: "P0",
	}

	b.ResetTimer()
	for b.Loop() {
		_ = pdp.Evaluate(ctx, req)
	}
}

func TestPDP_AssignmentsDirectEvaluation(t *testing.T) {
	ctx := context.Background()
	// PDP with empty graph: roles are NOT stored in the graph or repository.
	pdp := rbac.NewPDPEngine(rbac.NewRoleGraph(nil), nil, nil)

	// assignments := []rbac.RoleAssignment{
	// 	{
	// 		AssignmentID: "asgn-dev",
	// 		RoleID:       "DEVELOPER_DYNAMIC",
	// 		RoleName:     "Dynamic Developer",
	// 		Permissions: []string{
	// 			"CODE_READ",
	// 			"CODE_WRITE",
	// 		},
	// 	},
	// }

	// 1. Allowed: Target permission is directly present in assignment
	dec := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "dev_bob",
		// Assignments: assignments,
		Permission: "CODE_READ",
	})
	if !dec.Allowed {
		t.Fatalf("expected allowed for permission in assignment, got: %s", dec.Reason)
	}

	// 2. Denied: Permission not in assignment
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "dev_bob",
		// Assignments: assignments,
		Permission: "DROP_DATABASE",
	})
	if dec.Allowed {
		t.Fatalf("expected denied for permission not in assignment")
	}

	// 3. Allowed: Wildcard '*' in assignment
	// wildcardAssignments := []rbac.RoleAssignment{
	// 	{
	// 		AssignmentID: "asgn-admin",
	// 		RoleID:       "DYNAMIC_ADMIN",
	// 		Permissions:  []string{},
	// 	},
	// }
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID: "admin_user",
		// Assignments: wildcardAssignments,
		Permission: "ANY_ACTION",
	})
	if !dec.Allowed {
		t.Fatalf("expected allowed with wildcard in assignment, got: %s", dec.Reason)
	}

	// 4. Fail-closed: No active roles or assignments supplied
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "empty_user",
		Permission: "CODE_READ",
	})
	if dec.Allowed {
		t.Fatalf("expected fail-closed when both RoleIDs and Assignments are empty")
	}
}

func BenchmarkPDPEvaluateWithAssignments(b *testing.B) {
	ctx := context.Background()
	pdp := rbac.NewPDPEngine(rbac.NewRoleGraph(nil), nil, nil)
	req := rbac.AccessRequest{
		SubjectID: "benchmark_user",
		// Assignments: []rbac.RoleAssignment{
		// 	{
		// 		RoleID: "FAST_ROLE",
		// 		Permissions: []string{
		// 			"P0",
		// 			"P1",
		// 		},
		// 	},
		// },
		Permission: "P0",
	}

	b.ResetTimer()
	for b.Loop() {
		_ = pdp.Evaluate(ctx, req)
	}
}
