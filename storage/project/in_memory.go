package project

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"

	"youtrack_backend/internal/domain"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// ---------------------------------------------------------------------------
// In-memory implementation (tests, memory driver)
// ---------------------------------------------------------------------------

// InMemoryRepository is a thread-safe project store used by unit tests and the
// in-memory development driver. Seed projects with Upsert and establish
// visibility with Grant/GrantAll, SetLeader, or AddMember before listing.
type InMemoryRepository struct {
	mu         sync.RWMutex
	projects   map[string]*domain.Project
	grants     map[string]map[string]struct{}
	orgGrants  map[string]map[string]struct{}
	allUsers   map[string]struct{}
	favorites  map[string]map[string]struct{}
	members    map[string]map[string][]string // projectID -> userID -> role names
	owners     map[string]string              // projectID -> leader userID
	directory  map[string]*domain.ProjectMember
	subsystems map[string]map[string]*domain.ProjectSubsystem // projectID -> subsystemID -> subsystem
	templates  map[string]*domain.ProjectTemplate
}

// NewInMemoryRepository creates an empty in-memory project store.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		projects:   make(map[string]*domain.Project),
		grants:     make(map[string]map[string]struct{}),
		orgGrants:  make(map[string]map[string]struct{}),
		allUsers:   make(map[string]struct{}),
		favorites:  make(map[string]map[string]struct{}),
		members:    make(map[string]map[string][]string),
		owners:     make(map[string]string),
		directory:  make(map[string]*domain.ProjectMember),
		subsystems: make(map[string]map[string]*domain.ProjectSubsystem),
		templates:  make(map[string]*domain.ProjectTemplate),
	}
}

// Upsert creates or replaces a project record (test seeding helper).
func (r *InMemoryRepository) Upsert(project *domain.Project) {
	if project == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *project
	r.projects[project.ID] = &stored
}

// Grant gives a user read access to a single project.
func (r *InMemoryRepository) Grant(userID, projectID string) {
	if userID == "" || projectID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	projectIDs, ok := r.grants[userID]
	if !ok {
		projectIDs = make(map[string]struct{})
		r.grants[userID] = projectIDs
	}
	projectIDs[projectID] = struct{}{}
}

// GrantOrganization gives a user read access to all projects within an organization.
func (r *InMemoryRepository) GrantOrganization(userID, orgID string) {
	if userID == "" || orgID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	orgIDs, ok := r.orgGrants[userID]
	if !ok {
		orgIDs = make(map[string]struct{})
		r.orgGrants[userID] = orgIDs
	}
	orgIDs[orgID] = struct{}{}
}

// GrantAll gives a user read access to every project, mirroring a globally
// scoped role assignment.
func (r *InMemoryRepository) GrantAll(userID string) {
	if userID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.allUsers[userID] = struct{}{}
}

// AddMember attaches a user to a project (test seeding helper).
func (r *InMemoryRepository) AddMember(projectID, userID string, roles ...string) {
	if projectID == "" || userID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	roster, ok := r.members[projectID]
	if !ok {
		roster = make(map[string][]string)
		r.members[projectID] = roster
	}
	roster[userID] = append([]string(nil), roles...)
}

// SetLeader marks a user as the project's owner, which also makes them a member
// (test seeding helper).
func (r *InMemoryRepository) SetLeader(projectID, userID string) {
	if projectID == "" || userID == "" {
		return
	}
	r.mu.Lock()
	r.owners[projectID] = userID
	roster, ok := r.members[projectID]
	if !ok {
		roster = make(map[string][]string)
		r.members[projectID] = roster
	}
	if _, present := roster[userID]; !present {
		roster[userID] = []string{}
	}
	r.mu.Unlock()
}

// PutUser registers a user's directory record so it can appear in a project's
// roster (test seeding helper).
func (r *InMemoryRepository) PutUser(user *domain.ProjectMember) {
	if user == nil || user.UserID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *user
	r.directory[user.UserID] = &copied
}

