package rbac_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"youtrack_backend/internal/services/rbac"
)

func TestContext_ScopeMatchEvaluator(t *testing.T) {
	ctx := context.Background()
	scopeEval := rbac.ScopeMatchEvaluator()

	// 1. Global scope grants access to anything
	reqGlobal := rbac.AccessRequest{
		ScopeType: "GLOBAL",
		Resource:  "project:project_alpha",
	}
	allowed, _ := scopeEval(ctx, reqGlobal)
	if !allowed {
		t.Errorf("global scope should allow any resource")
	}

	// 2. Organization scope matching
	reqOrgMatch := rbac.AccessRequest{
		ScopeType: "ORGANIZATION:org_100",
		Context: map[string]any{
			"resource_scope": "ORGANIZATION:org_100:project_50",
		},
	}
	allowed, _ = scopeEval(ctx, reqOrgMatch)
	if !allowed {
		t.Errorf("organization scope should match organization resource")
	}

	reqOrgMismatch := rbac.AccessRequest{
		ScopeType: "ORGANIZATION:org_100",
		Context: map[string]any{
			"resource_scope": "ORGANIZATION:org_200:project_50",
		},
	}
	allowed, reason := scopeEval(ctx, reqOrgMismatch)
	if allowed {
		t.Errorf("organization scope mismatch should be denied")
	}
	if !strings.Contains(reason, "scope mismatch") {
		t.Errorf("expected reason to mention scope mismatch, got: %s", reason)
	}

	// 3. Project scope matching
	reqProjMatch := rbac.AccessRequest{
		ScopeType: "PROJECT:proj_alpha",
		Context: map[string]any{
			"resource_scope": "PROJECT:proj_alpha",
		},
	}
	allowed, _ = scopeEval(ctx, reqProjMatch)
	if !allowed {
		t.Errorf("matching project scope should be permitted")
	}

	reqProjMismatch := rbac.AccessRequest{
		ScopeType: "PROJECT:proj_alpha",
		Context: map[string]any{
			"resource_scope": "PROJECT:proj_beta",
		},
	}
	allowed, _ = scopeEval(ctx, reqProjMismatch)
	if allowed {
		t.Errorf("different project scope must be denied")
	}
}

func TestContext_InherentOwnershipEvaluator(t *testing.T) {
	ctx := context.Background()
	ownerEval := rbac.InherentOwnershipEvaluator()

	// 1. Matching author ID
	reqOwner := rbac.AccessRequest{
		SubjectID: "alice",
		Resource:  "issue:100",
		Context: map[string]any{
			"author_id": "alice",
		},
	}
	allowed, _ := ownerEval(ctx, reqOwner)
	if !allowed {
		t.Errorf("inherent owner should be permitted")
	}

	// 2. Matching boolean flag is_reporter
	reqReporter := rbac.AccessRequest{
		SubjectID: "bob",
		Resource:  "issue:101",
		Context: map[string]any{
			"is_reporter": true,
		},
	}
	allowed, _ = ownerEval(ctx, reqReporter)
	if !allowed {
		t.Errorf("reporter should be permitted via is_reporter flag")
	}

	// 3. Non-owner
	reqNonOwner := rbac.AccessRequest{
		SubjectID: "charlie",
		Resource:  "issue:100",
		Context: map[string]any{
			"author_id": "alice",
		},
	}
	allowed, reason := ownerEval(ctx, reqNonOwner)
	if allowed {
		t.Errorf("non-owner must be denied")
	}
	if !strings.Contains(reason, "does not own resource") {
		t.Errorf("expected reason to mention non-ownership, got: %s", reason)
	}
}

func TestContext_IPWhitelistEvaluator(t *testing.T) {
	ctx := context.Background()
	ipEval := rbac.IPWhitelistEvaluator([]string{
		"192.168.1.0/24",
		"10.0.0.0/8",
	})

	// 1. IP in allowed subnet
	reqAllowed := rbac.AccessRequest{
		Context: map[string]any{
			"client_ip": "192.168.1.55",
		},
	}
	allowed, _ := ipEval(ctx, reqAllowed)
	if !allowed {
		t.Errorf("IP in allowed subnet should pass")
	}

	// 2. IP outside allowed subnet
	reqProhibited := rbac.AccessRequest{
		Context: map[string]any{
			"client_ip": "203.0.113.195",
		},
	}
	allowed, reason := ipEval(ctx, reqProhibited)
	if allowed {
		t.Errorf("external untrusted IP must be denied")
	}
	if !strings.Contains(reason, "is not permitted by whitelist") {
		t.Errorf("expected reason to indicate whitelist rejection, got: %s", reason)
	}
}

