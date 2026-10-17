package perms2

import (
	"errors"
	"slices"
)

// IPermissionService defines the interface for managing and evaluating permissions.
type IPermissionService interface {
	GetPermission(id PermKey) *Permission
	GetAllPermissions() []Permission
	ContainsPermission(permission string) error
	GetPermissionsByScope(scope ScopeLevel) []Permission
	GetPermissionsByEntity(entity EntityType) []Permission
	ResolveImplied(permissionIDs []PermKey) []PermKey
	ResolveDependent(permissionIDs []PermKey) []PermKey
	GetImpliedPermissions(id PermKey) []PermKey
	GetDependentPermissions(id PermKey) []PermKey
	DependsOn(perm PermKey, target PermKey) bool
	ResolveRevocation(activePermissionIDs []PermKey, permissionToRemove PermKey) []PermKey
	ValidatePermissionsForScope(permissionIDs []PermKey, targetScope ScopeLevel) []PermKey
	HasPermission(grantedPermissions []PermKey, requiredPermission PermKey) bool
	CheckInherentAccess(action InherentAction, isAuthorOrReporter bool, hasPermission func(PermKey) bool) bool
}

// Service provides YouTrack permission management, validation, and evaluation.
type service struct {
	list []Permission
}

// NewService instantiates a new permissions service populated with the YouTrack catalog.
func NewService() IPermissionService {
	permisssions := GetDefaultPermissions()

	return &service{
		list: permisssions,
	}
}

func (s *service) GetPermission(id PermKey) *Permission {
	for _, perm := range s.list {
		if perm.ID == id {
			return &perm
		}
	}
	return nil

}

func (s *service) GetAllPermissions() []Permission {
	out := make([]Permission, len(s.list))
	copy(out, s.list)
	return out
}

// GetPermissionsByScope retrieves all permissions belonging to a specific scope level.
func (s *service) GetPermissionsByScope(scope ScopeLevel) []Permission {

	var res []Permission
	for _, p := range s.list {
		if p.Scope == scope {
			res = append(res, p)
		}
	}
	return res
}

// GetPermissionsByEntity retrieves permissions targeting a specific entity type.
func (s *service) GetPermissionsByEntity(entity EntityType) []Permission {

	var res []Permission
	for _, p := range s.list {
		if p.Entity == entity {
			res = append(res, p)
		}
	}
	return res
}

func (s *service) ContainsPermission(permission string) error {
	if slices.Contains(permissionsNames, permission) {
		return nil
	}
	return errors.New("this permission: " + permission + "does not found in all permissions.")
}

// ResolveImplied computes the transitive closure of implied permissions.
// As defined by YouTrack: when you add a permission with implied permissions to a role,
// the implied permissions are added automatically.
func (s *service) ResolveImplied(permissionIDs []PermKey) []PermKey {

	resultMap := make(map[PermKey]struct{})
	var queue []PermKey

	for _, id := range permissionIDs {
		perm := s.GetPermission(id)
		if perm != nil {
			if _, seen := resultMap[perm.ID]; !seen {
				resultMap[perm.ID] = struct{}{}
				queue = append(queue, perm.ID)
			}
		}
	}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		p := s.GetPermission(currentID)
		if p == nil {
			continue
		}
		for _, impliedID := range p.ImpliedPerms {
			if _, seen := resultMap[impliedID]; !seen {
				resultMap[impliedID] = struct{}{}
				queue = append(queue, impliedID)
			}
		}
	}

	out := make([]PermKey, 0, len(resultMap))
	for id := range resultMap {
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}

// ResolveDependent computes the transitive closure of dependent permissions.
// As defined by YouTrack: when evaluating permissions that rely on a set of target permissions,
// any permission whose operation technically depends on them is included transitively.
func (s *service) ResolveDependent(permissionIDs []PermKey) []PermKey {

	resultMap := make(map[PermKey]struct{})
	var queue []PermKey

	for _, id := range permissionIDs {
		perm := s.GetPermission(id)
		if perm != nil {
			if _, seen := resultMap[perm.ID]; !seen {
				resultMap[perm.ID] = struct{}{}
				queue = append(queue, perm.ID)
			}
		}
	}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		p := s.GetPermission(currentID)
		if p == nil {
			continue
		}
		for _, depID := range p.DependentPerms {
			if _, seen := resultMap[depID]; !seen {
				resultMap[depID] = struct{}{}
				queue = append(queue, depID)
			}
		}
	}

	out := make([]PermKey, 0, len(resultMap))
	for id := range resultMap {
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}

// GetDependentPermissions returns the direct dependent permissions for a given permission.
func (s *service) GetDependentPermissions(id PermKey) []PermKey {
	p := s.GetPermission(id)
	if p == nil {
		return nil
	}
	out := make([]PermKey, len(p.DependentPerms))
	copy(out, p.DependentPerms)
	slices.Sort(out)
	return out
}