func (r *InMemoryRepository) GetProjectByUserID(ctx context.Context, userID string, includeArchived bool) ([]*domain.Project, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	r.mu.RLock()

	_, seesEverything := r.allUsers[userID]
	granted := r.grants[userID]

	visible := make([]*domain.Project, 0, len(r.projects))
	for _, project := range r.projects {
		if !includeArchived && project.Archived {
			continue
		}
		if !seesEverything {
			_, hasGrant := granted[project.ID]
			var hasOrgGrant bool
			if project.OrganizationID != "" {
				if orgs, ok := r.orgGrants[userID]; ok {
					_, hasOrgGrant = orgs[project.OrganizationID]
				}
			}
			_, isMember := r.members[project.ID][userID]
			isOwner := r.owners[project.ID] == userID
			if !hasGrant && !hasOrgGrant && !isOwner && !isMember && !project.IsPublic {
				continue
			}
		}
		copied := *project
		// `pinned` is a per-identity preference, so the stored column is
		// replaced by this identity's favorite flag rather than OR-ed with it:
		// merging would leave globally seeded projects impossible to un-star.
		copied.Pinned = r.isFavoriteLocked(userID, project.ID)
		visible = append(visible, &copied)
	}
	r.mu.RUnlock()

	sortProjects(visible)
	requestlogger.LogStorageResult(ctx, projectSelectOp+"(userID: "+userID+", includeArchived: "+strconv.FormatBool(includeArchived)+")", visible)
	return visible, nil
}

// ListAll returns all projects without filtering by user membership.
func (r *InMemoryRepository) ListAll(ctx context.Context, includeArchived bool) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*domain.Project, 0, len(r.projects))
	for _, project := range r.projects {
		if !includeArchived && project.Archived {
			continue
		}
		copied := *project
		all = append(all, &copied)
	}
	sortProjects(all)
	return all, nil
}

// ListPublic returns all projects flagged as is_public = true.
func (r *InMemoryRepository) ListPublic(ctx context.Context, includeArchived bool) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	public := make([]*domain.Project, 0)
	for _, project := range r.projects {
		if !includeArchived && project.Archived {
			continue
		}
		if project.IsPublic {
			copied := *project
			public = append(public, &copied)
		}
	}
	sortProjects(public)
	return public, nil
}

// isFavoriteLocked reports whether the identity starred the project. The caller
// must already hold the lock.
func (r *InMemoryRepository) isFavoriteLocked(userID, projectID string) bool {
	starred, ok := r.favorites[userID]
	if !ok {
		return false
	}
	_, ok = starred[projectID]
	return ok
}

// SetFavorite records or clears the identity's favorite state for a project.
func (r *InMemoryRepository) SetFavorite(ctx context.Context, userID, projectID string, favorite bool) error {
	if userID == "" {
		return ErrUserIDRequired
	}
	if projectID == "" {
		return ErrProjectIDRequired
	}

	r.mu.Lock()

	starred, ok := r.favorites[userID]
	if !ok {
		starred = make(map[string]struct{})
		r.favorites[userID] = starred
	}
	if favorite {
		starred[projectID] = struct{}{}
		r.mu.Unlock()
		requestlogger.LogStorageResult(ctx, projectFavoriteOp, map[string]any{"favorite": true})
		return nil
	}
	delete(starred, projectID)
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, projectFavoriteOp, map[string]any{"favorite": false})
	return nil
}

// GetProjectByID returns the single project the identity may read, or
// ErrProjectNotVisible when it is absent or out of scope. Archived state is
// deliberately ignored, matching IsVisibleToUser.
func (r *InMemoryRepository) GetProjectByID(ctx context.Context, projectID string) (*domain.Project, error) {

	r.mu.RLock()
	copied := *r.projects[projectID]
	// `pinned` is a per-identity preference, so the stored column is replaced by
	// this identity's favorite flag exactly as the listing does.
	copied.Pinned = r.isFavoriteLocked("", projectID)
	r.mu.RUnlock()
	requestlogger.LogStorageResult(ctx, projectGetOp, &copied)
	return &copied, nil
}

