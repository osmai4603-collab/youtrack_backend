package api

import (
	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// isRequest27 يتحقق مما إذا كانت شجرة الحقول تطابق نمط الطلب #27 حصريًا:
// fields=team(name,users(id,login,name,avatarUrl,email)),leader(id)
//
// يجب أن يتطابق النمط بدقة حتى لا يُوجَّه الطلبات الأخرى (مثل #21 الذي يحوي
// id,name,shortName,plugins...) إلى هذا المسار. لذلك نتحقق من:
//   - الجذر لا يحتوي إلا على leader و team حصراً
//   - leader يحتوي على id فقط
//   - team يحتوي على name و users فقط (بدون id أو auditTargetId...)
func isRequest27(tree *fields.FieldTree) bool {
	if tree == nil || tree.Children == nil {
		return false
	}
	if len(tree.Children) != 2 {
		return false
	}
	leader := tree.Child("leader")
	team := tree.Child("team")
	if leader == nil || team == nil {
		return false
	}
	// leader(id) فقط
	if len(leader.Children) != 1 || !leader.Has("id") {
		return false
	}
	// team(name, users(...)) فقط
	if len(team.Children) != 2 || !team.Has("name") || !team.Has("users") {
		return false
	}
	return true
}

// project27ToMap يحوّل ProjectTeamAndLeader إلى خريطة JSON تطابق استجابة request27.txt حرفياً:
// الحقول المطلوبة فقط + $type على كل مستوى، مع إمكانية إخراج email/avatarUrl بقيمة null.
func project27ToMap(p *model.ProjectTeamAndLeader, leaderTree, teamTree *fields.FieldTree) map[string]any {
	result := make(map[string]any)

	if p.Leader != nil && (leaderTree == nil || leaderTree.Has("id")) {
		leader := make(map[string]any)
		leader["id"] = p.Leader.ID
		if leaderTree == nil || leaderTree.Has("$type") {
			if p.Leader.Type != "" {
				leader["$type"] = p.Leader.Type
			}
		}
		// $type يُضمَّن دائماً حسب استجابة YouTrack حتى لو لم يُطلب صراحة
		if p.Leader.Type != "" {
			leader["$type"] = p.Leader.Type
		}
		result["leader"] = leader
	}

	if p.Team != nil && (teamTree == nil || teamTree.Has("name") || teamTree.Has("users")) {
		team := make(map[string]any)
		if teamTree == nil || teamTree.Has("name") {
			team["name"] = p.Team.Name
		}
		if teamTree == nil || teamTree.Has("users") {
			if len(p.Team.Users) > 0 {
				users := make([]any, 0, len(p.Team.Users))
				var userTree *fields.FieldTree
				if teamTree != nil {
					userTree = teamTree.Child("users")
				}
				for _, u := range p.Team.Users {
					users = append(users, user27ToMap(u, userTree))
				}
				team["users"] = users
			}
		}
		if p.Team.Type != "" {
			team["$type"] = p.Team.Type
		}
		result["team"] = team
	}

	if p.Type != "" {
		result["$type"] = p.Type
	}
	return result
}

// user27ToMap يحوّل User27 إلى خريطة JSON تحترم شجرة الحقول الفرعية.
func user27ToMap(u *model.User27, tree *fields.FieldTree) map[string]any {
	m := make(map[string]any)
	if tree == nil || tree.Has("id") {
		m["id"] = u.ID
	}
	if tree == nil || tree.Has("login") {
		m["login"] = u.Login
	}
	if tree == nil || tree.Has("name") {
		m["name"] = u.Name
	}
	if tree == nil || tree.Has("avatarUrl") {
		m["avatarUrl"] = u.AvatarURL
	}
	if tree == nil || tree.Has("email") {
		m["email"] = u.Email
	}
	if u.Type != "" {
		m["$type"] = u.Type
	}
	return m
}
