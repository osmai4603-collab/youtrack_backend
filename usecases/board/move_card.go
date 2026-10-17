package boardusecase

import (
	"context"
	"errors"
	"log/slog"

	"youtrack_backend/internal/domain"
	boardstorage "youtrack_backend/internal/storage/board"
)

var (
	// ErrIssueIDRequired is returned when the request names no issue.
	ErrIssueIDRequired = domain.ErrIssueIDRequired
	// ErrIssueNotVisible is returned when the issue does not exist or the
	// caller may not reach it.
	ErrIssueNotVisible = domain.ErrIssueNotVisible
	// ErrInvalidState is returned when the requested state is not one of the
	// three accepted values.
	ErrInvalidState = errors.New("invalid issue state")
)

// MoveCardUseCase updates an issue's state when a board card is dragged to a
// different column.
type MoveCardUseCase struct {
	board  boardstorage.Repository
	logger *slog.Logger
}

// NewMoveCard builds the move-card use case.
func NewMoveCard(logger *slog.Logger, board boardstorage.Repository) *MoveCardUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &MoveCardUseCase{
		board:  board,
		logger: logger,
	}
}

// MoveCardParams carries the input for a card move.
type MoveCardParams struct {
	UserID   string
	IssueID  string
	NewState string
}

// Execute validates the new state and delegates the update to the repository.
func (uc *MoveCardUseCase) Execute(ctx context.Context, params *MoveCardParams) error {
	if params == nil || params.UserID == "" {
		return ErrUserIDRequired
	}
	if params.IssueID == "" {
		return ErrIssueIDRequired
	}
	if !domain.ValidIssueState(params.NewState) {
		return ErrInvalidState
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if uc.board == nil {
		return errors.New("board repository is not configured")
	}

	if err := uc.board.UpdateIssueState(ctx, params.IssueID, params.NewState); err != nil {
		uc.logger.LogAttrs(ctx, slog.LevelError, "move card failed",
			slog.String("issue_id", params.IssueID),
			slog.String("new_state", params.NewState),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