// ListMembers returns the project's roster, ordered owner-first then by name so
// the result matches the ordering the SQL guarantees.
func (r *InMemoryRepository) GetListTeamMembersForProject(ctx context.Context, projectID string) ([]*domain.TeamMember, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}

	r.mu.RLock()
	// leaderID := r.owners[projectID]
	roster := r.members[projectID]
	members := make([]*domain.TeamMember, 0, len(roster))
	seen := make(map[string]struct{}, len(roster))
	for memberUserID, roles := range roster {
		if _, ok := seen[memberUserID]; ok {
			continue
		}
		seen[memberUserID] = struct{}{}

		member := &domain.TeamMember{
			ID:        memberUserID,
			UserID:    memberUserID,
			ProjectID: projectID,
			Roles:     append([]string(nil), roles...),
			//IsOwner:   memberUserID == leaderID,
		}
		if member.Roles == nil {
			member.Roles = []string{}
		}
		// if known, ok := r.directory[memberUserID]; ok {
		// 	member.Login = known.Login
		// 	member.Email = known.Email
		// 	member.FullName = known.FullName
		// 	member.Name = known.Name
		// 	member.AvatarURL = known.AvatarURL
		// 	member.Banned = known.Banned
		// 	member.CreatedAt = known.CreatedAt
		// }
		members = append(members, member)
	}
	r.mu.RUnlock()

	// sort.SliceStable(members, func(i, j int) bool {
	// 	if members[i].IsOwner != members[j].IsOwner {
	// 		return members[i].IsOwner
	// 	}
	// 	if members[i].FullName != members[j].FullName {
	// 		return members[i].FullName < members[j].FullName
	// 	}
	// 	return members[i].Login < members[j].Login
	// })
	logProjectMembers(ctx, projectMembersOp, members, slog.String("projectID", projectID))
	return members, nil
}

// Create stores a new project, stamping the timestamps the schema expects.
//
// A nil or unnamed project is rejected here rather than left to the database so
// the in-memory and Postgres stores reject the same inputs.
func (r *InMemoryRepository) Create(ctx context.Context, project *domain.Project) error {
	if project == nil {
		return ErrProjectNameRequired
	}
	if project.Name == "" {
		return ErrProjectNameRequired
	}
	if project.ID == "" {
		return ErrProjectIDRequired
	}

	r.mu.Lock()
	if _, exists := r.projects[project.ID]; exists {
		r.mu.Unlock()
		return fmt.Errorf("create project: %w", domain.ErrConflict)
	}

	now := time.Now().UTC()
	stored := *project
	if stored.CreatedAt == nil {
		stored.CreatedAt = &now
	}
	stored.UpdatedAt = &now
	if stored.StartingNumber == 0 {
		stored.StartingNumber = 1
	}
	r.projects[stored.ID] = &stored
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "projects.Create", &stored)
	return nil
}

// Update overwrites the mutable fields of an existing project.
//
// The write is a full replacement of the stored record, matching the Postgres
// implementation: callers read, merge, then write. A missing project is
// reported as not visible so an update cannot be used to probe for ids.
func (r *InMemoryRepository) Update(ctx context.Context, project *domain.Project) error {
	if project == nil {
		return ErrProjectNameRequired
	}
	if project.ID == "" {
		return ErrProjectIDRequired
	}
	if project.Name == "" {
		return ErrProjectNameRequired
	}

	r.mu.Lock()
	existing, ok := r.projects[project.ID]
	if !ok {
		r.mu.Unlock()
		return ErrProjectNotVisible
	}

	merged := *project
	// The favorite flag is per identity and lives in project_favorites.
	merged.CreationTime = existing.CreationTime
	merged.Pinned = existing.Pinned
	merged.CreatedBy = existing.CreatedBy
	merged.CreatedAt = existing.CreatedAt
	now := time.Now().UTC()
	merged.UpdatedAt = &now
	if merged.StartingNumber == 0 {
		merged.StartingNumber = 1
	}
	r.projects[merged.ID] = &merged
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "projects.Update", &merged)
	return nil
}

