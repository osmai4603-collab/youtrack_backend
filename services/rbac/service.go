package rbac

import (
	"context"
	"fmt"
	"time"
	perms2 "youtrack_backend/internal/services/permissions_services"
)

// RoleRepository abstracts persistent storage operations for roles (e.g. PostgreSQL).
type RoleRepository interface {
	GetRole(ctx context.Context, id string) (*Role, error)
	GetRolesWhereIDs(ctx context.Context, ids []string) ([]*Role, error)
	CreateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, roleID string) error
}

// PermissionCatalog abstracts external permission catalog operations, such as expanding implied permissions.
type PermissionCatalog interface {
	ResolveImplied(permissionIDs []string) []string
}

// IRBACService is the unified facade interface providing comprehensive RBAC capabilities,
// integrating PAP, PDP, PEP, SoD, Session Management, JIT, and Symmetric Auditing.
type IRBACService interface {
	// --- Policy Administration Point (PAP) ---
	CreateRole(ctx context.Context, role *Role) (*Role, error)
	GetRole(ctx context.Context, id RoleID) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, id RoleID) error
	GetAllRoles(ctx context.Context) ([]*Role, error)

	// --- Separation of Duties (Constrained RBAC) ---
	AddSoDConstraint(ctx context.Context, constraint SoDConstraint) error
	RemoveSoDConstraint(ctx context.Context, id string) error
	GetSoDConstraint(ctx context.Context, id string) (SoDConstraint, error)
	ListSoDConstraints(ctx context.Context) ([]SoDConstraint, error)

	// --- User Assignments (Provisioning) ---
	AssignRoleToUser(ctx context.Context, ua UserAssignment) error
	RevokeRoleFromUser(ctx context.Context, userID string, roleID RoleID, scope string) error
	GetUserAssignments(ctx context.Context, userID string) ([]UserAssignment, error)

	// --- Sessions & JIT Activation ---
	CreateSession(ctx context.Context, userID string, initialRoles []RoleID, ttl time.Duration) (*Session, error)
	ActivateRoleInSession(ctx context.Context, sessionID string, role RoleID, ttl time.Duration) error
	DeactivateRoleInSession(ctx context.Context, sessionID string, role RoleID) error
	TerminateSession(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// --- Policy Decision Point (PDP) & Enforcement (PEP) ---
	Evaluate(ctx context.Context, req AccessRequest) Decision
	Authorize(ctx context.Context, req AccessRequest) error
	AuthorizePermission(ctx context.Context, subjectID string, permission string) error
	AuthorizeAny(ctx context.Context, subjectID string, permissions ...string) error
	AuthorizeAll(ctx context.Context, subjectID string, permissions ...string) error
	// HasAuthorize(ctx context.Context, permission string) error

	// --- Symmetric RBAC & Governance ---
	GetEffectivePermissions(ctx context.Context, roleID RoleID) ([]string, error)
	GetUserEffectivePermissions(ctx context.Context, userID string) ([]string, error)
	GetRolesWithPermission(ctx context.Context, permission string) ([]RoleID, error)
	GetUsersWithPermission(ctx context.Context, permission string) ([]string, error)
	FindOrphanedRoles(ctx context.Context) ([]RoleID, error)
	GenerateComplianceReport(ctx context.Context) (*ComplianceReport, error)

	// --- Context Manager Configuration ---
	GetContextManager() *ContextManager
}

// service implements IRBACService.
type service struct {
	graph      *RoleGraph
	sodMgr     *SoDManager
	pdp        *PDPEngine
	enforcer   *Enforcer
	contextMgr *ContextManager
	sessionMgr *SessionManager
	auditor    *SymmetricAuditor
	repo       RoleRepository
	catalog    PermissionCatalog
}

// ServiceOption configures optional components for IRBACService.
type ServiceOption func(*serviceConfig)

type serviceConfig struct {
	sessionRepo RBACSessionRepository
	roleRepo    RBACRoleRepository
}

// WithSessionRepository overrides the default in-memory session store.
func WithSessionRepository(repo RBACSessionRepository) ServiceOption {
	return func(c *serviceConfig) {
		c.sessionRepo = repo
	}
}

// WithRoleRegistryRepository overrides the default in-memory role registry store.
func WithRoleRegistryRepository(repo RBACRoleRepository) ServiceOption {
	return func(c *serviceConfig) {
		c.roleRepo = repo
	}
}

// NewService constructs a fully-wired, production-ready RBAC service facade.
func NewService(repo RoleRepository, catalog PermissionCatalog, permService perms2.IPermissionService, opts ...ServiceOption) IRBACService {
	var cfg serviceConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	graph := NewRoleGraph(cfg.roleRepo)
	sodMgr := NewSoDManager()

	var impliedResolver ImpliedPermissionResolver
	if catalog != nil {
		impliedResolver = catalog
	}

	pdp := NewPDPEngine(graph, impliedResolver, permService)
	contextMgr := NewContextManager()
	pdp.SetContextManager(contextMgr)

	enforcer := NewEnforcer(pdp, permService)

	sessionMgr := NewSessionManager(cfg.sessionRepo)
	sessionMgr.SetGraph(graph)
	sessionMgr.SetSoDManager(sodMgr)

	auditor := NewSymmetricAuditor(graph, sodMgr)

	return &service{
		graph:      graph,
		sodMgr:     sodMgr,
		pdp:        pdp,
		enforcer:   enforcer,
		contextMgr: contextMgr,
		sessionMgr: sessionMgr,
		auditor:    auditor,
		repo:       repo,
		catalog:    catalog,
	}
}

// --- Policy Administration Point (PAP) ---

