package sqlstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// InboxStore تطبيق عمليات صندوق الوارد على PostgreSQL.
type InboxStore struct {
	db *pgxpool.Pool
}

// Threads يجلب خيوط الرسائل للمستخدم مع دعم الجلب الانتقائي عبر FieldTree.
func (s *InboxStore) Threads(ctx context.Context, userID string, top int, skip int, tree *fields.FieldTree) ([]*model.InboxThread, error) {
	columns, _ := s.getThreadColumns(tree)
	selectClause := ""
	for i, col := range columns {
		if i > 0 {
			selectClause += ", "
		}
		selectClause += col
	}

	query := fmt.Sprintf("SELECT %s FROM inbox_threads WHERE user_id = $1 ORDER BY updated DESC LIMIT $2 OFFSET $3", selectClause)
	rows, err := s.db.Query(ctx, query, userID, top, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threads := []*model.InboxThread{}
	for rows.Next() {
		t := &model.InboxThread{Type: "InboxThread"}
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}
			s.mapColumnToThread(t, col, val)
		}

		// جلب الرسائل إذا طُلبت
		if tree == nil || tree.Has("messages") {
			msgs, err := s.getMessages(ctx, t.ID, tree.Child("messages"))
			if err == nil {
				t.Messages = msgs
			}
		}

		threads = append(threads, t)
	}

	return threads, rows.Err()
}

func (s *InboxStore) getThreadColumns(tree *fields.FieldTree) ([]string, map[string]bool) {
	// الخريطة الافتراضية إذا كان المخطط فارغاً
	defaultCols := []string{"id", "read", "muted", "notified", "target_type", "thread_id", "timestamp", "updated", "subject_text", "subject_target_id"}
	if tree == nil || tree.IsEmpty() {
		return defaultCols, nil
	}

	colMap := make(map[string]bool)
	cols := []string{}

	// الحقول الأساسية
	addCol := func(c string) {
		if !colMap[c] {
			colMap[c] = true
			cols = append(cols, c)
		}
	}

	if tree.Has("id") {
		addCol("id")
	}
	if tree.Has("read") {
		addCol("read")
	}
	if tree.Has("muted") {
		addCol("muted")
	}
	if tree.Has("notified") {
		addCol("notified")
	}
	if tree.Has("targetType") {
		addCol("target_type")
	}
	if tree.Has("threadId") {
		addCol("thread_id")
	}
	if tree.Has("timestamp") {
		addCol("timestamp")
	}
	if tree.Has("updated") {
		addCol("updated")
	}

	// معالجة الموضوع (Subject)
	if tree.Has("subject") {
		addCol("subject_text")
		addCol("subject_target_id")
	}

	// إذا لم يتم طلب أي شيء من الأعمدة الأساسية ولكن طُلبت الرسائل مثلاً، نحتاج المعرف لربطهم
	if len(cols) == 0 {
		return defaultCols, nil
	}
	// التأكد دائماً من وجود id لربط العلاقات إذا طُلبت
	if tree.Has("messages") || tree.Has("subject") {
		addCol("id")
	}

	return cols, colMap
}

func (s *InboxStore) mapColumnToThread(t *model.InboxThread, col string, val any) {
	switch col {
	case "id":
		t.ID = val.(string)
	case "read":
		t.Read = val.(bool)
	case "muted":
		t.Muted = val.(bool)
	case "notified":
		t.Notified = val.(bool)
	case "target_type":
		t.TargetType = val.(string)
	case "thread_id":
		t.ThreadID = val.(string)
	case "timestamp":
		t.Timestamp = val.(int64)
	case "updated":
		t.Updated = val.(int64)
	case "subject_text":
		if t.Subject == nil {
			t.Subject = &model.InboxSubject{Type: "InboxSubject"}
		}
		t.Subject.Text = val.(string)
	case "subject_target_id":
		if t.Subject == nil {
			t.Subject = &model.InboxSubject{Type: "InboxSubject"}
		}
		t.Subject.ID = val.(string)
	}
}

