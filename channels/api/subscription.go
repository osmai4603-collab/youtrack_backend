package api

import (
	"encoding/json"
	"net/http"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// SubscriptionHandler يعالج طلبات الاشتراك في قائمة المشاكل (Request #18).
type SubscriptionHandler struct {
	app *app.YouTrackApp
}

func NewSubscriptionHandler(a *app.YouTrackApp) *SubscriptionHandler {
	return &SubscriptionHandler{app: a}
}

func (api *API) InitSubscription() {
	handler := NewSubscriptionHandler(api.newApp())
	subHandler := api.APISessionRequiredWithPermission(app.PermissionSubscriptionRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.SubscribeIssueList(w, r)
	})
	api.BaseRoutes.APIRoot.Method("GET", "/issueListSubscription", subHandler)
	api.BaseRoutes.APIRoot.Method("POST", "/issueListSubscription", subHandler)
}

// SubscribeIssueList يعالج طلبات الاشتراك بقائمة المشاكل.
func (h *SubscriptionHandler) SubscribeIssueList(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("Subscription.SubscribeIssueList", "user context missing"))
		return
	}

	fieldTree := c.FieldsTree
	var sub *model.IssueListSubscriptionBean
	var err *model.AppError

	if r.Method == http.MethodGet {
		ticket := r.URL.Query().Get("ticket")
		if ticket == "" {
			sub, err = h.app.SubscribeIssueList(c.AppContext, c.UserID(), nil, fieldTree)
		} else {
			sub, err = h.app.GetIssueListSubscriptionByTicket(c.AppContext, c.UserID(), ticket, fieldTree)
		}
	} else {
		var req model.IssueListSubscriptionRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		sub, err = h.app.SubscribeIssueList(c.AppContext, c.UserID(), &req, fieldTree)
	}

	if err != nil {
		writeError(w, err)
		return
	}

	writeModel(w, subscriptionToMap(sub, fieldTree))
}

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

func subscriptionContextToMap(c *model.SubscriptionContext) map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{
		"id":    c.ID,
		"$type": c.Type,
	}
}

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
