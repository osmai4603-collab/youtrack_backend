package sqlstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// SavedQueryStore تطبيق عمليات الاستعلامات المحفوظة على PostgreSQL (Request #15).
type SavedQueryStore struct {
	db *pgxpool.Pool
}

// sqDefaultColumns أعمدة جدول saved_queries الإفتراضية مع المفاتيح الأجنبية اللازمة
// للعلاقات (owner, مشاركة القراءة/التحديث).
var sqDefaultColumns = []string{
	"id", "issues_url", "name", "query", "pinned", "pinned_by_default",
	"pinned_in_helpdesk", "is_updatable", "is_deletable", "is_shareable",
	"sort_order_sortable", "owner_id", "visible_for_id", "updateable_by_id",
}

// sqColumns يبني قائمة أعمدة saved_queries المطلوبة حسب شجرة الحقول.
func sqColumns(tree *fields.FieldTree) []string {
	if tree == nil || tree.IsEmpty() {
		return sqDefaultColumns
	}

	colMap := make(map[string]bool)
	cols := []string{}
	addCol := func(c string) {
		if !colMap[c] {
			colMap[c] = true
			cols = append(cols, c)
		}
	}

	if tree.Has("id") {
		addCol("id")
	}
	if tree.Has("issuesUrl") {
		addCol("issues_url")
	}
	if tree.Has("name") {
		addCol("name")
	}
	if tree.Has("query") {
		addCol("query")
	}
	if tree.Has("pinnedByDefault") {
		addCol("pinned_by_default")
	}
	if tree.Has("pinned") {
		addCol("pinned")
	}
	if tree.Has("pinnedInHelpdesk") {
		addCol("pinned_in_helpdesk")
	}
	if tree.Has("isUpdatable") {
		addCol("is_updatable")
	}
	if tree.Has("isDeletable") {
		addCol("is_deletable")
	}
	if tree.Has("isShareable") {
		addCol("is_shareable")
	}
	if tree.Has("sortOrder") {
		addCol("sort_order_sortable")
	}
	// المفاتيح الأجنبية فقط عندما يُطلب كائن العلاقة
	if tree.Has("owner") {
		addCol("owner_id")
	}
	if tree.Has("readSharingSettings") {
		addCol("visible_for_id")
	}
	if tree.Has("updateSharingSettings") {
		addCol("updateable_by_id")
	}

	if len(cols) == 0 {
		return sqDefaultColumns
	}
	return cols
}

// SavedQueries يجلب الاستعلامات المحفوظة مع الجلب الانتقائي الدقيق لكل جدول
// حسب شجرة الحقول FieldTree (مطابق لـ request15.txt).
func (s *SavedQueryStore) SavedQueries(ctx context.Context, tree *fields.FieldTree, top int, skip int) ([]*model.SavedQuery, error) {
	columns := sqColumns(tree)
	selectClause := strings.Join(columns, ", ")

	query := fmt.Sprintf("SELECT %s FROM saved_queries ORDER BY name ASC", selectClause)
	var args []any
	if top > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = []any{top, skip}
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.SavedQuery{}
	for rows.Next() {
		sq := &model.SavedQuery{Type: "SavedQuery"}
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		for i, col := range columns {
			if values[i] == nil {
				continue
			}
			s.mapColumnToSavedQuery(sq, col, values[i])
		}

		// جلب المالك إذا طُلب
		if tree == nil || tree.Has("owner") {
			if sq.OwnerID != "" {
				if owner, err := s.getOwner(ctx, sq.OwnerID, tree.Child("owner")); err == nil {
					sq.Owner = owner
				}
			}
		}

		// جلب إعدادات مشاركة القراءة إذا طُلب
		if tree == nil || tree.Has("readSharingSettings") {
			if settings, err := s.getSharingSettings(ctx, sq.VisibleForID, tree.Child("readSharingSettings")); err == nil {
				sq.ReadSharingSettings = settings
			}
		}

		// جلب إعدادات مشاركة التحديث إذا طُلب
		if tree == nil || tree.Has("updateSharingSettings") {
			if settings, err := s.getSharingSettings(ctx, sq.UpdateableByID, tree.Child("updateSharingSettings")); err == nil {
				sq.UpdateSharingSettings = settings
			}
		}

		items = append(items, sq)
	}

	return items, rows.Err()
}

// mapColumnToSavedQuery يملأ حقول الاستعلام المحفوظ من عمود يُرجع من قاعدة البيانات.
func (s *SavedQueryStore) mapColumnToSavedQuery(sq *model.SavedQuery, col string, val any) {
	switch col {
	case "id":
		sq.ID = val.(string)
	case "issues_url":
		sq.IssuesURL = val.(string)
	case "name":
		sq.Name = val.(string)
	case "query":
		sq.Query = val.(string)
	case "pinned":
		sq.Pinned = val.(bool)
	case "pinned_by_default":
		sq.PinnedByDefault = val.(bool)
	case "pinned_in_helpdesk":
		sq.PinnedInHelpdesk = val.(bool)
	case "is_updatable":
		sq.IsUpdatable = val.(bool)
	case "is_deletable":
		sq.IsDeletable = val.(bool)
	case "is_shareable":
		sq.IsShareable = val.(bool)
	case "sort_order_sortable":
		sq.SortOrderSortable = val.(bool)
	case "owner_id":
		sq.OwnerID = val.(string)
	case "visible_for_id":
		sq.VisibleForID = val.(string)
	case "updateable_by_id":
		sq.UpdateableByID = val.(string)
	}
}

