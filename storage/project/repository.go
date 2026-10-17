// Package project provides the persistence layer for the project records a
// user is allowed to read. Visibility is not a column on `projects`: it is
// derived from the role assignments the identity holds, so the repository
// resolves it while querying instead of trusting a caller-supplied filter.
package project

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sort"
	"strconv"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	perms2 "youtrack_backend/internal/services/permissions_services"

	"github.com/jackc/pgx/v5"
)

// The sentinels are declared once in the domain package and aliased here, so
// that errors.Is keeps matching across the storage/use-case boundary instead of
// comparing two distinct values that happen to share a name.
var (
	// ErrUserIDRequired is returned when an empty user ID is used.
	ErrUserIDRequired = domain.ErrUserIDRequired
	// ErrProjectIDRequired is returned when an empty project ID is used.
	ErrProjectIDRequired = domain.ErrProjectIDRequired
	// ErrProjectNotVisible is returned when the project is absent or outside
	// the caller's scope; the two cases are intentionally indistinguishable.
	ErrProjectNotVisible = domain.ErrProjectNotVisible
	// ErrProjectNameRequired is returned when a project is created unnamed.
	ErrProjectNameRequired = domain.ErrProjectNameRequired
	// ErrSubsystemNameRequired is returned when a subsystem is created unnamed.
	ErrSubsystemNameRequired = domain.ErrSubsystemNameRequired
	// ErrSubsystemNotFound is returned when a project has no such subsystem.
	// The Postgres implementation converts pgx's no-rows error into this
	// sentinel so both stores fail identically.
	ErrSubsystemNotFound = domain.ErrSubsystemNotFound
	// ErrStartingNumberInvalid is returned for a non-positive starting number.
	ErrStartingNumberInvalid = domain.ErrStartingNumberInvalid
	// ErrSubsystemIDRequired is returned when a subsystem is created without an
	// ID.
	ErrSubsystemIDRequired = domain.ErrSubsystemIDRequired
	// ErrConflict is returned when a write would duplicate an existing record.
	ErrConflict = domain.ErrConflict
)

// Repository defines the read contract backing the project endpoints, plus the
// per-identity favorite flag.
type Repository interface {
	// ListVisibleToUser returns the projects the identity may read, optionally
	// including archived ones. Archived projects are excluded by default. Each
	// returned project carries `Pinned` set to the identity's own favorite
	// state, not to the global `projects.pinned` column.
	GetProjectByUserID(ctx context.Context, userID string, includeArchived bool) ([]*domain.Project, error)
	// ListAll returns all projects without filtering by user membership.
	ListAll(ctx context.Context, includeArchived bool) ([]*domain.Project, error)
	// ListPublic returns all public projects (is_public = TRUE).
	ListPublic(ctx context.Context, includeArchived bool) ([]*domain.Project, error)
	// GetVisibleToUser returns the single project the identity may read, or
	// ErrProjectNotVisible when it does not exist or is out of scope. Archived
	// state is ignored, mirroring IsVisibleToUser, so a project hidden from the
	// default listing is still addressable by id.
	GetProjectByID(ctx context.Context, projectID string) (*domain.Project, error)
	// ListMembers returns the project's roster: its leader, its team members,
	// the members of its project-team group, and every identity holding a role
	// scoped to the project. It returns ErrProjectNotVisible when the project is
	// absent or outside the caller's scope.
	GetListTeamMembersForProject(ctx context.Context, projectID string) ([]*domain.TeamMember, error)
	// SetFavorite records the identity's favorite state for a project. Starring
	// an already-starred project is a no-op, and clearing an absent favorite is
	// a no-op, so the call is idempotent.
	SetFavorite(ctx context.Context, userID, projectID string, favorite bool) error

	// Create inserts a new project. The caller supplies the ID; the store
	// stamps created_at/updated_at. The global `pinned` column is deliberately
	// not written: favorites are per identity and live in project_favorites.
	Create(ctx context.Context, project *domain.Project) error
	// Update overwrites the mutable fields of an existing project.
	//
	// This is a full write, not a partial patch: callers are expected to read the
	// project, apply the change, and write the merged result. Letting the store
	// merge instead would make a deliberate "" or false impossible to express,
	// since it could not be told apart from an omitted field. Archived and
	// StartingNumber are included, so archiving and renumbering are ordinary
	// updates rather than separate statements.
	Update(ctx context.Context, project *domain.Project) error
	// Delete removes a project. Members, favorites and subsystems scoped to it
	// go with it through the schema's ON DELETE CASCADE. Unauthenticated by
	// design — see the warning above the write block.
	Delete(ctx context.Context, projectID string) error
	// ListTemplates returns the project templates offered at creation time.
	ListTemplates(ctx context.Context) ([]*domain.ProjectTemplate, error)
	// ListSubsystems returns a project's subsystems, ordered by name.
	ListSubsystems(ctx context.Context, projectID string) ([]*domain.ProjectSubsystem, error)
	// GetSubsystem returns one subsystem of a project, or ErrSubsystemNotFound.
	GetSubsystem(ctx context.Context, projectID, subsystemID string) (*domain.ProjectSubsystem, error)
	// CreateSubsystem adds a subsystem to a project. A name already used within
	// that project is rejected as a conflict rather than duplicated.
	CreateSubsystem(ctx context.Context, subsystem *domain.ProjectSubsystem) error
}

