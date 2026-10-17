// Package board provides the persistence layer for the Agile Board view.
//
// It queries the issues, project_subsystems, and sprints tables to assemble the
// board payload the frontend expects. The board itself is a projection of
// existing data rather than a distinct aggregate, so there is no dedicated
// board table — the view is built from the issue's state, the project's
// subsystems, and the project's sprints.
package board

import (
	"context"

	"youtrack_backend/internal/domain"
)

// Repository is the data-access contract the board use cases depend on.
type Repository interface {
	// GetBoardDetails fetches the issues, subsystems, and sprints needed to
	// render the Agile Board for a project. When sprintID is non-empty, only
	// issues assigned to that sprint are included.
	GetBoardDetails(ctx context.Context, projectID string, sprintID string) (*domain.AgileBoardData, error)

	// UpdateIssueState sets the state column of a single issue and bumps its
	// updated timestamp. The caller is responsible for validating that newState
	// is one of the three accepted values.
	UpdateIssueState(ctx context.Context, issueID string, newState string) error
}
