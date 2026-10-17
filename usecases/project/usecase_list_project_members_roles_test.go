package projectusecase

import (
	"context"
	"testing"

	"youtrack_backend/internal/domain"
)

type memberListReader struct {
	members []*domain.TeamMember
}

func (r memberListReader) GetListTeamMembersForProject(ctx context.Context, projectID string) ([]*domain.TeamMember, error) {
	return r.members, nil
}

type memberRolesFinder struct{}

func (memberRolesFinder) GetRoleIDsByUserForProjectScope(_ context.Context, userID, _ string) ([]string, error) {
	return []string{"user-role:" + userID}, nil
}

func (memberRolesFinder) GetRoleIDsByGroupForProjectScope(_ context.Context, groupID, _ string) ([]string, error) {
	return []string{"group-role:" + groupID}, nil
}

func TestListProjectMembersResolvesUserAndGroupRoles(t *testing.T) {
	members := []*domain.TeamMember{
		{UserID: "user-1", ProjectID: "project-1"},
		{MemberGroupID: "group-1", MemberType: "GROUP", ProjectID: "project-1"},
	}
	uc := NewListProjectMembers(nil, memberListReader{members: members}, nil, nil, memberRolesFinder{}, nil)

	result, err := uc.Execute(context.Background(), &ListProjectMembersParams{
		UserID:    "viewer-1",
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if got := result.Members[0].Roles; len(got) != 1 || got[0] != "user-role:user-1" {
		t.Errorf("user roles = %v, want [user-role:user-1]", got)
	}
	if got := result.Members[1].Roles; len(got) != 1 || got[0] != "group-role:group-1" {
		t.Errorf("group roles = %v, want [group-role:group-1]", got)
	}
}