// NewRepository returns a Postgres-backed repository, or an in-memory one when
// no database handle is available.
func NewRepository(db postgres.DBTX) Repository {
	return &PostgresRepository{
		db:      db,
		permSvc: perms2.NewService(),
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL implementation
// ---------------------------------------------------------------------------

// PostgresRepository implements Repository against the `projects` table.
type PostgresRepository struct {
	db      postgres.DBTX
	permSvc perms2.IPermissionService
}

// NewPostgresRepository creates a new PostgresRepository instance.
func NewPostgresRepository(db postgres.DBTX) *PostgresRepository {
	return &PostgresRepository{
		db:      db,
		permSvc: perms2.NewService(),
	}
}

// ---------------------------------------------------------------------------
// Operation names
// ---------------------------------------------------------------------------

const projectSelectOp = "GetListPorjectsForUser"

const projectFavoriteOp = "projects.SetFavorite"

const projectGetOp = "GetProjectByID"

const projectMembersOp = "GetListMembersForProject"

// Operation names for the write paths and for the two auxiliary resources. They
// only label translated database errors, so they are grouped here with the
// other project operations rather than next to each query.
const (
	projectWriteOp     = "projects.Write"
	projectTemplateOp  = "projects.ListTemplates"
	projectSubsystemOp = "projects.Subsystems"
)

// ---------------------------------------------------------------------------
// Shared scan helpers
// ---------------------------------------------------------------------------

func scanProject(row interface{ Scan(...any) error }) (*domain.Project, error) {
	var p domain.Project
	if err := row.Scan(
		&p.ID,
		&p.Name,
		&p.ShortName,
		&p.Archived,
		&p.Template,
		&p.Description,
		&p.OrganizationID,
		&p.CreatedBy,
		&p.LeaderID,
		&p.FromEmail,
		&p.ReplyToEmail,
		&p.IconURL,
		&p.StartingNumber,
		&p.TeamID,
		&p.IsPublic,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

// scanSubsystem reads the seven columns the subsystem queries project, in the
// same order.
func scanSubsystem(row interface{ Scan(...any) error }) (*domain.ProjectSubsystem, error) {
	var s domain.ProjectSubsystem
	if err := row.Scan(
		&s.ID,
		&s.ProjectID,
		&s.Name,
		&s.OwnerID,
		&s.Color,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// scanProjectMember reads the ten columns listProjectMembersQuery projects, in
// the same order. created_at is nullable, so it lands in a pointer.
// func scanProjectMember(row interface{ Scan(...any) error }) (*domain.TeamMember, error) {
// 	var (
// 		m     domain.ProjectMember
// 		roles []string
// 	)
// 	if err := row.Scan(
// 		&m.UserID,
// 		&m.Login,
// 		&m.Email,
// 		&m.FullName,
// 		&m.Name,
// 		&m.AvatarURL,
// 		&m.Banned,
// 		&m.CreatedAt,
// 		&m.IsOwner,
// 		&roles,
// 	); err != nil {
// 		return nil, err
// 	}
// 	// An empty array_agg arrives as a nil slice, which would serialize as null
// 	// instead of an empty list.
// 	if roles == nil {
// 		roles = []string{}
// 	}
// 	m.Roles = roles
// 	return &m, nil
// }

// grantingPermissionRefs flattens the permission set that transitively grants
// target into a flat list of ids and names, so the SQL matches role rows
// whichever form the row happens to store.
func grantingPermissionRefs(permSvc perms2.IPermissionService, target perms2.PermKey) []string {
	if permSvc == nil {
		permSvc = perms2.NewService()
	}
	seen := make(map[string]struct{})
	var refs []string
	addRef := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			refs = append(refs, s)
		}
	}

	addRef(string(target))
	targetPerm := permSvc.GetPermission(target)
	if targetPerm != nil {
		addRef(targetPerm.DisplayName)
	}

	for _, p := range permSvc.GetAllPermissions() {
		implied := permSvc.ResolveImplied([]perms2.PermKey{p.ID})
		if slices.Contains(implied, target) {
			addRef(string(p.ID))
			addRef(p.DisplayName)
		}
	}

	return refs
}

// sortProjects applies the same ordering the SQL query guarantees: pinned
// projects first, then alphabetical by name.
func sortProjects(projects []*domain.Project) {
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].Pinned != projects[j].Pinned {
			return projects[i].Pinned
		}
		return projects[i].Name < projects[j].Name
	})
}

// sortMembers mirrors the SQL ordering: the owner first, then by full name, then
// by login so the order is total.
func sortMembers(members []*domain.ProjectMember) {
	sort.SliceStable(members, func(i, j int) bool {
		if members[i].IsOwner != members[j].IsOwner {
			return members[i].IsOwner
		}
		if members[i].FullName != members[j].FullName {
			return members[i].FullName < members[j].FullName
		}
		return members[i].Login < members[j].Login
	})
}

// Ensure pgx.ErrNoRows is used for error matching.
var _ = pgx.ErrNoRows
var _ = errors.Is

func logProjects(ctx context.Context, op string, projects []*domain.Project, args ...slog.Attr) {
	if len(projects) == 0 {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
			slog.String("result", "[]"),
		)
		return
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
		args...,
	)
	for index, project := range projects {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "  "+strconv.Itoa((index+1)),
			slog.String("id", project.ID),
			slog.String("teamID", project.TeamID),
			slog.String("name", project.Name),
		)
	}
}

func logProjectMembers(ctx context.Context, op string, members []*domain.TeamMember, args ...slog.Attr) {
	length := len(members)
	if length == 0 {
		requestlogger.LogStorageResult(ctx, projectMembersOp, "[]")
		return
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, op, args...)
	for index, mbr := range members {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "    "+strconv.Itoa((index+1)),
			slog.String("id", mbr.ID),
			slog.String("memberTP", mbr.MemberType),
			slog.String("memberGroup", mbr.MemberGroupID),
		)
	}

}