// GetImpliedPermissions returns the direct implied permissions for a given permission.
func (s *service) GetImpliedPermissions(id PermKey) []PermKey {
	p := s.GetPermission(id)
	if p == nil {
		return nil
	}
	out := make([]PermKey, len(p.ImpliedPerms))
	copy(out, p.ImpliedPerms)
	slices.Sort(out)
	return out
}

// DependsOn checks whether perm depends directly or transitively on target.
func (s *service) DependsOn(perm PermKey, target PermKey) bool {
	if perm == target {
		return false
	}
	dependents := s.ResolveDependent([]PermKey{target})
	return slices.Contains(dependents, perm)
}

// ResolveRevocation determines which permissions remain after revoking a target permission.
// As defined by YouTrack: when you remove a permission with dependent permissions from a role,
// the dependent permissions are removed automatically.
func (s *service) ResolveRevocation(activePermissionIDs []PermKey, permissionToRemove PermKey) []PermKey {

	activeSet := make(map[PermKey]struct{}, len(activePermissionIDs))
	for _, id := range activePermissionIDs {
		activeSet[id] = struct{}{}
	}

	// Queue permissions to delete (starting with the target)
	toDelete := make(map[PermKey]struct{})
	queue := []PermKey{permissionToRemove}
	toDelete[permissionToRemove] = struct{}{}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		p := s.GetPermission(curr)
		if p != nil {
			for _, depID := range p.DependentPerms {
				if _, alreadyMarked := toDelete[depID]; !alreadyMarked {
					toDelete[depID] = struct{}{}
					queue = append(queue, depID)
				}
			}
		}

	}

	var remaining []PermKey
	for id := range activeSet {
		if _, deleted := toDelete[id]; !deleted {
			remaining = append(remaining, id)
		}
	}
	slices.Sort(remaining)
	return remaining
}

// ValidatePermissionsForScope filters permissions based on YouTrack scope isolation rules:
// - Global assignment: all permissions apply.
// - Organization assignment: global permissions no longer propagate and have no effect.
// - Project assignment: global and organization-level permissions have no effect.
func (s *service) ValidatePermissionsForScope(permissionIDs []PermKey, targetScope ScopeLevel) []PermKey {

	var valid []PermKey
	for _, id := range permissionIDs {
		perm := s.GetPermission(id)
		if perm == nil {
			continue
		}

		switch targetScope {
		case ScopeGlobal:
			// Global assignments allow all scopes to apply
			valid = append(valid, id)
		case ScopeOrganization:
			// Only organization and project permissions take effect
			if perm.Scope == ScopeOrganization || perm.Scope == ScopeProject {
				valid = append(valid, id)
			}
		case ScopeProject:
			// Only project-level permissions take effect
			if perm.Scope == ScopeProject {
				valid = append(valid, id)
			}
		}
	}
	return valid
}

func (s *service) HasPermission(grantedPermissions []PermKey, requiredPermission PermKey) bool {
	return slices.Contains(grantedPermissions, requiredPermission)
}

// CheckInherentAccess evaluates inherent rights granted to creators/reporters without explicit permissions:
// - Issue reporters inherit permission to view public fields, update public fields, and add links to their issues.
// - File attachers inherit modifying their files and restricting visibility without Update Attachment.
// - All users can delete files they attached themselves without Delete Attachment.
// - Issue comment creators inherit reading their own comments.
// - Issue work item creators inherit reading their own work items.
// - Article comment creators inherit reading and updating their own comments.
// - Users cannot delete their own issues without Delete Issue.
func (s *service) CheckInherentAccess(action InherentAction, isAuthorOrReporter bool, hasPermission func(PermKey) bool) bool {
	if !isAuthorOrReporter {
		// Not the author/reporter; must have explicit permission
		return false
	}

	switch action {
	case InherentReadOwnIssuePublicFields, InherentUpdateOwnIssuePublicFields, InherentLinkOwnIssue:
		// Requires CREATE_ISSUE permission in the project
		return hasPermission(PermCreateIssue)

	case InherentModifyOwnAttachment, InherentRestrictOwnAttachment:
		// Requires CREATE_ATTACHMENT_ISSUE permission
		return hasPermission(PermAddAttachment)

	case InherentDeleteOwnAttachment:
		// Any user can delete files they attached themselves
		return true

	case InherentReadOwnIssueComment:
		// Requires CREATE_COMMENT permission
		return hasPermission(PermCreateIssueComment)

	case InherentReadOwnWorkItem:
		// Requires CREATE_WORK_ITEM permission
		return hasPermission(PermCreateWorkItem)

	case InherentReadOwnArticleComment, InherentUpdateOwnArticleComment, InherentDeleteOwnArticleComment:
		// Requires CREATE_ARTICLE_COMMENT permission
		return hasPermission(PermCreateArticleComment)

	default:
		return false
	}
}
