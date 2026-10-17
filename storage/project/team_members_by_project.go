package project

import (
	"context"
	"log/slog"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
)

// listProjectMembersQuery builds the roster from the places a project acquires
// members: its leader, the members of its team group (a group of group_type
// 'TEAM' referenced by projects.team_id), and the identities holding a role
// scoped to the project — either directly or through a group they belong to.
//
// UNION (not UNION ALL) removes the overlap for free — the leader is normally
// also in the team, and reporting them twice would be wrong. Roles come from the
// project-scoped assignments only; a global or organization-scoped role is not
// a membership in this project.
const queryTeamMembersByProject = `
	SELECT gm.id, COALESCE(gm.user_id, ''), COALESCE(gm.member_group_id, ''), gm.member_type FROM group_members gm WHERE gm.is_team_member = TRUE AND EXISTS (
		SELECT 1 FROM projects p WHERE p.id = $1 AND p.team_id = gm.group_id
	)
`

func (r *PostgresRepository) GetListTeamMembersForProject(ctx context.Context, projectID string) ([]*domain.TeamMember, error) {
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}

	rows, err := r.db.Query(ctx, queryTeamMembersByProject, projectID)
	if err != nil {
		return nil, postgres.TranslateError(projectMembersOp, err)
	}
	defer rows.Close()

	members := make([]*domain.TeamMember, 0, 8)
	for rows.Next() {
		var member domain.TeamMember
		err := rows.Scan(&member.ID, &member.UserID, &member.MemberGroupID, &member.MemberType)
		if err != nil {
			return nil, postgres.TranslateError(projectMembersOp, err)
		}
		member.ProjectID = projectID
		members = append(members, &member)
	}

	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(projectMembersOp, err)
	}

	logProjectMembers(ctx, projectMembersOp, members, slog.String("projectID", projectID))
	return members, nil
}