func (s *service) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	if role == nil {
		return nil, fmt.Errorf("%w: role cannot be nil", ErrInvalidInput)
	}

	// 1. Add to the in-memory role registry.
	if err := s.graph.AddRole(role); err != nil {
		return nil, err
	}

	// 2. Persist to repository if configured
	if s.repo != nil {
		persisted, err := s.repo.CreateRole(ctx, role)
		if err != nil {
			// Rollback from graph
			_ = s.graph.DeleteRole(role.ID)
			return nil, err
		}
		return persisted, nil
	}

	return role, nil
}

func (s *service) GetRole(ctx context.Context, id RoleID) (*Role, error) {
	// First check memory graph
	role, err := s.graph.GetRole(id)
	if err == nil {
		return role, nil
	}

	// Fallback to repository
	if s.repo != nil {
		persisted, repoErr := s.repo.GetRole(ctx, id)
		if repoErr == nil && persisted != nil {
			_ = s.graph.AddRole(persisted)
			return persisted, nil
		}
	}

	return nil, err
}

func (s *service) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	if err := s.graph.UpdateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *service) DeleteRole(ctx context.Context, id RoleID) error {
	// 1. Delete from graph (checks immutable protection)
	if err := s.graph.DeleteRole(id); err != nil {
		return err
	}

	// 2. Delete from repository if configured
	if s.repo != nil {
		return s.repo.DeleteRole(ctx, id)
	}

	return nil
}

func (s *service) GetAllRoles(ctx context.Context) ([]*Role, error) {
	return s.graph.GetAllRoles(), nil
}

// --- Separation of Duties (Constrained RBAC) ---

func (s *service) AddSoDConstraint(ctx context.Context, constraint SoDConstraint) error {
	return s.sodMgr.AddConstraint(constraint)
}

func (s *service) RemoveSoDConstraint(ctx context.Context, id string) error {
	return s.sodMgr.RemoveConstraint(id)
}

func (s *service) GetSoDConstraint(ctx context.Context, id string) (SoDConstraint, error) {
	return s.sodMgr.GetConstraint(id)
}

func (s *service) ListSoDConstraints(ctx context.Context) ([]SoDConstraint, error) {
	return s.sodMgr.ListConstraints(), nil
}

// --- User Assignments (Provisioning) ---

func (s *service) AssignRoleToUser(ctx context.Context, ua UserAssignment) error {
	return s.auditor.AssignRole(ua)
}

func (s *service) RevokeRoleFromUser(ctx context.Context, userID string, roleID RoleID, scope string) error {
	return s.auditor.RevokeRole(userID, roleID, scope)
}

func (s *service) GetUserAssignments(ctx context.Context, userID string) ([]UserAssignment, error) {
	return s.auditor.GetUserAssignments(userID), nil
}

// --- Sessions & JIT Activation ---

func (s *service) CreateSession(ctx context.Context, userID string, initialRoles []RoleID, ttl time.Duration) (*Session, error) {
	return s.sessionMgr.CreateSession(userID, initialRoles, ttl)
}

func (s *service) ActivateRoleInSession(ctx context.Context, sessionID string, role RoleID, ttl time.Duration) error {
	return s.sessionMgr.ActivateRole(sessionID, role, ttl)
}

func (s *service) DeactivateRoleInSession(ctx context.Context, sessionID string, role RoleID) error {
	return s.sessionMgr.DeactivateRole(sessionID, role)
}

func (s *service) TerminateSession(ctx context.Context, sessionID string) error {
	return s.sessionMgr.TerminateSession(sessionID)
}

func (s *service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.sessionMgr.GetSession(sessionID)
}

// --- Policy Decision Point (PDP) & Enforcement (PEP) ---

func (s *service) Evaluate(ctx context.Context, req AccessRequest) Decision {
	return s.pdp.Evaluate(ctx, req)
}

func (s *service) Authorize(ctx context.Context, req AccessRequest) error {
	return s.enforcer.Authorize(ctx, req)
}

func (s *service) AuthorizePermission(ctx context.Context, subjectID string, permission string) error {
	return s.enforcer.AuthorizePermission(ctx, subjectID, permission)
}

func (s *service) AuthorizeAny(ctx context.Context, subjectID string, permissions ...string) error {
	return s.enforcer.AuthorizeAny(ctx, subjectID, permissions...)
}

func (s *service) AuthorizeAll(ctx context.Context, subjectID string, permissions ...string) error {
	return s.enforcer.AuthorizeAll(ctx, subjectID, permissions...)
}

// --- Symmetric RBAC & Governance ---

func (s *service) GetEffectivePermissions(ctx context.Context, roleID RoleID) ([]string, error) {
	return s.graph.GetEffectivePermissions(roleID)
}

func (s *service) GetUserEffectivePermissions(ctx context.Context, userID string) ([]string, error) {
	return s.auditor.GetUserEffectivePermissions(userID)
}

func (s *service) GetRolesWithPermission(ctx context.Context, permission string) ([]RoleID, error) {
	return s.auditor.GetRolesWithPermission(permission)
}

func (s *service) GetUsersWithPermission(ctx context.Context, permission string) ([]string, error) {
	return s.auditor.GetUsersWithPermission(permission)
}

func (s *service) FindOrphanedRoles(ctx context.Context) ([]RoleID, error) {
	return s.auditor.FindOrphanedRoles(), nil
}

func (s *service) GenerateComplianceReport(ctx context.Context) (*ComplianceReport, error) {
	sessions := make([]Session, 0)
	return s.auditor.GenerateComplianceReport(sessions), nil
}

func (s *service) GetContextManager() *ContextManager {
	return s.contextMgr
}
