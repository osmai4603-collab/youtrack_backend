package rbac

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// AttributeEvaluator is a function that assesses contextual attributes (subject, resource, environment)
// for a given access request. It returns (allowed bool, reason string).
type AttributeEvaluator func(ctx context.Context, req AccessRequest) (bool, string)

// ContextManager coordinates contextual, attribute-based policy rules (ABAC)
// on top of role-based authorization (Hybrid RBAC+ABAC per NIST SP 800-162).
type ContextManager struct {
	mu              sync.RWMutex
	globalRules     []AttributeEvaluator
	permissionRules map[string][]AttributeEvaluator
}

// NewContextManager instantiates an empty ContextManager.
func NewContextManager() *ContextManager {
	return &ContextManager{
		permissionRules: make(map[string][]AttributeEvaluator),
	}
}

// AddGlobalRule adds a rule that applies to all permissions evaluated by the PDP.
func (cm *ContextManager) AddGlobalRule(evaluator AttributeEvaluator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.globalRules = append(cm.globalRules, evaluator)
}

// AddPermissionRule binds an evaluator to a specific target permission key.
func (cm *ContextManager) AddPermissionRule(permission string, evaluator AttributeEvaluator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.permissionRules[permission] = append(cm.permissionRules[permission], evaluator)
}

// EvaluateContext executes all applicable global and permission-specific rules.
// If any rule returns false, access is denied (Fail-Closed).
func (cm *ContextManager) EvaluateContext(ctx context.Context, req AccessRequest) (bool, string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 1. Evaluate global rules
	for _, rule := range cm.globalRules {
		if allowed, reason := rule(ctx, req); !allowed {
			return false, reason
		}
	}

	// 2. Evaluate permission-specific rules
	if rules, exists := cm.permissionRules[req.Permission]; exists {
		for _, rule := range rules {
			if allowed, reason := rule(ctx, req); !allowed {
				return false, reason
			}
		}
	}

	return true, "contextual constraints satisfied"
}

// --- Pre-built Reusable Context Evaluators ---

// ScopeMatchEvaluator verifies that the request scope is compatible with the resource scope.
// Hierarchical scopes supported:
// - "GLOBAL": encompasses all resources across the entire organization.
// - "ORGANIZATION:<org_id>": encompasses all projects within that organization.
// - "PROJECT:<proj_id>": only permits access if the targeted resource belongs to the same project.
func ScopeMatchEvaluator() AttributeEvaluator {
	return func(ctx context.Context, req AccessRequest) (bool, string) {
		reqScope := strings.TrimSpace(req.ScopeType)
		if reqScope == "" || reqScope == "GLOBAL" {
			return true, "global scope applies to all resources"
		}

		resScope, _ := req.Context["resource_scope"].(string)
		resScope = strings.TrimSpace(resScope)
		if resScope == "" {
			// If resource scope is not specified in context, check req.Resource prefix
			resScope = req.Resource
		}

		if strings.HasPrefix(reqScope, "ORGANIZATION:") {
			orgID := strings.TrimPrefix(reqScope, "ORGANIZATION:")
			if strings.HasPrefix(resScope, "ORGANIZATION:"+orgID) || strings.Contains(resScope, orgID) {
				return true, "resource belongs to organization scope"
			}
			return false, fmt.Sprintf("scope mismatch: request scope %s does not match resource %s", reqScope, resScope)
		}

		if strings.HasPrefix(reqScope, "PROJECT:") {
			projID := strings.TrimPrefix(reqScope, "PROJECT:")
			if strings.HasPrefix(resScope, "PROJECT:"+projID) || strings.Contains(resScope, projID) {
				return true, "resource belongs to project scope"
			}
			return false, fmt.Sprintf("scope mismatch: request scope %s does not match resource %s", reqScope, resScope)
		}

		if reqScope == resScope {
			return true, "scope matched exactly"
		}

		return false, fmt.Sprintf("scope mismatch: required %s, resource has %s", reqScope, resScope)
	}
}

