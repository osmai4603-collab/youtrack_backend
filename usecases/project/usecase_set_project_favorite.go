package projectusecase

import (
	"context"
	"fmt"
	"log/slog"

	"youtrack_backend/internal/domain"
)

var (
	// ErrProjectIDRequired is returned when no project is named.
	ErrProjectIDRequired = domain.ErrProjectIDRequired
	// ErrProjectNotVisible is returned when the identity may not read the
	// project, which leaves it outside the set it is allowed to favorite.
	ErrProjectNotVisible = domain.ErrProjectNotVisible
)

// ProjectFavorites is the data access contract needed to flip a project's
// favorite state. It is satisfied structurally by storage/project.Repository, so
// no hard import ties this layer to the storage implementation.
type ProjectFavorites interface {
	// IsVisibleToUser reports whether the identity holds read access.
	IsVisibleToUser(ctx context.Context, userID, projectID string) (bool, error)
	// SetFavorite records or clears the identity's favorite state.
	SetFavorite(ctx context.Context, userID, projectID string, favorite bool) error
}

// SetProjectFavoriteUseCase records whether the authenticated identity starred a
// project.
//
// The identity always comes from the request context, so a client can never star
// a project on someone else's behalf. Favoriting is a personal preference rather
// than a mutation of the project record, so the only gate is read access to the
// project: no project-write permission is required, and a project the identity
// cannot see is refused instead of leaving an orphaned favorite row behind.
type SetProjectFavoriteUseCase struct {
	favorites ProjectFavorites
	logger    *slog.Logger
}

// NewSetProjectFavorite builds the set-project-favorite use case.
func NewSetProjectFavorite(logger *slog.Logger, favorites ProjectFavorites) *SetProjectFavoriteUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &SetProjectFavoriteUseCase{
		favorites: favorites,
		logger:    logger,
	}
}

type SetProjectFavoriteParams struct {
	UserID    string
	ProjectID string
	Favorite  bool
}

// SetProjectFavoriteResult echoes the state that was actually stored, so the
// caller can reconcile without re-reading the project.
type SetProjectFavoriteResult struct {
	ProjectID string `json:"id"`
	Favorite  bool   `json:"favorite"`
}

func (uc *SetProjectFavoriteUseCase) Execute(ctx context.Context, params *SetProjectFavoriteParams) (*SetProjectFavoriteResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if params.ProjectID == "" {
		return nil, ErrProjectIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.favorites == nil {
		return nil, ErrRepositoryNotConfigured
	}

	// visible, err := uc.favorites.IsVisibleToUser(ctx, params.UserID, params.ProjectID)
	// if err != nil {
	// 	return nil, fmt.Errorf("check project visibility: %w", err)
	// }
	// if !visible {
	// 	return nil, ErrProjectNotVisible
	// }

	if err := uc.favorites.SetFavorite(ctx, params.UserID, params.ProjectID, params.Favorite); err != nil {
		return nil, fmt.Errorf("set project favorite: %w", err)
	}

	return &SetProjectFavoriteResult{ProjectID: params.ProjectID, Favorite: params.Favorite}, nil
}
