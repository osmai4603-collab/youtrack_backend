package rbac_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"youtrack_backend/internal/services/rbac"
)

func TestSession_Lifecycle(t *testing.T) {
	sm := rbac.NewSessionManager(nil)

	// 1. Create session
	sess, err := sm.CreateSession("alice", []string{"DEVELOPER"}, time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if sess.ID == "" || sess.UserID != "alice" {
		t.Errorf("invalid session created: %+v", sess)
	}

	// 2. Get session
	retrieved, err := sm.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if retrieved.ID != sess.ID {
		t.Errorf("expected session ID %s, got %s", sess.ID, retrieved.ID)
	}

	// 3. Terminate session
	if err := sm.TerminateSession(sess.ID); err != nil {
		t.Fatalf("failed to terminate session: %v", err)
	}

	// 4. Verify gone
	if _, err := sm.GetSession(sess.ID); !errors.Is(err, rbac.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound after termination, got: %v", err)
	}
}

func TestSession_Expiration(t *testing.T) {
	sm := rbac.NewSessionManager(nil)

	// Create session with 20ms TTL
	sess, err := sm.CreateSession("bob", []string{"VIEWER"}, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for expiration
	time.Sleep(30 * time.Millisecond)

	// GetSession must detect expiration and return ErrSessionExpired
	_, err = sm.GetSession(sess.ID)
	if !errors.Is(err, rbac.ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired for elapsed session, got: %v", err)
	}
}

func TestSession_JITRoleActivationAndAutoEviction(t *testing.T) {
	sm := rbac.NewSessionManager(nil)

	sess, err := sm.CreateSession("charlie", []string{"BASE_USER"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Activate JIT elevated role for 50 milliseconds
	err = sm.ActivateRole(sess.ID, "TEMP_OPERATOR", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to activate JIT role: %v", err)
	}

	// Immediately, active roles must include TEMP_OPERATOR
	roles, err := sm.GetActiveRoles(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 active roles, got %d: %v", len(roles), roles)
	}

	// Wait for JIT role lease to expire
	time.Sleep(70 * time.Millisecond)

	// Now TEMP_OPERATOR should be automatically evicted!
	rolesAfter, err := sm.GetActiveRoles(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rolesAfter) != 1 || rolesAfter[0] != "BASE_USER" {
		t.Errorf("expected only BASE_USER after JIT expiry, got: %v", rolesAfter)
	}
}

func TestSession_DSDEnforcement(t *testing.T) {
	graph := rbac.NewRoleGraph(nil)
	sodMgr := rbac.NewSoDManager()
	sm := rbac.NewSessionManager(nil)
	sm.SetGraph(graph)
	sm.SetSoDManager(sodMgr)

	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "DSD_AUDIT",
		Name:       "Treasurer vs Auditor",
		Type:       rbac.DynamicSoD,
		Roles:      []string{"TREASURER", "AUDITOR"},
		MaxAllowed: 1,
	})

	// 1. Creating session with both conflicting roles should fail
	_, err := sm.CreateSession("david", []string{"TREASURER", "AUDITOR"}, time.Hour)
	if err == nil {
		t.Errorf("creating session with DSD violating initial roles should fail")
	}

	// 2. Create session with TREASURER
	sess, err := sm.CreateSession("david", []string{"TREASURER"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Activating AUDITOR in same session must be rejected with ErrDSDViolation
	err = sm.ActivateRole(sess.ID, "AUDITOR", 0)
	if !errors.Is(err, rbac.ErrDSDViolation) {
		t.Errorf("expected ErrDSDViolation when activating conflicting role, got: %v", err)
	}

	// 4. Deactivate TREASURER
	if err := sm.DeactivateRole(sess.ID, "TREASURER"); err != nil {
		t.Fatal(err)
	}

	// 5. Now activating AUDITOR should succeed
	err = sm.ActivateRole(sess.ID, "AUDITOR", 0)
	if err != nil {
		t.Errorf("activating AUDITOR after deactivating TREASURER should succeed, got: %v", err)
	}
}

func TestSession_GetActivePermissions(t *testing.T) {
	graph := rbac.NewRoleGraph(nil)
	_ = graph.AddRole(&rbac.Role{
		ID:          "DEV",
		Permissions: []string{"GIT_PUSH"},
	})
	_ = graph.AddRole(&rbac.Role{
		ID:          "LEAD",
		Permissions: []string{"MERGE_PROD"},
	})

	sm := rbac.NewSessionManager(nil)
	sm.SetGraph(graph)

	sess, err := sm.CreateSession("elena", []string{"LEAD"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	perms, err := sm.GetActivePermissions(sess.ID)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"MERGE_PROD"}
	if len(perms) != len(expected) {
		t.Fatalf("expected %d permissions, got %d: %v", len(expected), len(perms), perms)
	}
	for i, exp := range expected {
		if perms[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, perms[i])
		}
	}
}

func TestSession_Concurrency(t *testing.T) {
	sm := rbac.NewSessionManager(nil)

	var wg sync.WaitGroup
	const users = 20

	for i := 0; i < users; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			uID := fmt.Sprintf("user_%d", idx)
			sess, err := sm.CreateSession(uID, []string{"ROLE_1"}, time.Minute)
			if err != nil {
				t.Errorf("failed concurrent session creation: %v", err)
				return
			}
			_ = sm.ActivateRole(sess.ID, "ROLE_2", 10*time.Second)
			_, _ = sm.GetActiveRoles(sess.ID)
			_ = sm.DeactivateRole(sess.ID, "ROLE_1")
		}(i)
	}

	wg.Wait()
}
