package boardusecase

import (
	"context"
	"errors"
	"log/slog"

	"youtrack_backend/internal/domain"
	boardstorage "youtrack_backend/internal/storage/board"
)

var (
	ErrUserIDRequired    = domain.ErrUserIDRequired
	ErrProjectIDRequired = domain.ErrProjectIDRequired
)

// GetBoardDetailsUseCase resolves the Agile Board view for a project.
//
// Authorization is handled by the HTTP middleware (RequireProject with
// PermIssueReadIssue), so this use case trusts that the caller has already been
// verified. It delegates entirely to the board repository, which assembles the
// swimlane structure from issues, subsystems, and sprints.
type GetBoardDetailsUseCase struct {
	board  boardstorage.Repository
	logger *slog.Logger
}

// NewGetBoardDetails builds the use case.
func NewGetBoardDetails(logger *slog.Logger, board boardstorage.Repository) *GetBoardDetailsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetBoardDetailsUseCase{
		board:  board,
		logger: logger,
	}
}

// GetBoardDetailsParams carries the input for the board-details query.
type GetBoardDetailsParams struct {
	UserID    string
	ProjectID string
	SprintID  string // optional
}

// Execute fetches the full board payload.
func (uc *GetBoardDetailsUseCase) Execute(ctx context.Context, params *GetBoardDetailsParams) (*domain.AgileBoardData, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if params.ProjectID == "" {
		return nil, ErrProjectIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if uc.board == nil {
		return nil, errors.New("board repository is not configured")
	}

	data, err := uc.board.GetBoardDetails(ctx, params.ProjectID, params.SprintID)
	if err != nil {
		uc.logger.LogAttrs(ctx, slog.LevelError, "board details query failed",
			slog.String("project_id", params.ProjectID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return data, nil
}
