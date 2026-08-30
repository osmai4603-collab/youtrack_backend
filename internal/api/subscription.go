package api

import (
	"encoding/json"
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// SubscriptionHandler يعالج طلبات الاشتراك في قوائم المشاكل (Request #18).
type SubscriptionHandler struct {
	app *app.App
}

func NewSubscriptionHandler(a *app.App) *SubscriptionHandler {
	return &SubscriptionHandler{app: a}
}

// SubscribeIssueList يعالج POST /api/issueListSubscription مع دعم GET للتوافقية.
// يقرأ معامل fields ويطبق الجلب الانتقائي قبل تسلسل الاستجابة.
func (h *SubscriptionHandler) SubscribeIssueList(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}

	fieldTree := fields.Parse(r.URL.Query().Get("fields"))

	var sub *model.IssueListSubscriptionBean
	var err error

	if r.Method == http.MethodGet {
		ticket := r.URL.Query().Get("ticket")
		if ticket == "" {
			sub, err = h.app.SubscribeIssueList(r.Context(), userID, nil, fieldTree)
		} else {
			sub, err = h.app.GetIssueListSubscriptionByTicket(r.Context(), userID, ticket, fieldTree)
		}
	} else {
		var req model.IssueListSubscriptionRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		sub, err = h.app.SubscribeIssueList(r.Context(), userID, &req, fieldTree)
	}

	if err != nil {
		writeError(w, err)
		return
	}

	writeModel(w, subscriptionToMap(sub, fieldTree))
}

// subscriptionToMap يحوّل اشتراك قائمة المشاكل إلى خريطة تحترم معامل fields
// وتضمّن $type على كل مستوى كما في استجابة YouTrack الأصلية.
func subscriptionToMap(sub *model.IssueListSubscriptionBean, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if sub == nil {
		sub = &model.IssueListSubscriptionBean{}
	}

	if tree == nil || tree.IsEmpty() {
		result["id"] = nullOr(sub.ID)
		result["ticket"] = sub.Ticket
		result["query"] = sub.Query
		result["subscribe"] = sub.Subscribe
		if sub.Context != nil {
			result["context"] = subscriptionContextToMap(sub.Context)
		}
		if sub.Issues != nil {
			result["issues"] = subscriptionIssuesToMap(sub.Issues, nil)
		}
		result["folderId"] = nullOr(sub.FolderID)
	} else {
		if tree.Has("id") {
			result["id"] = nullOr(sub.ID)
		}
		if tree.Has("ticket") {
			result["ticket"] = sub.Ticket
		}
		if tree.Has("query") {
			result["query"] = sub.Query
		}
		if tree.Has("subscribe") {
			result["subscribe"] = sub.Subscribe
		}
		if tree.Has("context") {
			if sub.Context != nil {
				result["context"] = subscriptionContextToMap(sub.Context)
			} else {
				result["context"] = subscriptionContextToMap(&model.SubscriptionContext{Type: "Project", ID: "0-0"})
			}
		}
		if tree.Has("issues") && sub.Issues != nil {
			result["issues"] = subscriptionIssuesToMap(sub.Issues, tree.Child("issues"))
		}
		if tree.Has("folderId") {
			result["folderId"] = nullOr(sub.FolderID)
		}
	}
	result["$type"] = "IssueListSubscriptionBean"
	return result
}

// subscriptionContextToMap يحوّل سياق الاشتراك إلى خريطة مع $type.
func subscriptionContextToMap(c *model.SubscriptionContext) map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{
		"id":    c.ID,
		"$type": c.Type,
	}
}

// subscriptionIssuesToMap يحوّل قائمة المشاكل المرتبطة بالاشتراك إلى مصفوفة
// خرائط تحترم شجرة الحقول وتضمّن $type.
func subscriptionIssuesToMap(items []*model.IssueListSubscriptionItem, tree *fields.FieldTree) []any {
	result := make([]any, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		m := make(map[string]any)
		if tree == nil || tree.IsEmpty() || tree.Has("id") {
			m["id"] = it.ID
		}
		if tree == nil || tree.IsEmpty() || tree.Has("matches") {
			m["matches"] = it.Matches
		}
		m["$type"] = "IssueListSubscriptionItem"
		result = append(result, m)
	}
	return result
}