func TestContext_CompositeEvaluators(t *testing.T) {
	ctx := context.Background()

	evalTrue := func(_ context.Context, _ rbac.AccessRequest) (bool, string) {
		return true, "condition A passed"
	}
	evalFalse := func(_ context.Context, _ rbac.AccessRequest) (bool, string) {
		return false, "condition B failed"
	}

	// AllOf
	allOfFail := rbac.AllOf(evalTrue, evalFalse)
	allowed, reason := allOfFail(ctx, rbac.AccessRequest{})
	if allowed {
		t.Errorf("AllOf with one false must fail")
	}
	if reason != "condition B failed" {
		t.Errorf("unexpected failure reason: %s", reason)
	}

	allOfPass := rbac.AllOf(evalTrue, evalTrue)
	allowed, _ = allOfPass(ctx, rbac.AccessRequest{})
	if !allowed {
		t.Errorf("AllOf with all true must pass")
	}

	// AnyOf
	anyOfPass := rbac.AnyOf(evalFalse, evalTrue)
	allowed, _ = anyOfPass(ctx, rbac.AccessRequest{})
	if !allowed {
		t.Errorf("AnyOf with one true must pass")
	}

	anyOfFail := rbac.AnyOf(evalFalse, evalFalse)
	allowed, _ = anyOfFail(ctx, rbac.AccessRequest{})
	if allowed {
		t.Errorf("AnyOf with all false must fail")
	}
}

func TestContext_HybridPDPIntegration(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph(nil)

	// Developer role has UPDATE_ISSUE
	_ = graph.AddRole(&rbac.Role{
		ID:          "DEVELOPER",
		Permissions: []string{"UPDATE_ISSUE"},
	})

	pdp := rbac.NewPDPEngine(graph, nil, nil)

	// Context Manager: require Scope matching AND internal IP
	cm := rbac.NewContextManager()
	cm.AddGlobalRule(rbac.ScopeMatchEvaluator())
	cm.AddPermissionRule("UPDATE_ISSUE", rbac.IPWhitelistEvaluator([]string{"10.0.0.0/16"}))

	pdp.SetContextManager(cm)

	// 1. User has DEVELOPER, matching project, AND allowed internal IP -> Allowed
	validReq := rbac.AccessRequest{
		SubjectID: "alice",
		// RoleIDs:    []string{"DEVELOPER"},
		Permission: "UPDATE_ISSUE",
		ScopeType:  "PROJECT:ALPHA",
		Context: map[string]any{
			"resource_scope": "PROJECT:ALPHA",
			"client_ip":      "10.0.5.22",
		},
	}
	dec := pdp.Evaluate(ctx, validReq)
	if !dec.Allowed {
		t.Errorf("expected allowed in valid hybrid request, got reason: %s", dec.Reason)
	}

	// 2. User has DEVELOPER and matching project, BUT untrusted IP -> Denied
	untrustedIPReq := validReq
	untrustedIPReq.Context = map[string]any{
		"resource_scope": "PROJECT:ALPHA",
		"client_ip":      "198.51.100.4",
	}
	dec = pdp.Evaluate(ctx, untrustedIPReq)
	if dec.Allowed {
		t.Errorf("expected denied due to untrusted IP")
	}
	if !strings.Contains(dec.Reason, "contextual evaluation failed") {
		t.Errorf("expected reason to mention contextual evaluation failed, got: %s", dec.Reason)
	}

	// 3. User has DEVELOPER and valid IP, BUT mismatched project scope -> Denied
	mismatchedScopeReq := validReq
	mismatchedScopeReq.Context = map[string]any{
		"resource_scope": "PROJECT:BETA",
		"client_ip":      "10.0.5.22",
	}
	dec = pdp.Evaluate(ctx, mismatchedScopeReq)
	if dec.Allowed {
		t.Errorf("expected denied due to scope mismatch")
	}
}

func TestContext_TimeWindowEvaluator(t *testing.T) {
	ctx := context.Background()

	// Current hour UTC
	currentHour := time.Now().UTC().Hour()

	// Window covering current hour
	winAllow := rbac.TimeWindowEvaluator(currentHour, currentHour+1, time.UTC)
	allowed, _ := winAllow(ctx, rbac.AccessRequest{})
	if !allowed {
		t.Errorf("request within current hour window should be allowed")
	}

	// Window outside current hour
	winDeny := rbac.TimeWindowEvaluator((currentHour+2)%24, (currentHour+3)%24, time.UTC)
	allowed, reason := winDeny(ctx, rbac.AccessRequest{})
	if allowed {
		t.Errorf("request outside window should be denied")
	}
	if !strings.Contains(reason, "outside permitted time window") {
		t.Errorf("expected reason to mention outside permitted time window, got: %s", reason)
	}
}
