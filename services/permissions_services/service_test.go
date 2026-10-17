package perms2

import (
	"testing"
)

func TestBuildDefaultCatalog(t *testing.T) {
	svc := NewService()
	perms := svc.GetAllPermissions()

	if len(perms) < 40 {
		t.Fatalf("expected at least 40 permissions, got %d", len(perms))
	}

	for _, p := range perms {
		if p.ID == "" {
			t.Errorf("permission with empty ID found")
		}
		if p.DisplayName == "" {
			t.Errorf("permission %s has empty DisplayName", p.ID)
		}
		if p.Description == "" {
			t.Errorf("permission %s has empty Description", p.ID)
		}
		if p.Entity == "" {
			t.Errorf("permission %s has empty Entity", p.ID)
		}
		if p.Scope == "" {
			t.Errorf("permission %s has empty Scope", p.ID)
		}
		if p.Operation == "" {
			t.Errorf("permission %s has empty Operation", p.ID)
		}
	}
}

func TestResolveImplied(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name     string
		input    []PermKey
		expected []PermKey
	}{
		{
			name:     "Create Issue implies Read Project Basic",
			input:    []PermKey{PermCreateIssue},
			expected: []PermKey{PermCreateIssue, PermReadProjectBasic}, // CREATE_ISSUE < READ_PROJECT_BASIC
		},
		{
			name:  "Update User implies Update Profile, Read User Details, and Read User Basic",
			input: []PermKey{PermUpdateUser},
			expected: []PermKey{
				PermReadUserDetails, // READ_USER
				PermReadUserBasic,   // READ_USER_BASIC
				PermUpdateSelf,      // UPDATE_PROFILE
				PermUpdateUser,      // UPDATE_USER
			},
		},
		{
			name:  "Update Issue Private Fields implies Read Issue Private Fields, Update Issue, and Read Project Basic",
			input: []PermKey{PermUpdateIssuePrivateFields},
			expected: []PermKey{
				PermReadIssuePrivateFields,   // PRIVATE_READ_ISSUE
				PermUpdateIssuePrivateFields, // PRIVATE_UPDATE_ISSUE
				PermReadProjectBasic,         // READ_PROJECT_BASIC
				PermUpdateIssue,              // UPDATE_ISSUE
			},
		},
		{
			name:  "Update Project implies Read Project Full and Read Project Basic",
			input: []PermKey{PermUpdateProject},
			expected: []PermKey{
				PermReadProjectFull,  // READ_PROJECT
				PermReadProjectBasic, // READ_PROJECT_BASIC
				PermUpdateProject,    // UPDATE_PROJECT
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := svc.ResolveImplied(tc.input)
			if len(res) != len(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, res)
			}
			for i, exp := range tc.expected {
				if res[i] != exp {
					t.Errorf("at index %d: expected %s, got %s", i, exp, res[i])
				}
			}
		})
	}
}

func TestResolveDependent(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name     string
		input    []PermKey
		expected []PermKey
	}{
		{
			name:     "Read Project Full is depended upon by Update Project and Delete Project",
			input:    []PermKey{PermReadProjectFull},
			expected: []PermKey{PermDeleteProject, PermReadProjectFull, PermUpdateProject},
		},
		{
			name:     "Read User Details is depended upon by Update User and Delete User",
			input:    []PermKey{PermReadUserDetails},
			expected: []PermKey{PermDeleteUser, PermReadUserDetails, PermUpdateUser},
		},
		{
			name:     "Read Article is depended upon by Create, Delete, and Update Article",
			input:    []PermKey{PermReadArticle},
			expected: []PermKey{PermCreateArticle, PermDeleteArticle, PermReadArticle, PermUpdateArticle},
		},
		{
			name:     "Read Article Comment is depended upon by Create, Delete, and Update Article Comment",
			input:    []PermKey{PermReadArticleComment},
			expected: []PermKey{PermCreateArticleComment, PermDeleteArticleComment, PermReadArticleComment, PermUpdateArticleComment},
		},
		{
			name:     "Leaf permission has only itself as dependent",
			input:    []PermKey{PermDeleteArticle},
			expected: []PermKey{PermDeleteArticle},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := svc.ResolveDependent(tc.input)
			if len(res) != len(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, res)
			}
			for i, exp := range tc.expected {
				if res[i] != exp {
					t.Errorf("at index %d: expected %s, got %s", i, exp, res[i])
				}
			}
		})
	}
}

func TestGetDependentAndImpliedPermissions(t *testing.T) {
	svc := NewService()

	// Direct Dependent
	depReadProj := svc.GetDependentPermissions(PermReadProjectFull)
	expectedDep := []PermKey{PermDeleteProject, PermUpdateProject}
	if len(depReadProj) != len(expectedDep) {
		t.Fatalf("expected %v, got %v", expectedDep, depReadProj)
	}
	for i, exp := range expectedDep {
		if depReadProj[i] != exp {
			t.Errorf("expected %s, got %s", exp, depReadProj[i])
		}
	}

	// Leaf has empty dependent perms
	if len(svc.GetDependentPermissions(PermDeleteArticle)) != 0 {
		t.Errorf("expected empty dependent perms for leaf node")
	}

	// Unknown permission returns nil
	if svc.GetDependentPermissions("NON_EXISTENT") != nil {
		t.Errorf("expected nil for non-existent permission")
	}

	// Direct Implied
	impCreateIssue := svc.GetImpliedPermissions(PermCreateIssue)
	if len(impCreateIssue) != 1 || impCreateIssue[0] != PermReadProjectBasic {
		t.Errorf("expected [READ_PROJECT_BASIC], got %v", impCreateIssue)
	}

	// Root has empty implied perms
	if len(svc.GetImpliedPermissions(PermReadProjectBasic)) != 0 {
		t.Errorf("expected empty implied perms for root node")
	}

	// Unknown permission returns nil
	if svc.GetImpliedPermissions("NON_EXISTENT") != nil {
		t.Errorf("expected nil for non-existent permission")
	}
}

