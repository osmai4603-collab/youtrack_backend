package rbac_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"youtrack_backend/internal/services/rbac"
)

func TestRolePermissionsAreIndependent(t *testing.T) {
	roles := rbac.NewRoleGraph(nil)
	if err := roles.AddRole(&rbac.Role{ID: "VIEWER", Permissions: []string{"READ"}}); err != nil {
		t.Fatal(err)
	}
	if err := roles.AddRole(&rbac.Role{ID: "EDITOR", Permissions: []string{"WRITE"}}); err != nil {
		t.Fatal(err)
	}

	permissions, err := roles.GetEffectivePermissions("EDITOR")
	if err != nil {
		t.Fatal(err)
	}
	if len(permissions) != 1 || permissions[0] != "WRITE" {
		t.Fatalf("expected only explicitly granted WRITE permission, got %v", permissions)
	}
	if hasRead, err := roles.HasPermission("EDITOR", "READ"); err != nil || hasRead {
		t.Errorf("EDITOR must not inherit READ from VIEWER (hasRead=%v, err=%v)", hasRead, err)
	}
}

func TestRoleGraph_ImmutableRoleProtection(t *testing.T) {
	roles := rbac.NewRoleGraph(nil)
	if err := roles.AddRole(&rbac.Role{
		ID:          "SYSTEM_ADMIN",
		Name:        "System Administrator",
		IsImmutable: true,
		Permissions: []string{"ALL"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := roles.DeleteRole("SYSTEM_ADMIN"); !errors.Is(err, rbac.ErrRoleImmutable) {
		t.Errorf("expected ErrRoleImmutable when deleting immutable role, got: %v", err)
	}
	if err := roles.UpdateRole(&rbac.Role{ID: "SYSTEM_ADMIN", Name: "Modified Admin"}); !errors.Is(err, rbac.ErrRoleImmutable) {
		t.Errorf("expected ErrRoleImmutable when updating immutable role, got: %v", err)
	}
}

func TestRoleGraph_CloneIndependence(t *testing.T) {
	roles := rbac.NewRoleGraph(nil)
	if err := roles.AddRole(&rbac.Role{ID: "ROLE", Permissions: []string{"READ"}}); err != nil {
		t.Fatal(err)
	}

	clone := roles.Clone()
	if err := clone.AddRole(&rbac.Role{ID: "NEW_ROLE", Permissions: []string{"WRITE"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := roles.GetRole("NEW_ROLE"); !errors.Is(err, rbac.ErrRoleNotFound) {
		t.Errorf("original registry should not be affected by changes to clone")
	}
	if _, err := clone.GetRole("NEW_ROLE"); err != nil {
		t.Errorf("clone should contain NEW_ROLE: %v", err)
	}
}

func TestRoleGraph_ConcurrentOperations(t *testing.T) {
	roles := rbac.NewRoleGraph(nil)
	if err := roles.AddRole(&rbac.Role{ID: "ROOT", Permissions: []string{"ROOT_PERM"}}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	const workers = 30
	for i := 0; i < workers; i++ {
		wg.Add(1)
		roleID := fmt.Sprintf("WORKER_ROLE_%d", i)
		go func(id string) {
			defer wg.Done()
			if err := roles.AddRole(&rbac.Role{ID: id, Permissions: []string{"WORKER_PERM"}}); err != nil {
				t.Errorf("failed to add %s: %v", id, err)
				return
			}
			if _, err := roles.GetEffectivePermissions(id); err != nil {
				t.Errorf("failed to retrieve permissions for %s: %v", id, err)
			}
		}(roleID)
	}
	wg.Wait()

	if got, want := len(roles.GetAllRoles()), workers+1; got != want {
		t.Errorf("expected %d roles, got %d", want, got)
	}
}

func BenchmarkGetEffectivePermissions(b *testing.B) {
	roles := rbac.NewRoleGraph(nil)
	_ = roles.AddRole(&rbac.Role{ID: "ROLE", Permissions: []string{"P0", "P1", "P2", "P3", "P4"}})

	b.ResetTimer()
	for b.Loop() {
		_, _ = roles.GetEffectivePermissions("ROLE")
	}
}
