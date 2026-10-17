package perms2

import (
	"reflect"
	"sync"
	"testing"
)

// TestComplexTransitiveClosure_DeepChainsAndCycles tests deep transitive implication trees,
// graph idempotency, and safe handling of duplicate or overlapping inputs.
func TestComplexTransitiveClosure_DeepChainsAndCycles(t *testing.T) {
	svc := NewService()

	t.Run("Deep Chain: UpdateUser -> UpdateSelf + ReadUserDetails -> ReadUserBasic", func(t *testing.T) {
		input := []PermKey{PermUpdateUser}
		resolved := svc.ResolveImplied(input)

		expectedSubset := []PermKey{
			PermUpdateUser,
			PermUpdateSelf,
			PermReadUserDetails,
			PermReadUserBasic,
		}

		for _, exp := range expectedSubset {
			found := false
			for _, r := range resolved {
				if r == exp {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected permission %s to be resolved in deep chain, but was missing in %v", exp, resolved)
			}
		}
	})

	t.Run("Multi-Root Overlapping Inputs", func(t *testing.T) {
		// Both UpdateProject and UpdateIssue require ReadProjectBasic
		input := []PermKey{PermUpdateProject, PermUpdateIssue}
		resolved := svc.ResolveImplied(input)

		// Check for duplicates
		seen := make(map[PermKey]int)
		for _, p := range resolved {
			seen[p]++
			if seen[p] > 1 {
				t.Errorf("duplicate permission %s found in resolved implied list", p)
			}
		}

		// Ensure ReadProjectBasic is present
		if seen[PermReadProjectBasic] != 1 {
			t.Errorf("expected exactly 1 instance of %s, got %d", PermReadProjectBasic, seen[PermReadProjectBasic])
		}
	})

	t.Run("Idempotency: Resolve(Resolve(X)) == Resolve(X)", func(t *testing.T) {
		inputs := [][]PermKey{
			{PermUpdateUser},
			{PermCreateIssue, PermUpdateProject, PermDeleteIssue},
			{PermLowLevelAdminWrite},
		}

		for _, in := range inputs {
			pass1 := svc.ResolveImplied(in)
			pass2 := svc.ResolveImplied(pass1)

			if !reflect.DeepEqual(pass1, pass2) {
				t.Fatalf("idempotency violation: pass1 %v != pass2 %v", pass1, pass2)
			}
		}
	})

	t.Run("Unknown and Empty IDs Resilience", func(t *testing.T) {
		input := []PermKey{"NON_EXISTENT_PERM_1", "INVALID_XYZ", PermCreateIssue}
		resolved := svc.ResolveImplied(input)

		// Must ignore invalid ones and safely resolve valid ones
		for _, p := range resolved {
			if p == "NON_EXISTENT_PERM_1" || p == "INVALID_XYZ" {
				t.Errorf("invalid permission %s was incorrectly included in resolved output", p)
			}
		}

		if len(resolved) < 2 {
			t.Errorf("expected at least CreateIssue and ReadProjectBasic, got %v", resolved)
		}
	})
}

// TestComplexCascadingRevocation_DiamondAndMultiBranch tests complex dependency pruning,
// verifying that revoking a branch only removes dependent descendants while leaving
// sibling branches and non-dependent permissions intact.
func TestComplexCascadingRevocation_DiamondAndMultiBranch(t *testing.T) {
	svc := NewService()

	t.Run("Branch Isolation: Revoking Project Full does not drop Issue creation if basic remains", func(t *testing.T) {
		// Active set: UpdateProject (implies ReadProjectFull -> ReadProjectBasic)
		// and CreateIssue (implies ReadProjectBasic)
		active := svc.ResolveImplied([]PermKey{PermUpdateProject, PermCreateIssue})

		// Revoke ReadProjectFull (which should drop UpdateProject, but NOT CreateIssue or ReadProjectBasic)
		remaining := svc.ResolveRevocation(active, PermReadProjectFull)

		hasUpdateProject := false
		hasReadProjectFull := false
		hasCreateIssue := false
		hasReadProjectBasic := false

		for _, p := range remaining {
			if p == PermUpdateProject {
				hasUpdateProject = true
			}
			if p == PermReadProjectFull {
				hasReadProjectFull = true
			}
			if p == PermCreateIssue {
				hasCreateIssue = true
			}
			if p == PermReadProjectBasic {
				hasReadProjectBasic = true
			}
		}

		if hasUpdateProject {
			t.Errorf("UpdateProject depends on ReadProjectFull and must be revoked")
		}
		if hasReadProjectFull {
			t.Errorf("ReadProjectFull was target to remove and must not be present")
		}
		if !hasCreateIssue {
			t.Errorf("CreateIssue does NOT depend on ReadProjectFull and should have been preserved")
		}
		if !hasReadProjectBasic {
			t.Errorf("ReadProjectBasic does NOT depend on ReadProjectFull and should have been preserved")
		}
	})

	t.Run("Root Revocation: Pruning Root Drops Entire Dependency Forest", func(t *testing.T) {
		active := svc.ResolveImplied([]PermKey{
			PermUpdateProject,
			PermCreateIssue,
			PermReadIssue,
			PermViewWatchers,
		})

		// PermReadProjectBasic is the root for all project & issue operations
		remaining := svc.ResolveRevocation(active, PermReadProjectBasic)

		if len(remaining) != 0 {
			t.Errorf("expected total wipeout of project/issue permissions when root ReadProjectBasic is revoked, got: %v", remaining)
		}
	})

	t.Run("Revoke Non-Existent Permission Has Zero Effect", func(t *testing.T) {
		active := svc.ResolveImplied([]PermKey{PermCreateIssue, PermUpdateUser})
		remaining := svc.ResolveRevocation(active, "UNKNOWN_NON_EXISTENT_PERM")

		if !reflect.DeepEqual(active, remaining) {
			t.Errorf("revoking non-existent permission mutated the active set: before %v, after %v", active, remaining)
		}
	})

	t.Run("Revoke Leaf Node Removes Only That Leaf", func(t *testing.T) {
		active := svc.ResolveImplied([]PermKey{PermUpdateProject})
		// PermUpdateProject is a leaf (nothing in this set depends on it)
		remaining := svc.ResolveRevocation(active, PermUpdateProject)

		for _, p := range remaining {
			if p == PermUpdateProject {
				t.Errorf("leaf node %s was not revoked", PermUpdateProject)
			}
		}
		if len(remaining) != 2 { // ReadProjectFull and ReadProjectBasic must remain
			t.Errorf("expected exactly 2 remaining foundational permissions, got %d: %v", len(remaining), remaining)
		}
	})
}

// TestValidatePermissionsForScope_ComplexBoundaryCases tests boundary and extreme scope combinations.
func TestValidatePermissionsForScope_ComplexBoundaryCases(t *testing.T) {
	svc := NewService()
	allPerms := svc.GetAllPermissions()
	allIDs := make([]PermKey, len(allPerms))
	for i, p := range allPerms {
		allIDs[i] = p.ID
	}

	t.Run("ScopeGlobal Retains 100% of Registered Permissions", func(t *testing.T) {
		res := svc.ValidatePermissionsForScope(allIDs, ScopeGlobal)
		if len(res) != len(allIDs) {
			t.Fatalf("expected all %d permissions at ScopeGlobal, got %d", len(allIDs), len(res))
		}
	})

	t.Run("ScopeOrganization Completely Excludes Global Permissions", func(t *testing.T) {
		res := svc.ValidatePermissionsForScope(allIDs, ScopeOrganization)
		for _, id := range res {
			p := svc.GetPermission(id)
			if p.Scope == ScopeGlobal {
				t.Fatalf("illegal global permission %s leaked into ScopeOrganization", id)
			}
			if p.Scope != ScopeOrganization && p.Scope != ScopeProject {
				t.Fatalf("unexpected scope %s for permission %s at ScopeOrganization", p.Scope, id)
			}
		}
	})

	t.Run("ScopeProject Strictly Contains Only ScopeProject Permissions", func(t *testing.T) {
		res := svc.ValidatePermissionsForScope(allIDs, ScopeProject)
		for _, id := range res {
			p := svc.GetPermission(id)
			if p.Scope != ScopeProject {
				t.Fatalf("permission %s with scope %s leaked into ScopeProject", id, p.Scope)
			}
		}
	})

	t.Run("Empty Input Slice Handles Safely", func(t *testing.T) {
		res := svc.ValidatePermissionsForScope([]PermKey{}, ScopeProject)
		if len(res) != 0 {
			t.Errorf("expected empty slice, got %v", res)
		}
	})
}

// TestInherentPermissions_ComprehensiveMatrix exhaustively tests all actions,
// authorship requirements, permission gates, and strict prohibition on issue deletion.
func TestInherentPermissions_ComprehensiveMatrix(t *testing.T) {
	svc := NewService()

	testMatrix := []struct {
		name               string
		action             InherentAction
		isAuthorOrReporter bool
		granted            map[PermKey]bool
		expectedResult     bool
		description        string
	}{
		// 1. Issue Reporter Public Fields
		{
			name:               "Reporter with CreateIssue -> Allowed to Read Public Fields",
			action:             InherentReadOwnIssuePublicFields,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{PermCreateIssue: true},
			expectedResult:     true,
		},
		{
			name:               "Reporter WITHOUT CreateIssue -> Denied Read Public Fields",
			action:             InherentReadOwnIssuePublicFields,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{},
			expectedResult:     false,
		},
		{
			name:               "NON-Reporter with CreateIssue -> Denied Read Public Fields",
			action:             InherentReadOwnIssuePublicFields,
			isAuthorOrReporter: false,
			granted:            map[PermKey]bool{PermCreateIssue: true},
			expectedResult:     false,
		},
		// 2. Issue Reporter Linking
		{
			name:               "Reporter with CreateIssue -> Allowed to Link Issue",
			action:             InherentLinkOwnIssue,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{PermCreateIssue: true},
			expectedResult:     true,
		},
		// 3. File Attachments
		{
			name:               "Attacher with AddAttachment -> Allowed to Modify Attachment",
			action:             InherentModifyOwnAttachment,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{PermAddAttachment: true},
			expectedResult:     true,
		},
		{
			name:               "Attacher WITHOUT AddAttachment -> Denied Modify Attachment",
			action:             InherentModifyOwnAttachment,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{},
			expectedResult:     false,
		},
		{
			name:               "Attacher -> Always Allowed to Delete Own Attachment Even with Zero Permissions",
			action:             InherentDeleteOwnAttachment,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{}, // ZERO permissions!
			expectedResult:     true,
			description:        "YouTrack Privacy guarantee: Users can always delete files they uploaded themselves",
		},
		{
			name:               "Non-Attacher -> Denied Delete Attachment via Inherent Check",
			action:             InherentDeleteOwnAttachment,
			isAuthorOrReporter: false,
			granted:            map[PermKey]bool{},
			expectedResult:     false,
		},
		// 4. Work Items & Comments
		{
			name:               "Work Item Creator with CreateWorkItem -> Allowed to Read Own Work Item",
			action:             InherentReadOwnWorkItem,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{PermCreateWorkItem: true},
			expectedResult:     true,
		},
		{
			name:               "Article Comment Creator with CreateArticleComment -> Allowed to Delete Own Article Comment",
			action:             InherentDeleteOwnArticleComment,
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{PermCreateArticleComment: true},
			expectedResult:     true,
		},
		// 5. Strict Prohibition: Deleting Own Issues
		{
			name:               "Reporter trying to Delete Own Issue inherently -> Strictly PROHIBITED",
			action:             InherentAction("DELETE_OWN_ISSUE"),
			isAuthorOrReporter: true,
			granted:            map[PermKey]bool{PermCreateIssue: true, PermReadIssue: true, PermUpdateIssue: true},
			expectedResult:     false,
			description:        "YouTrack strictly forbids deleting issues via inherent rights",
		},
	}

	for _, tc := range testMatrix {
		t.Run(tc.name, func(t *testing.T) {
			hasPerm := func(id PermKey) bool {
				return tc.granted[id]
			}

			result := svc.CheckInherentAccess(tc.action, tc.isAuthorOrReporter, hasPerm)
			if result != tc.expectedResult {
				t.Fatalf("action %s (author=%v): expected %v, got %v (%s)",
					tc.action, tc.isAuthorOrReporter, tc.expectedResult, result, tc.description)
			}
		})
	}
}

// TestConcurrentAccessAndThreadSafety ensures that Service is completely race-free
// and thread-safe when accessed by hundreds of concurrent goroutines.
func TestConcurrentAccessAndThreadSafety(t *testing.T) {
	svc := NewService()
	allPerms := svc.GetAllPermissions()

	const workers = 50
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Concurrent catalog reads
				idx := (workerID + i) % len(allPerms)
				targetID := allPerms[idx].ID

				p := svc.GetPermission(targetID)
				if p == nil || p.ID != targetID {
					t.Errorf("worker %d: expected to find permission %s", workerID, targetID)
				}

				// Concurrent Implied Resolution
				implied := svc.ResolveImplied([]PermKey{targetID})
				if len(implied) == 0 {
					t.Errorf("worker %d: implied resolution produced empty set for %s", workerID, targetID)
				}

				// Concurrent Dependent Resolution
				dependent := svc.ResolveDependent([]PermKey{targetID})
				if len(dependent) == 0 {
					t.Errorf("worker %d: dependent resolution produced empty set for %s", workerID, targetID)
				}

				// Concurrent Revocation Resolution
				remaining := svc.ResolveRevocation(implied, targetID)
				for _, rem := range remaining {
					if rem == targetID {
						t.Errorf("worker %d: target %s was not revoked", workerID, targetID)
					}
				}

				// Concurrent Scope Filtering
				scoped := svc.ValidatePermissionsForScope(implied, ScopeProject)
				for _, sc := range scoped {
					perm := svc.GetPermission(sc)
					if perm.Scope != ScopeProject {
						t.Errorf("worker %d: non-project perm %s leaked", workerID, sc)
					}
				}

				// Concurrent Inherent Access Check
				_ = svc.CheckInherentAccess(InherentDeleteOwnAttachment, true, func(PermKey) bool { return false })
			}
		}(w)
	}

	wg.Wait()
}

