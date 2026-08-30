package sqlstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// SubscriptionStore تطبيق عمليات الاشتراك في قوائم المشاكل على PostgreSQL
// (Request #18: /api/issueListSubscription).
type SubscriptionStore struct {
	db *pgxpool.Pool
}

// subDefaultColumns أعمدة جدول issue_list_subscriptions الإفتراضية.
var subDefaultColumns = []string{
	"id", "ticket", "query", "subscribe", "context_type", "context_id", "folder_id",
}

// subColumns يبني قائمة أعمدة issue_list_subscriptions المطلوبة حسب شجرة
// الحقول، مع إضافة السياق دائمًا عن جمع أعمدة context.
func subColumns(tree *fields.FieldTree) []string {
	if tree == nil || tree.IsEmpty() {
		return subDefaultColumns
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
	if tree.Has("ticket") {
		addCol("ticket")
	}
	if tree.Has("query") {
		addCol("query")
	}
	if tree.Has("subscribe") {
		addCol("subscribe")
	}
	if tree.Has("context") {
		addCol("context_type")
		addCol("context_id")
	}
	if tree.Has("folderId") {
		addCol("folder_id")
	}

	if len(cols) == 0 {
		return subDefaultColumns
	}
	return cols
}

// SubscribeIssueList ينشئ أو يجلب اشتراكًا في قائمة المشاكل مع توليد تذكرة
// والجلب الانتقائي حسب شجرة الحقول (مطابق لـ request18.txt).
func (s *SubscriptionStore) SubscribeIssueList(ctx context.Context, userID string, req *model.IssueListSubscriptionRequest, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	query := ""
	subscribe := true
	contextType := "Project"
	contextID := "0-0"
	folderID := ""
	if req != nil {
		if req.Query != "" {
			query = req.Query
		}
		if req.Subscribe != nil {
			subscribe = *req.Subscribe
		}
		if req.Context != nil {
			if req.Context.Type != "" {
				contextType = req.Context.Type
			}
			if req.Context.ID != "" {
				contextID = req.Context.ID
			}
		}
		if req.FolderID != "" {
			folderID = req.FolderID
		}
	}

	ticket := uuid.NewString()
	var subID int
	err := s.db.QueryRow(ctx, `
		INSERT INTO issue_list_subscriptions
			(ticket, user_id, query, subscribe, context_type, context_id, folder_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`, ticket, nullStr(userID), query, subscribe, contextType, contextID, nullStr(folderID)).
		Scan(&subID)
	if err != nil {
		return nil, err
	}

	if req != nil && len(req.Issues) > 0 {
		for i, it := range req.Issues {
			if it == nil {
				continue
			}
			issueID := it.ID
			matches := it.Matches
			if _, err := s.db.Exec(ctx, `
				INSERT INTO issue_list_subscription_issues (subscription_id, issue_id, matches, ordinal)
				VALUES ($1, $2, $3, $4)`, subID, issueID, matches, i); err != nil {
				return nil, err
			}
		}
	}

	return s.getByID(ctx, userID, subID, tree)
}

// GetIssueListSubscriptionByTicket يجلب اشتراكًا حسب تذكرته (الطريقة GET).
func (s *SubscriptionStore) GetIssueListSubscriptionByTicket(ctx context.Context, userID string, ticket string, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	var subID int
	err := s.db.QueryRow(ctx, `SELECT id FROM issue_list_subscriptions WHERE ticket = $1`, ticket).Scan(&subID)
	if err != nil {
		return nil, err
	}
	return s.getByID(ctx, userID, subID, tree)
}

// getByID يبني كائن الاشتراك من الصف مع الجلب الانتقائي الدقيق حسب شجرة
// الحقول لكل جدول.
func (s *SubscriptionStore) getByID(ctx context.Context, userID string, subID int, tree *fields.FieldTree) (*model.IssueListSubscriptionBean, error) {
	columns := subColumns(tree)
	selectClause := strings.Join(columns, ", ")
	query := fmt.Sprintf("SELECT %s FROM issue_list_subscriptions WHERE id = $1", selectClause)

	rows, err := s.db.Query(ctx, query, subID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	values, err := rows.Values()
	if err != nil {
		return nil, err
	}

	sub := &model.IssueListSubscriptionBean{Type: "IssueListSubscriptionBean"}
	for i, col := range columns {
		if values[i] == nil {
			continue
		}
		s.mapColumnToSubscription(sub, col, values[i])
	}

	// جلب السياق فقط إذا طُلب
	if tree == nil || tree.Has("context") {
		sub.Context = &model.SubscriptionContext{
			Type: sub.ContextType,
			ID:   sub.ContextID,
		}
	}

	// جلب المشاكل المرتبطة فقط إذا طُلب
	if tree == nil || tree.Has("issues") {
		var issueTree *fields.FieldTree
		if tree != nil {
			issueTree = tree.Child("issues")
		}
		issues, err := s.getIssues(ctx, subID, issueTree)
		if err == nil {
			sub.Issues = issues
		}
	}
	return sub, nil
}

// mapColumnToSubscription يملأ حقول الاشتراك من عمود مُرجع من قاعدة البيانات.
func (s *SubscriptionStore) mapColumnToSubscription(sub *model.IssueListSubscriptionBean, col string, val any) {
	switch col {
	case "id":
		sub.ID = intToString(val)
	case "ticket":
		sub.Ticket = val.(string)
	case "query":
		sub.Query = val.(string)
	case "subscribe":
		sub.Subscribe = val.(bool)
	case "context_type":
		sub.ContextType = val.(string)
	case "context_id":
		sub.ContextID = val.(string)
	case "folder_id":
		sub.FolderID = val.(string)
	}
}

// getIssues يجلب المشاكل المرتبطة بالاشتراك مع احترام شجرة الحقول.
func (s *SubscriptionStore) getIssues(ctx context.Context, subID int, tree *fields.FieldTree) ([]*model.IssueListSubscriptionItem, error) {
	cols := []string{"issue_id", "matches"}
	if tree != nil && !tree.IsEmpty() {
		if !tree.Has("id") {
			cols = []string{"matches"}
		}
		if !tree.Has("matches") {
			cols = []string{"issue_id"}
		}
	}
	selectClause := strings.Join(cols, ", ")
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM issue_list_subscription_issues
		WHERE subscription_id = $1 ORDER BY ordinal ASC`, selectClause), subID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.IssueListSubscriptionItem{}
	for rows.Next() {
		it := &model.IssueListSubscriptionItem{Type: "IssueListSubscriptionItem"}
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		for i, col := range cols {
			if values[i] == nil {
				continue
			}
			switch col {
			case "issue_id":
				it.ID = values[i].(string)
			case "matches":
				it.Matches = values[i].(bool)
			}
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// nullStr يحوّل سلسلة فارغة إلى NULL.
func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// intToString يحوّل قيمة int/Int32 إلى سلسلة.
func intToString(v any) string {
	switch t := v.(type) {
	case int:
		return fmt.Sprintf("%d", t)
	case int32:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}