// Delete removes a project together with its subsystems and favorites.
func (r *InMemoryRepository) Delete(ctx context.Context, projectID string) error {
	if projectID == "" {
		return ErrProjectIDRequired
	}

	r.mu.Lock()
	if _, ok := r.projects[projectID]; !ok {
		r.mu.Unlock()
		return ErrProjectNotVisible
	}
	delete(r.projects, projectID)
	delete(r.subsystems, projectID)
	for userID, starred := range r.favorites {
		delete(starred, projectID)
		if len(starred) == 0 {
			delete(r.favorites, userID)
		}
	}
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "projects.Delete", map[string]any{"deleted": true})
	return nil
}

// ListTemplates returns the project templates ordered by name.
func (r *InMemoryRepository) ListTemplates(ctx context.Context) ([]*domain.ProjectTemplate, error) {
	r.mu.RLock()

	templates := make([]*domain.ProjectTemplate, 0, len(r.templates))
	for _, template := range r.templates {
		copied := *template
		templates = append(templates, &copied)
	}
	r.mu.RUnlock()
	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })
	requestlogger.LogStorageResult(ctx, projectTemplateOp, templates)
	return templates, nil
}

// PutTemplate stores a project template (test seeding helper).
func (r *InMemoryRepository) PutTemplate(template *domain.ProjectTemplate) {
	if template == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *template
	r.templates[template.ID] = &stored
}

// ListSubsystems returns a project's subsystems ordered by name.
func (r *InMemoryRepository) ListSubsystems(ctx context.Context, projectID string) ([]*domain.ProjectSubsystem, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}

	r.mu.RLock()

	stored := r.subsystems[projectID]
	subsystems := make([]*domain.ProjectSubsystem, 0, len(stored))
	for _, subsystem := range stored {
		copied := *subsystem
		subsystems = append(subsystems, &copied)
	}
	r.mu.RUnlock()
	sort.Slice(subsystems, func(i, j int) bool { return subsystems[i].Name < subsystems[j].Name })
	requestlogger.LogStorageResult(ctx, "projects.ListSubsystems", subsystems)
	return subsystems, nil
}

// GetSubsystem returns one subsystem of a project.
func (r *InMemoryRepository) GetSubsystem(ctx context.Context, projectID, subsystemID string) (*domain.ProjectSubsystem, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}
	if subsystemID == "" {
		return nil, ErrSubsystemNotFound
	}

	r.mu.RLock()
	subsystem, ok := r.subsystems[projectID][subsystemID]
	if !ok {
		r.mu.RUnlock()
		return nil, ErrSubsystemNotFound
	}
	copied := *subsystem
	r.mu.RUnlock()
	requestlogger.LogStorageResult(ctx, "projects.GetSubsystem", &copied)
	return &copied, nil
}

// CreateSubsystem adds a subsystem, rejecting a duplicate name within the same
// project so the store mirrors the unique index in the schema.
func (r *InMemoryRepository) CreateSubsystem(ctx context.Context, subsystem *domain.ProjectSubsystem) error {
	if subsystem == nil || subsystem.Name == "" {
		return ErrSubsystemNameRequired
	}
	if subsystem.ProjectID == "" {
		return ErrProjectIDRequired
	}
	if subsystem.ID == "" {
		return ErrSubsystemIDRequired
	}

	r.mu.Lock()
	for _, existing := range r.subsystems[subsystem.ProjectID] {
		if existing.Name == subsystem.Name {
			r.mu.Unlock()
			return fmt.Errorf("create subsystem: %w", domain.ErrConflict)
		}
	}

	now := time.Now().UnixMilli()
	stored := *subsystem
	if stored.CreatedAt == 0 {
		stored.CreatedAt = now
	}
	stored.UpdatedAt = now
	if r.subsystems[stored.ProjectID] == nil {
		r.subsystems[stored.ProjectID] = make(map[string]*domain.ProjectSubsystem)
	}
	r.subsystems[stored.ProjectID][stored.ID] = &stored
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "projects.CreateSubsystem", &stored)
	return nil
}