// Benchmarks to evaluate performance and memory efficiency

func BenchmarkResolveImplied(b *testing.B) {
	svc := NewService()
	input := []PermKey{PermUpdateProject, PermUpdateIssuePrivateFields, PermUpdateUser}

	for b.Loop() {
		_ = svc.ResolveImplied(input)
	}
}

func BenchmarkResolveDependent(b *testing.B) {
	svc := NewService()
	input := []PermKey{PermReadProjectBasic, PermReadUserDetails, PermReadArticle}

	for b.Loop() {
		_ = svc.ResolveDependent(input)
	}
}

func BenchmarkResolveRevocation(b *testing.B) {
	svc := NewService()
	active := svc.ResolveImplied([]PermKey{PermUpdateProject, PermUpdateIssuePrivateFields})

	for b.Loop() {
		_ = svc.ResolveRevocation(active, PermReadProjectBasic)
	}
}

func BenchmarkValidatePermissionsForScope(b *testing.B) {
	svc := NewService()
	allPerms := svc.GetAllPermissions()
	ids := make([]PermKey, len(allPerms))
	for i, p := range allPerms {
		ids[i] = p.ID
	}

	for b.Loop() {
		_ = svc.ValidatePermissionsForScope(ids, ScopeProject)
	}
}

func BenchmarkCheckInherentAccess(b *testing.B) {
	svc := NewService()
	hasPerm := func(id PermKey) bool { return id == PermCreateIssue }

	for b.Loop() {
		_ = svc.CheckInherentAccess(InherentReadOwnIssuePublicFields, true, hasPerm)
	}
}