// InherentOwnershipEvaluator verifies that the requesting subject is the inherent author or reporter
// of the target entity (e.g. YouTrack inherent permissions for issue reporter/comment author).
func InherentOwnershipEvaluator() AttributeEvaluator {
	return func(ctx context.Context, req AccessRequest) (bool, string) {
		if req.Context == nil {
			return false, "inherent ownership denied: no context attributes provided"
		}

		// Check boolean flag if resolved upstream
		if isAuthor, ok := req.Context["is_author"].(bool); ok && isAuthor {
			return true, "subject is author/owner of the resource"
		}
		if isReporter, ok := req.Context["is_reporter"].(bool); ok && isReporter {
			return true, "subject is reporter/owner of the resource"
		}

		// Check explicit owner ID
		ownerID, _ := req.Context["owner_id"].(string)
		if ownerID != "" && ownerID == req.SubjectID {
			return true, "subject ID matches resource owner ID"
		}

		authorID, _ := req.Context["author_id"].(string)
		if authorID != "" && authorID == req.SubjectID {
			return true, "subject ID matches resource author ID"
		}

		return false, fmt.Sprintf("subject %s does not own resource %s", req.SubjectID, req.Resource)
	}
}

// TimeWindowEvaluator enforces that requests occur within permissible hours (e.g., 08:00 - 18:00 UTC).
func TimeWindowEvaluator(startHour, endHour int, loc *time.Location) AttributeEvaluator {
	if loc == nil {
		loc = time.UTC
	}
	return func(ctx context.Context, req AccessRequest) (bool, string) {
		now := time.Now().In(loc)
		hour := now.Hour()

		if hour >= startHour && hour < endHour {
			return true, fmt.Sprintf("request within permitted time window (%02d:00-%02d:00)", startHour, endHour)
		}
		return false, fmt.Sprintf("request outside permitted time window (current: %02d:00, permitted: %02d:00-%02d:00)",
			hour, startHour, endHour)
	}
}

// IPWhitelistEvaluator verifies that the client IP address in context originates from an authorized CIDR subnet.
func IPWhitelistEvaluator(allowedCIDRs []string) AttributeEvaluator {
	var ipNets []*net.IPNet
	for _, cidr := range allowedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			ipNets = append(ipNets, ipNet)
		}
	}

	return func(ctx context.Context, req AccessRequest) (bool, string) {
		if req.Context == nil {
			return false, "client IP check failed: no context provided"
		}

		ipStr, _ := req.Context["client_ip"].(string)
		ipStr = strings.TrimSpace(ipStr)
		if ipStr == "" {
			return false, "client IP check failed: 'client_ip' attribute missing in context"
		}

		clientIP := net.ParseIP(ipStr)
		if clientIP == nil {
			return false, fmt.Sprintf("client IP check failed: invalid IP format '%s'", ipStr)
		}

		for _, ipNet := range ipNets {
			if ipNet.Contains(clientIP) {
				return true, fmt.Sprintf("client IP %s is within allowed subnet %s", ipStr, ipNet.String())
			}
		}

		return false, fmt.Sprintf("client IP %s is not permitted by whitelist", ipStr)
	}
}

// AllOf creates a composite evaluator requiring all sub-evaluators to pass (Logical AND).
func AllOf(evaluators ...AttributeEvaluator) AttributeEvaluator {
	return func(ctx context.Context, req AccessRequest) (bool, string) {
		for _, eval := range evaluators {
			if allowed, reason := eval(ctx, req); !allowed {
				return false, reason
			}
		}
		return true, "all composite conditions satisfied"
	}
}

// AnyOf creates a composite evaluator passing if at least one sub-evaluator passes (Logical OR).
func AnyOf(evaluators ...AttributeEvaluator) AttributeEvaluator {
	return func(ctx context.Context, req AccessRequest) (bool, string) {
		var failures []string
		for _, eval := range evaluators {
			if allowed, reason := eval(ctx, req); allowed {
				return true, reason
			} else {
				failures = append(failures, reason)
			}
		}
		return false, fmt.Sprintf("none of conditions passed: %s", strings.Join(failures, "; "))
	}
}