// getOwner يجلب مالك الاستعلام مع اسم نوع المستخدم إذا طُلب.
func (s *SavedQueryStore) getOwner(ctx context.Context, userID string, tree *fields.FieldTree) (*model.User, error) {
	u := &model.User{Type: "User"}
	err := s.db.QueryRow(ctx, `
		SELECT id, login, COALESCE(email,''), COALESCE(full_name,''), COALESCE(name,''), COALESCE(avatar_url,''),
		       COALESCE(user_type_id,''), is_email_verified, guest, online, banned, ban_badge, can_read_profile, is_locked
		FROM users WHERE id = $1`, userID).
		Scan(&u.ID, &u.Login, &u.Email, &u.FullName, &u.Name, &u.AvatarURL, &u.UserTypeID,
			&u.IsEmailVerified, &u.Guest, &u.Online, &u.Banned, &u.BanBadge, &u.CanReadProfile, &u.IsLocked)
	if err != nil {
		return nil, err
	}
	if tree == nil || tree.Has("userType") {
		if u.UserTypeID != "" {
			var typeName string
			if err := s.db.QueryRow(ctx, `SELECT name FROM user_types WHERE id = $1`, u.UserTypeID).Scan(&typeName); err == nil && typeName != "" {
				u.UserType = &model.UserType{ID: u.UserTypeID, Name: typeName, Type: "UserType"}
			}
		}
	}
	return u, nil
}

// getSharingSettings يبني إعدادات المشاركة للمجموعة المحددة (المستخدمون المصرّح لهم
// غير مخزَّنين في المخطط الحالي فتُرجَع دائمًا مصفوفة فارغة).
func (s *SavedQueryStore) getSharingSettings(ctx context.Context, groupID string, tree *fields.FieldTree) (*model.SavedQuerySharingSettings, error) {
	settings := &model.SavedQuerySharingSettings{
		PermittedGroups: []*model.SavedQueryGroup{},
		PermittedUsers:  []*model.User{},
		Type:            "WatchFolderSharingSettings",
	}
	if groupID == "" {
		return settings, nil
	}

	var groupTree *fields.FieldTree
	if tree != nil {
		groupTree = tree.Child("permittedGroups")
	}
	g, err := s.getGroup(ctx, groupID, groupTree)
	if err == nil {
		settings.PermittedGroups = []*model.SavedQueryGroup{g}
	}
	return settings, nil
}

// getGroup يجلب مجموعة مستخدمين واحدة مع الفريق المرتبط إذا طُلب.
func (s *SavedQueryStore) getGroup(ctx context.Context, groupID string, tree *fields.FieldTree) (*model.SavedQueryGroup, error) {
	g := &model.SavedQueryGroup{}
	err := s.db.QueryRow(ctx, `
		SELECT id, COALESCE(name,''), COALESCE(group_type,''), all_users_group, COALESCE(icon,''),
		       COALESCE(description,''), COALESCE(audit_target_id,''), is_updatable, is_removable, COALESCE(team_for_project_id,'')
		FROM user_groups WHERE id = $1`, groupID).
		Scan(&g.ID, &g.Name, &g.GroupType, &g.AllUsersGroup, &g.Icon, &g.Description,
			&g.AuditTargetID, &g.IsUpdatable, &g.IsRemovable, &g.TeamForProjectID)
	if err != nil {
		return nil, err
	}
	g.Type = groupTypeName(g.GroupType, g.AllUsersGroup, g.Name)

	if tree == nil || tree.Has("teamForProject") {
		if g.TeamForProjectID != "" {
			var p model.ProjectRef
			if err := s.db.QueryRow(ctx, `SELECT id, COALESCE(name,''), COALESCE(icon_url,'') FROM projects WHERE id = $1`, g.TeamForProjectID).
				Scan(&p.ID, &p.Name, &p.Icon); err == nil {
				p.Type = "Project"
				g.TeamForProject = &p
			}
		}
	}
	return g, nil
}

// groupTypeName يشتق الاسم الديناميكي ($type) للمجموعة كما في YouTrack.
func groupTypeName(groupType string, allUsersGroup bool, name string) string {
	if allUsersGroup {
		return "AllUsersGroup"
	}
	lowerType := strings.ToLower(groupType)
	if strings.Contains(lowerType, "registered") || strings.Contains(strings.ToLower(name), "registered") {
		return "RegisteredUsersGroup"
	}
	if strings.Contains(lowerType, "all") || strings.Contains(strings.ToLower(name), "all users") {
		return "AllUsersGroup"
	}
	return "UserGroup"
}