func (s *InboxStore) getMessages(ctx context.Context, threadID string, tree *fields.FieldTree) ([]*model.InboxMessage, error) {
	columns, _ := s.getMessageColumns(tree)
	selectClause := ""
	for i, col := range columns {
		if i > 0 {
			selectClause += ", "
		}
		selectClause += col
	}

	query := fmt.Sprintf("SELECT %s FROM inbox_messages WHERE thread_id = $1 ORDER BY timestamp ASC", selectClause)
	rows, err := s.db.Query(ctx, query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs := []*model.InboxMessage{}
	for rows.Next() {
		m := &model.InboxMessage{TypeField: "InboxMessage"}
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}
			s.mapColumnToMessage(m, col, val, tree)
		}

		// جلب البارامترات إذا طُلبت
		if tree == nil || tree.Has("params") {
			params, err := s.getMessageParams(ctx, m.ID)
			if err == nil {
				m.Params = params
			}
		}

		// جلب الأنشطة إذا طُلبت
		if tree == nil || tree.Has("activities") {
			activities, err := s.getActivities(ctx, m.ID, tree.Child("activities"))
			if err == nil {
				m.Activities = activities
			}
		}

		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (s *InboxStore) getMessageColumns(tree *fields.FieldTree) ([]string, map[string]bool) {
	defaultCols := []string{"id", "author_id", "author_group_id", "timestamp", "text", "type", "pseudo", "empty_field_text", "target_id", "target_type"}
	if tree == nil || tree.IsEmpty() {
		return defaultCols, nil
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
	if tree.Has("timestamp") {
		addCol("timestamp")
	}
	if tree.Has("text") {
		addCol("text")
	}
	if tree.Has("type") {
		addCol("type")
	}
	if tree.Has("pseudo") {
		addCol("pseudo")
	}
	if tree.Has("emptyFieldText") {
		addCol("empty_field_text")
	}
	if tree.Has("author") {
		addCol("author_id")
	}
	if tree.Has("authorGroup") {
		addCol("author_group_id")
	}

	if len(cols) == 0 {
		return defaultCols, nil
	}
	// Always include ID for relations
	if tree.Has("params") || tree.Has("activities") {
		addCol("id")
	}

	return cols, colMap
}

func (s *InboxStore) mapColumnToMessage(m *model.InboxMessage, col string, val any, tree *fields.FieldTree) {
	switch col {
	case "id":
		m.ID = val.(string)
	case "timestamp":
		m.Timestamp = val.(int64)
	case "text":
		m.Text = val.(string)
	case "type":
		m.Type = val.(string)
	case "pseudo":
		m.Pseudo = val.(bool)
	case "empty_field_text":
		m.EmptyFieldText = val.(string)
	case "author_id":
		m.Author = &model.User{ID: val.(string), Type: "User"}
	case "author_group_id":
		m.AuthorGroup = &model.UserGroup{ID: val.(string)}
	}
}

func (s *InboxStore) getMessageParams(ctx context.Context, messageID string) ([]*model.InboxMessageParam, error) {
	rows, err := s.db.Query(ctx, `SELECT key, value FROM inbox_message_params WHERE message_id = $1`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	params := []*model.InboxMessageParam{}
	for rows.Next() {
		p := &model.InboxMessageParam{Type: "InboxMessageParam"}
		if err := rows.Scan(&p.Key, &p.Value); err == nil {
			params = append(params, p)
		}
	}
	return params, nil
}

func (s *InboxStore) getActivities(ctx context.Context, messageID string, tree *fields.FieldTree) ([]*model.InboxActivity, error) {
	rows, err := s.db.Query(ctx, `SELECT id, category_id, added_id, removed_id, target_id, target_type, timestamp FROM inbox_activities WHERE message_id = $1`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []*model.InboxActivity{}
	for rows.Next() {
		a := &model.InboxActivity{Type: "InboxActivity"}
		var catID, addedID, removedID, targetID, targetType *string
		var ts int64
		if err := rows.Scan(&a.ID, &catID, &addedID, &removedID, &targetID, &targetType, &ts); err == nil {
			if catID != nil {
				a.Category = &model.ActivityCategory{ID: *catID, Type: "ActivityCategory"}
			}
			// يمكن التوسع هنا لجلب تفاصيل added/removed بناءً على النوع
			activities = append(activities, a)
		}
	}
	return activities, nil
}
