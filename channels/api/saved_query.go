package api

import (
	"net/http"
	"strings"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// SavedQueriesHandler يعالج طلبات الاستعلامات المحفوظة (Request #15).
type SavedQueriesHandler struct {
	app *app.YouTrackApp
}

func NewSavedQueriesHandler(a *app.YouTrackApp) *SavedQueriesHandler {
	return &SavedQueriesHandler{app: a}
}

func (api *API) InitSavedQuery() {
	handler := NewSavedQueriesHandler(app.New())
	api.BaseRoutes.APIRoot.Handle("/savedQueries", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.List(w, r)
	})).Methods("GET")
}

// List يعيد قائمة الاستعلامات المحفوظة مع احترام معامل fields.
func (h *SavedQueriesHandler) List(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	fieldTree := c.FieldsTree
	top, skip := parsePagination(r)

	items, err := h.app.GetSavedQueries(c.AppContext, fieldTree, top, skip)
	if err != nil {
		writeError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, sq := range items {
		result = append(result, savedQueryToMap(sq, fieldTree))
	}
	writeModel(w, result)
}

func savedQueryToMap(sq *model.SavedQuery, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if sq == nil {
		sq = &model.SavedQuery{}
	}

	if tree == nil || tree.IsEmpty() {
		result["id"] = sq.ID
		result["issuesUrl"] = nullOr(sq.IssuesURL)
		result["name"] = sq.Name
		result["query"] = sq.Query
		result["pinnedByDefault"] = sq.PinnedByDefault
		result["pinned"] = sq.Pinned
		result["pinnedInHelpdesk"] = sq.PinnedInHelpdesk
		result["isUpdatable"] = sq.IsUpdatable
		result["isDeletable"] = sq.IsDeletable
		result["isShareable"] = sq.IsShareable
		if sq.Owner != nil {
			result["owner"] = savedQueryUserToMap(sq.Owner, nil)
		}
		if sq.ReadSharingSettings != nil {
			result["readSharingSettings"] = sharingSettingsToMap(sq.ReadSharingSettings, nil)
		}
		if sq.UpdateSharingSettings != nil {
			result["updateSharingSettings"] = sharingSettingsToMap(sq.UpdateSharingSettings, nil)
		}
		result["sortOrder"] = savedQueryOrderToMap(sq.SortOrderSortable)
	} else {
		if tree.Has("id") {
			result["id"] = sq.ID
		}
		if tree.Has("issuesUrl") {
			result["issuesUrl"] = nullOr(sq.IssuesURL)
		}
		if tree.Has("name") {
			result["name"] = sq.Name
		}
		if tree.Has("query") {
			result["query"] = sq.Query
		}
		if tree.Has("pinnedByDefault") {
			result["pinnedByDefault"] = sq.PinnedByDefault
		}
		if tree.Has("pinned") {
			result["pinned"] = sq.Pinned
		}
		if tree.Has("pinnedInHelpdesk") {
			result["pinnedInHelpdesk"] = sq.PinnedInHelpdesk
		}
		if tree.Has("isUpdatable") {
			result["isUpdatable"] = sq.IsUpdatable
		}
		if tree.Has("isDeletable") {
			result["isDeletable"] = sq.IsDeletable
		}
		if tree.Has("isShareable") {
			result["isShareable"] = sq.IsShareable
		}
		if tree.Has("owner") && sq.Owner != nil {
			result["owner"] = savedQueryUserToMap(sq.Owner, tree.Child("owner"))
		}
		if tree.Has("readSharingSettings") && sq.ReadSharingSettings != nil {
			result["readSharingSettings"] = sharingSettingsToMap(sq.ReadSharingSettings, tree.Child("readSharingSettings"))
		}
		if tree.Has("updateSharingSettings") && sq.UpdateSharingSettings != nil {
			result["updateSharingSettings"] = sharingSettingsToMap(sq.UpdateSharingSettings, tree.Child("updateSharingSettings"))
		}
		if tree.Has("sortOrder") {
			result["sortOrder"] = savedQueryOrderToMap(sq.SortOrderSortable)
		}
	}
	result["$type"] = "SavedQuery"
	return result
}

func savedQueryUserToMap(u *model.User, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if u == nil {
		return nil
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = u.ID
		result["login"] = u.Login
		result["email"] = nullOr(u.Email)
		result["fullName"] = nullOr(u.FullName)
		result["avatarUrl"] = nullOr(u.AvatarURL)
		if u.UserType != nil {
			result["userType"] = userTypeToMap(u.UserType)
		}
		result["name"] = u.Name
		result["isEmailVerified"] = u.IsEmailVerified
		result["guest"] = u.Guest
		result["online"] = u.Online
		result["banned"] = u.Banned
		result["banBadge"] = nullOrPtr(u.BanBadge)
		result["canReadProfile"] = u.CanReadProfile
		result["isLocked"] = u.IsLocked
	} else {
		if tree.Has("id") {
			result["id"] = u.ID
		}
		if tree.Has("login") {
			result["login"] = u.Login
		}
		if tree.Has("email") {
			result["email"] = nullOr(u.Email)
		}
		if tree.Has("fullName") {
			result["fullName"] = nullOr(u.FullName)
		}
		if tree.Has("avatarUrl") {
			result["avatarUrl"] = nullOr(u.AvatarURL)
		}
		if tree.Has("userType") && u.UserType != nil {
			result["userType"] = userTypeToMap(u.UserType)
		}
		if tree.Has("name") {
			result["name"] = u.Name
		}
		if tree.Has("isEmailVerified") {
			result["isEmailVerified"] = u.IsEmailVerified
		}
		if tree.Has("guest") {
			result["guest"] = u.Guest
		}
		if tree.Has("online") {
			result["online"] = u.Online
		}
		if tree.Has("banned") {
			result["banned"] = u.Banned
		}
		if tree.Has("banBadge") {
			result["banBadge"] = nullOrPtr(u.BanBadge)
		}
		if tree.Has("canReadProfile") {
			result["canReadProfile"] = u.CanReadProfile
		}
		if tree.Has("isLocked") {
			result["isLocked"] = u.IsLocked
		}
	}
	result["$type"] = "User"
	return result
}

func userTypeToMap(t *model.UserType) map[string]any {
	if t == nil {
		return nil
	}
	return map[string]any{
		"id":    t.ID,
		"name":  t.Name,
		"$type": "UserType",
	}
}

func sharingSettingsToMap(s *model.SavedQuerySharingSettings, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if s == nil {
		s = &model.SavedQuerySharingSettings{}
	}

	if tree == nil || tree.IsEmpty() || tree.Has("permittedGroups") {
		var groupTree *fields.FieldTree
		if tree != nil {
			groupTree = tree.Child("permittedGroups")
		}
		groups := make([]map[string]any, 0, len(s.PermittedGroups))
		for _, g := range s.PermittedGroups {
			groups = append(groups, permittedGroupToMap(g, groupTree))
		}
		result["permittedGroups"] = groups
	}
	if tree == nil || tree.IsEmpty() || tree.Has("permittedUsers") {
		var userTree *fields.FieldTree
		if tree != nil {
			userTree = tree.Child("permittedUsers")
		}
		users := make([]map[string]any, 0, len(s.PermittedUsers))
		for _, u := range s.PermittedUsers {
			users = append(users, savedQueryUserToMap(u, userTree))
		}
		result["permittedUsers"] = users
	}
	result["$type"] = "WatchFolderSharingSettings"
	return result
}

func permittedGroupToMap(g *model.SavedQueryGroup, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if g == nil {
		g = &model.SavedQueryGroup{}
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = g.ID
		result["name"] = g.Name
		result["auditTargetId"] = nullOr(g.AuditTargetID)
		result["description"] = nullOr(g.Description)
		result["allUsersGroup"] = g.AllUsersGroup
		result["icon"] = nullOr(g.Icon)
		if g.TeamForProject != nil {
			result["teamForProject"] = projectRefToMap(g.TeamForProject, nil)
		} else {
			result["teamForProject"] = nil
		}
		result["isUpdatable"] = g.IsUpdatable
		result["isRemovable"] = g.IsRemovable
	} else {
		if tree.Has("id") {
			result["id"] = g.ID
		}
		if tree.Has("name") {
			result["name"] = g.Name
		}
		if tree.Has("auditTargetId") {
			result["auditTargetId"] = nullOr(g.AuditTargetID)
		}
		if tree.Has("description") {
			result["description"] = nullOr(g.Description)
		}
		if tree.Has("allUsersGroup") {
			result["allUsersGroup"] = g.AllUsersGroup
		}
		if tree.Has("icon") {
			result["icon"] = nullOr(g.Icon)
		}
		if tree.Has("teamForProject") {
			if g.TeamForProject != nil {
				result["teamForProject"] = projectRefToMap(g.TeamForProject, tree.Child("teamForProject"))
			} else {
				result["teamForProject"] = nil
			}
		}
		if tree.Has("isUpdatable") {
			result["isUpdatable"] = g.IsUpdatable
		}
		if tree.Has("isRemovable") {
			result["isRemovable"] = g.IsRemovable
		}
	}
	result["$type"] = queryGroupTypeName(g)
	return result
}

func projectRefToMap(p *model.ProjectRef, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if p == nil {
		return nil
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = p.ID
		result["name"] = p.Name
		result["icon"] = nullOr(p.Icon)
	} else {
		if tree.Has("id") {
			result["id"] = p.ID
		}
		if tree.Has("name") {
			result["name"] = p.Name
		}
		if tree.Has("icon") {
			result["icon"] = nullOr(p.Icon)
		}
	}
	result["$type"] = "Project"
	return result
}

func savedQueryOrderToMap(isSortable bool) map[string]any {
	return map[string]any{
		"isSortable": isSortable,
		"$type":      "SavedQueryIssueOrder",
	}
}

func queryGroupTypeName(g *model.SavedQueryGroup) string {
	if g == nil {
		return "UserGroup"
	}
	lowerType := strings.ToLower(g.GroupType)
	if strings.Contains(lowerType, "registered") || strings.Contains(strings.ToLower(g.Name), "registered") {
		return "RegisteredUsersGroup"
	}
	if strings.Contains(lowerType, "all") || strings.Contains(strings.ToLower(g.Name), "all users") {
		return "AllUsersGroup"
	}
	if g.Type != "" {
		return g.Type
	}
	return "UserGroup"
}
