package projectusecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"
	"unicode"

	"youtrack_backend/internal/domain"
	storageproject "youtrack_backend/internal/storage/project"
)

var (
	// ErrProjectNameRequired is returned when a project is created unnamed.
	ErrProjectNameRequired = domain.ErrProjectNameRequired
	// ErrStartingNumberInvalid is returned for a starting number below one. Zero
	// is not invalid: it means "not supplied", and is defaulted to
	// defaultStartingNumber.
	ErrStartingNumberInvalid = domain.ErrStartingNumberInvalid
	// ErrConflict is returned when the short name is already taken by another
	// project. The store reports a uniqueness violation this way rather than
	// leaking the translated database error.
	ErrConflict = domain.ErrConflict
	// ErrProjectIDUnavailable is returned when no id could be minted, so the
	// project was never written.
	ErrProjectIDUnavailable = errors.New("project id could not be generated")
)

// MaxProjectNameLength and MaxProjectShortNameLength are the accepted input
// bounds. They mirror the `projects` column widths so an over-long value is
// rejected here — as a validation failure the caller can act on — rather than
// reaching the database as a truncation error.
const (
	MaxProjectNameLength      = 255
	MaxProjectShortNameLength = 50
)

// projectIDLength and projectIDAlphabet shape the minted identifier. The
// schema stores it in VARCHAR(20), and the base36 alphabet keeps ids
// case-insensitively unique-insensitive to read, which matters because clients
// treat them as opaque handles.
const (
	projectIDLength   = 10
	projectIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// defaultStartingNumber is the sequence number the project's first issue gets.
const defaultStartingNumber = 1

// ProjectCreator is the data access contract needed to insert a project. It is
// satisfied structurally by storage/project.Repository, so no hard import ties
// this layer to the storage implementation.
//
// Create performs no authorization of its own: the route gates this use case
// behind the global create-project permission, which cannot be expressed as a
// project scope because the project does not exist yet.
type ProjectCreator interface {
	Create(ctx context.Context, project *domain.Project) error
}

// CreateProjectUseCase inserts a new project and returns it as stored.
//
// It mints the identifier, derives the short name when the caller omits it, and
// applies the defaults the schema would otherwise have to invent. It does not
// grant the creator any role on the project: membership lives in tables this
// use case does not write, so the creator is recorded as the leader (a column
// on `projects`) and nothing more.
type CreateProjectUseCase struct {
	projects ProjectCreator
	logger   *slog.Logger
	// newID is a field so a test can pin the identifier and make the assertions
	// deterministic; production always mints a random one.
	newID func() (string, error)
}

// NewCreateProject builds the create-project use case.
func NewCreateProject(logger *slog.Logger, projects ProjectCreator) *CreateProjectUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateProjectUseCase{
		projects: projects,
		logger:   logger,
		newID:    NewProjectID,
	}
}

// CreateProjectParams is the caller-supplied half of a project. Every field is
// optional except the identity and the name: a project with nothing but those
// two is a valid project, and the rest are filled from the schema defaults.
type CreateProjectParams struct {
	// UserID is the authenticated identity. It is never taken from the body,
	// so a client cannot create a project on somebody else's behalf.
	UserID         string
	Name           string
	ShortName      string
	ProjectTypeID  string
	IconURL        string
	Template       bool
	Archived       bool
	Restricted     bool
	HasArticles    bool
	IsDemo         bool
	Description    string
	IssuesURL      string
	LeaderID       string
	FromEmail      string
	ReplyToEmail   string
	TeamID         string
	OrganizationID string
	StartingNumber int64
}

// CreateProjectResult carries the created project so the caller can route to it
// without a second read.
type CreateProjectResult struct {
	Project *domain.Project
}

// Execute validates the request and inserts the project.
func (uc *CreateProjectUseCase) Execute(ctx context.Context, params *CreateProjectParams) (*CreateProjectResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}

	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if len(name) > MaxProjectNameLength {
		return nil, ErrProjectNameRequired
	}
	if params.StartingNumber < 0 {
		return nil, ErrStartingNumberInvalid
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.projects == nil {
		return nil, ErrRepositoryNotConfigured
	}

	id, err := uc.newID()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrProjectIDUnavailable, err)
	}

	startingNumber := params.StartingNumber
	if startingNumber == 0 {
		startingNumber = defaultStartingNumber
	}

	now := time.Now().UTC()
	iconURL := strings.TrimSpace(params.IconURL)
	description := strings.TrimSpace(params.Description)
	fromEmail := strings.TrimSpace(params.FromEmail)
	replyToEmail := strings.TrimSpace(params.ReplyToEmail)
	project := &domain.Project{
		ID:             id,
		Name:           name,
		ShortName:      resolveShortName(params.ShortName, name, id),
		CreatedBy:      params.UserID,
		ProjectTypeID:  strings.TrimSpace(params.ProjectTypeID),
		IconURL:        iconURL,
		Template:       params.Template,
		Archived:       params.Archived,
		Restricted:     params.Restricted,
		HasArticles:    params.HasArticles,
		IsDemo:         params.IsDemo,
		Description:    description,
		IssuesURL:      strings.TrimSpace(params.IssuesURL),
		LeaderID:       resolveLeader(params.LeaderID, params.UserID),
		FromEmail:      fromEmail,
		ReplyToEmail:   replyToEmail,
		TeamID:         strings.TrimSpace(params.TeamID),
		OrganizationID: strings.TrimSpace(params.OrganizationID),
		StartingNumber: startingNumber,
		CreatedAt:      &now,
		UpdatedAt:      &now,
		CreationTime:   now.UnixMilli(),
	}

	if err := uc.projects.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return &CreateProjectResult{Project: project}, nil
}

// resolveShortName returns the caller's short name when it supplies one, and
// otherwise derives one from the name.
//
// The column is UNIQUE, so a derived value that repeats is reported as a
// conflict rather than silently suffixed: two projects called "Mobile" should
// ask the caller to disambiguate, not invent names the caller never chose.
func resolveShortName(candidate, name, id string) string {
	if short := strings.ToUpper(strings.TrimSpace(candidate)); short != "" {
		return truncate(short, MaxProjectShortNameLength)
	}
	if derived := slugify(name); derived != "" {
		return truncate(derived, MaxProjectShortNameLength)
	}
	// A name with no letters or digits at all ("###") has nothing to derive
	// from, so the freshly minted id stands in and keeps the insert possible.
	return truncate(strings.ToUpper(id), MaxProjectShortNameLength)
}

// slugify keeps the letters and digits of a name and uppercases them, so a
// non-Latin name still yields a usable short name.
func slugify(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// resolveLeader defaults the leader to the creator. Passing someone else is
// allowed, but the project is not granted to them: membership is written
// elsewhere, so honouring the caller's choice here only sets a column.
func resolveLeader(candidate, creatorID string) string {
	if leader := strings.TrimSpace(candidate); leader != "" {
		return leader
	}
	return creatorID
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

// NewProjectID mints a random project identifier. The source is crypto/rand
// because the identifier is the only thing standing between two concurrent
// creates and a duplicate-key conflict.
func NewProjectID() (string, error) {
	limit := big.NewInt(int64(len(projectIDAlphabet)))
	out := make([]byte, projectIDLength)

	for i := range out {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("read random bytes: %w", err)
		}
		out[i] = projectIDAlphabet[n.Int64()]
	}

	return string(out), nil
}

// The project store satisfies the creator contract, so the wiring in production
// needs no adapter.
var _ ProjectCreator = (storageproject.Repository)(nil)
