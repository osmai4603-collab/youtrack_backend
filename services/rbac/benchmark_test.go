package rbac_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"youtrack_backend/internal/services/rbac"
)

// BenchmarkRolePermissionLookup benchmarks resolving direct role permissions.
func BenchmarkRolePermissionLookup(b *testing.B) {
	graph := rbac.NewRoleGraph(nil)
	const roles = 10
	for i := 0; i < roles; i++ {
		_ = graph.AddRole(&rbac.Role{
			ID:          fmt.Sprintf("ROLE_%d", i),
			Permissions: []string{fmt.Sprintf("PERM_%d_A", i), fmt.Sprintf("PERM_%d_B", i)},
		})
	}

	roleID := fmt.Sprintf("ROLE_%d", roles-1)

	b.ReportAllocs()
	for b.Loop() {
		_, _ = graph.GetEffectivePermissions(roleID)
	}
}

// BenchmarkPDPEvaluateParallel measures PDP evaluation throughput across multiple concurrent goroutines.
func BenchmarkPDPEvaluateParallel(b *testing.B) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)

	_ = graph.AddRole(&rbac.Role{
		ID:          "LEAD_DEV",
		Permissions: []string{"GIT_PUSH", "CODE_REVIEW"},
	})
	_ = graph.AddRole(&rbac.Role{
		ID:          "ARCHITECT",
		Permissions: []string{"GIT_PUSH", "DEPLOY_PROD"},
	})

	pdp := rbac.NewPDPEngine(graph, nil, nil)
	req := rbac.AccessRequest{
		SubjectID:  "architect_user",
		// RoleIDs:    []string{"ARCHITECT"},
		Permission: "GIT_PUSH",
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dec := pdp.Evaluate(ctx, req)
			if !dec.Allowed {
				b.Errorf("expected allowed decision in benchmark")
			}
		}
	})
}

// BenchmarkSoDValidation measures the performance of Static Separation of Duties (SSD) validation.
func BenchmarkSoDValidation(b *testing.B) {
	graph := rbac.NewRoleGraph(nil)
	sodMgr := rbac.NewSoDManager()

	_ = graph.AddRole(&rbac.Role{ID: "FINANCE_1"})
	_ = graph.AddRole(&rbac.Role{ID: "FINANCE_2"})
	_ = graph.AddRole(&rbac.Role{ID: "FINANCE_3"})

	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "SSD_RULE",
		Name:       "Finance SSD",
		Type:       rbac.StaticSoD,
		Roles:      []string{"FINANCE_1", "FINANCE_2"},
		MaxAllowed: 1,
	})

	existing := []string{"FINANCE_1"}

	b.ReportAllocs()
	for b.Loop() {
		_ = sodMgr.ValidateUserAssignment(existing, "FINANCE_3")
	}
}

// BenchmarkReversePermissionLookup measures reverse permission lookup across 1,000 users.
func BenchmarkReversePermissionLookup(b *testing.B) {
	graph := rbac.NewRoleGraph(nil)
	_ = graph.AddRole(&rbac.Role{ID: "BASE_VIEWER", Permissions: []string{"AUDIT_LOG_READ"}})
	_ = graph.AddRole(&rbac.Role{ID: "SEC_ANALYST", Permissions: []string{"AUDIT_LOG_READ"}})
	_ = graph.AddRole(&rbac.Role{ID: "CISO", Permissions: []string{"AUDIT_LOG_READ"}})

	auditor := rbac.NewSymmetricAuditor(graph, nil)

	// Populate 1,000 assignments
	for i := 0; i < 1000; i++ {
		role := "BASE_VIEWER"
		if i%5 == 0 {
			role = "SEC_ANALYST"
		} else if i%20 == 0 {
			role = "CISO"
		}
		_ = auditor.AssignRole(rbac.UserAssignment{
			UserID:  fmt.Sprintf("user_%d", i),
			RoleIDs: []string{role},
		})
	}

	b.ReportAllocs()
	for b.Loop() {
		users, err := auditor.GetUsersWithPermission("AUDIT_LOG_READ")
		if err != nil || len(users) == 0 {
			b.Fatalf("failed reverse lookup in benchmark")
		}
	}
}

// BenchmarkSessionRoleActivation measures JIT role activation and active permissions resolution.
func BenchmarkSessionRoleActivation(b *testing.B) {
	graph := rbac.NewRoleGraph(nil)
	_ = graph.AddRole(&rbac.Role{ID: "OPS", Permissions: []string{"SERVER_RESTART"}})

	sm := rbac.NewSessionManager(nil)
	sm.SetGraph(graph)
	sess, _ := sm.CreateSession("admin", []string{}, time.Hour)

	b.ReportAllocs()
	for b.Loop() {
		_ = sm.ActivateRole(sess.ID, "OPS", 5*time.Minute)
		_, _ = sm.GetActivePermissions(sess.ID)
	}
}
