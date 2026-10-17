package domain

import (
	"errors"
	"time"
)

// Project-scoped sentinels live here rather than in either the storage or the
// use-case package: both layers need to agree on them, and a sentinel declared
// in only one of them makes errors.Is fail across the layer boundary — which
// turns a mapped 400 or 404 into an unmapped 500.
var (
	// ErrUserIDRequired is returned when an empty user ID is used.
	ErrUserIDRequired = errors.New("user ID is required")
	// ErrProjectIDRequired is returned when an empty project ID is used.
	ErrProjectIDRequired = errors.New("project ID is required")
	// ErrProjectNotVisible is returned when the identity may not read the
	// project. A project that does not exist and one that is out of scope are
	// deliberately reported with the same sentinel, so the error cannot be used
	// to probe which project IDs are real.
	ErrProjectNotVisible = errors.New("project is not visible to the user")
	// ErrProjectNameRequired is returned when a project is created without a
	// name. The schema requires it too, but checking here turns a 400 into a
	// clear message instead of a translated NOT NULL violation.
	ErrProjectNameRequired = errors.New("project name is required")
	// ErrSubsystemNameRequired is returned when a subsystem is created without
	// a name.
	ErrSubsystemNameRequired = errors.New("subsystem name is required")
	// ErrSubsystemNotFound is returned when a project has no such subsystem.
	ErrSubsystemNotFound = errors.New("subsystem not found")
	// ErrStartingNumberInvalid is returned when issue numbering would start at
	// or below zero. Zero is rejected because the value is the first issue's
	// number, not an offset.
	ErrStartingNumberInvalid = errors.New("starting number must be greater than zero")
	// ErrSubsystemIDRequired is returned when a subsystem is created without an
	// ID. The caller has to supply it, mirroring project creation.
	ErrSubsystemIDRequired = errors.New("subsystem ID is required")
	// ErrConflict is returned when a write would duplicate an existing record:
	// a reused project or subsystem ID, or a subsystem name already used within
	// the same project. The Postgres store surfaces the schema's unique
	// violations as this same sentinel, so both stores report one conflict.
	ErrConflict = errors.New("conflicting project data")
)

// Project represents a project entity mapped from the `projects` table.
//
// Database-backed fields mirror the projects table. Nullable timestamps use
// pointers so NULL remains distinguishable from a real zero time. The remaining
// fields are API or per-user values and are not stored in projects.
type Project struct {
	ID             string     `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	ShortName      string     `json:"short_name" db:"short_name"`
	Archived       bool       `json:"archived" db:"archived"`
	Template       bool       `json:"template" db:"template"`
	Description    string     `json:"description,omitempty" db:"description"`
	OrganizationID string     `json:"organization_id,omitempty" db:"organization_id"`
	CreatedBy      string     `json:"created_by,omitempty" db:"created_by"`
	LeaderID       string     `json:"leader_id,omitempty" db:"leader_id"`
	FromEmail      string     `json:"from_email,omitempty" db:"from_email"`
	ReplyToEmail   string     `json:"reply_to_email,omitempty" db:"reply_to_email"`
	IconURL        string     `json:"icon_url,omitempty" db:"icon_url"`
	StartingNumber int64      `json:"starting_number" db:"starting_number"`
	TeamID         string     `json:"team_id,omitempty" db:"team_id"`
	IsPublic       bool       `json:"is_public" db:"is_public"`
	CreatedAt      *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" db:"updated_at"`

	ProjectTypeID string `json:"project_type_id,omitempty" db:"-"`
	Pinned        bool   `json:"pinned" db:"-"`
	Restricted    bool   `json:"restricted" db:"-"`
	HasArticles   bool   `json:"has_articles" db:"-"`
	IsDemo        bool   `json:"is_demo" db:"-"`
	IssuesURL     string `json:"issues_url,omitempty" db:"-"`
	CreationTime  int64  `json:"creation_time,omitempty" db:"-"`
}

// ProjectMember is one user attached to a project.
//
// The schema has no membership row of its own — membership is expressed by the
// team tables, the project-team group, and the project-scoped role assignments
// — so UserID is the member's identity and there is no separate membership ID.
// IsOwner reports whether the user is the project's leader.
type ProjectMember struct {
	UserID    string
	Login     string
	Email     string
	FullName  string
	Name      string
	AvatarURL string
	Banned    bool
	IsOwner   bool
	Roles     []string
	CreatedAt *time.Time
}

type TeamMember struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	GroupID       string     `json:"group_id"`
	MemberGroupID string     `json:"member_group_id"`
	MemberType    string     `json:"member_type"`
	ProjectID     string     `json:"project_id"`
	Roles         []string   `json:"roles"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	User          *User      `json:"user,omitempty"`
	Group         *Group     `json:"group,omitempty"`
}

// ProjectSubsystem is a slice of a project that issues can be assigned to.
type ProjectSubsystem struct {
	ID        string `json:"id" db:"id"`
	ProjectID string `json:"project_id" db:"project_id"`
	Name      string `json:"name" db:"name"`
	OwnerID   string `json:"owner_id,omitempty" db:"owner_id"`
	Color     int64  `json:"color" db:"color"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}

// ProjectTemplate is a preset offered when creating a project.
//
// DefaultFields is kept as a plain map rather than a typed struct: it is an
// opaque document the server stores and returns unchanged, so imposing a shape
// on it here would only invent a contract the client does not rely on.
type ProjectTemplate struct {
	ID            string            `json:"id" db:"id"`
	Name          string            `json:"name" db:"name"`
	Description   string            `json:"description" db:"description"`
	IconKey       string            `json:"icon_key" db:"icon_key"`
	DefaultFields map[string]string `json:"default_fields" db:"default_fields"`
}