func TestDependsOn(t *testing.T) {
	svc := NewService()

	// Direct dependency
	if !svc.DependsOn(PermUpdateProject, PermReadProjectFull) {
		t.Errorf("UpdateProject should depend on ReadProjectFull")
	}

	// Transitive dependency
	if !svc.DependsOn(PermUpdateProject, PermReadProjectBasic) {
		t.Errorf("UpdateProject should transitively depend on ReadProjectBasic")
	}

	if !svc.DependsOn(PermUpdateUser, PermReadUserBasic) {
		t.Errorf("UpdateUser should transitively depend on ReadUserBasic")
	}

	// Reverse should be false
	if svc.DependsOn(PermReadProjectBasic, PermUpdateProject) {
		t.Errorf("ReadProjectBasic should not depend on UpdateProject")
	}

	// Self dependency should be false
	if svc.DependsOn(PermUpdateProject, PermUpdateProject) {
		t.Errorf("Permission should not depend on itself")
	}

	// Unrelated permissions
	if svc.DependsOn(PermCreateIssue, PermReadArticle) {
		t.Errorf("CreateIssue should not depend on ReadArticle")
	}
}

func TestResolveRevocation(t *testing.T) {
	svc := NewService()

	// Initial set includes UpdateProject (which implies ReadProjectFull and ReadProjectBasic)
	// and CreateIssue (which implies ReadProjectBasic)
	active := svc.ResolveImplied([]PermKey{PermUpdateProject, PermCreateIssue})

	// Revoke ReadProjectBasic
	remaining := svc.ResolveRevocation(active, PermReadProjectBasic)

	// Since UpdateProject, ReadProjectFull, and CreateIssue all depend directly or transitively
	// on ReadProjectBasic, all of them must be dropped!
	for _, p := range remaining {
		if p == PermUpdateProject || p == PermReadProjectFull || p == PermReadProjectBasic || p == PermCreateIssue {
			t.Errorf("revocation failed: permission %s is still present in remaining list", p)
		}
	}

	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining permissions, got %v", remaining)
	}
}

func TestValidatePermissionsForScope(t *testing.T) {
	svc := NewService()

	mix := []PermKey{
		PermCreateUser,         // Global
		PermCreateOrganization, // Global
		PermUpdateOrganization, // Organization
		PermReadOrganization,   // Organization
		PermReadIssue,          // Project
		PermCreateIssue,        // Project
	}

	// 1. Global Scope: everything applies
	globalValid := svc.ValidatePermissionsForScope(mix, ScopeGlobal)
	if len(globalValid) != len(mix) {
		t.Errorf("expected all %d permissions valid at Global scope, got %d", len(mix), len(globalValid))
	}

	// 2. Organization Scope: Global permissions must not apply
	orgValid := svc.ValidatePermissionsForScope(mix, ScopeOrganization)
	for _, p := range orgValid {
		if p == PermCreateUser || p == PermCreateOrganization {
			t.Errorf("global permission %s should not apply at Organization scope", p)
		}
	}
	if len(orgValid) != 4 {
		t.Errorf("expected 4 permissions at Organization scope, got %d (%v)", len(orgValid), orgValid)
	}

	// 3. Project Scope: Global and Organization permissions must not apply
	projValid := svc.ValidatePermissionsForScope(mix, ScopeProject)
	for _, p := range projValid {
		if p != PermReadIssue && p != PermCreateIssue {
			t.Errorf("non-project permission %s should not apply at Project scope", p)
		}
	}
	if len(projValid) != 2 {
		t.Errorf("expected 2 permissions at Project scope, got %d (%v)", len(projValid), projValid)
	}
}

func TestInherentPermissions(t *testing.T) {
	svc := NewService()

	mockPerms := map[PermKey]bool{
		PermCreateIssue:          true,
		PermAddAttachment:        true,
		PermCreateIssueComment:   true,
		PermCreateArticleComment: true,
	}
	hasPerm := func(id PermKey) bool {
		return mockPerms[id]
	}

	// Reporter can view and update public fields of their own issues
	if !svc.CheckInherentAccess(InherentReadOwnIssuePublicFields, true, hasPerm) {
		t.Errorf("reporter should inherently have read access to own issue public fields")
	}
	if !svc.CheckInherentAccess(InherentUpdateOwnIssuePublicFields, true, hasPerm) {
		t.Errorf("reporter should inherently have update access to own issue public fields")
	}

	// Non-reporter cannot inherently access
	if svc.CheckInherentAccess(InherentUpdateOwnIssuePublicFields, false, hasPerm) {
		t.Errorf("non-reporter should NOT inherently have update access to issue public fields")
	}

	// Attachment uploader can delete own attachment
	if !svc.CheckInherentAccess(InherentDeleteOwnAttachment, true, hasPerm) {
		t.Errorf("uploader should inherently be able to delete their own attachment")
	}

	// Comment creator can read own issue comment
	if !svc.CheckInherentAccess(InherentReadOwnIssueComment, true, hasPerm) {
		t.Errorf("comment creator should inherently be able to read own comment")
	}

	// Article comment creator can read, update, and delete own article comments
	if !svc.CheckInherentAccess(InherentReadOwnArticleComment, true, hasPerm) {
		t.Errorf("article comment creator should inherently be able to read own comment")
	}
	if !svc.CheckInherentAccess(InherentUpdateOwnArticleComment, true, hasPerm) {
		t.Errorf("article comment creator should inherently be able to update own comment")
	}
}
