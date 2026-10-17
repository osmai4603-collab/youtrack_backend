package project

import (
	"strings"
	"testing"
)

func TestListVisibleToUserQueryRespectsRoleScope(t *testing.T) {
	for _, fragment := range []string{
		"gs.scope_type = 'GLOBAL'",
		"gs.scope_type = 'ORGANIZATION' AND gs.scope_id = COALESCE(p.organization_id, '')",
		"gs.scope_type = 'PROJECT' AND gs.scope_id = p.id",
		"ar.user_id = $1 OR ar.group_id IN (SELECT group_id FROM user_group_ids)",
		"p.created_by = $1",
		"p.leader_id = $1",
		"gm.is_team_member = TRUE",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("listVisibleToUserQuery is missing scope or existing visibility clause %q", fragment)
		}
	}
}
