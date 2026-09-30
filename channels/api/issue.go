package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// ListIssuesRequest يمثّل مخطط طلب قائمة القضايا.
type ListIssuesRequest struct {
	Query string // من query param "query"
}

// GetIssueRequest يمثّل مخطط طلب جلب قضية.
type GetIssueRequest struct {
	IssueID string // من المسار {id}
}

// CreateIssueRequest يمثّل مخطط طلب إنشاء قضية.
type CreateIssueRequest struct {
	Summary     *string `json:"summary"`
	Description string  `json:"description"`
	ProjectID   string  `json:"projectId"`
}

// IssueListResponse يمثّل مخطط استجابة قائمة القضايا.
type IssueListResponse struct {
	Issues []*model.Issue `json:"issues"`
	Count  int            `json:"count"`
}

// IssueItemResponse يمثّل مخطط استجابة قضية واحدة.
type IssueItemResponse struct {
	Issue *model.Issue `json:"issue"`
}

// CommentListItem يمثّل مخطط عنصر تعليق.
type CommentListItem struct {
	Comment *model.IssueComment `json:"comment"`
}

// IssueHandler يعالج طلبات القضايا.
type IssueHandler struct {
	app *app.YouTrackApp
}

func NewIssueHandler(a *app.YouTrackApp) *IssueHandler {
	return &IssueHandler{app: a}
}

func (api *API) InitIssue() {
	handler := NewIssueHandler(api.newApp())

	api.BaseRoutes.Issues.Method("GET", "/", api.APISessionRequiredWithPermission(app.PermissionIssueRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.List(w, r)
	}))

	api.BaseRoutes.Issues.Method("POST", "/", api.APISessionRequiredWithPermission(app.PermissionIssueWrite, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Create(w, r)
	}))

	api.BaseRoutes.APIRoot.Method("GET", "/sortedIssues", api.APISessionRequiredWithPermission(app.PermissionIssueRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetSortedIssues(w, r)
	}))

	countHandler := api.APISessionRequiredWithPermission(app.PermissionIssueRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Count(w, r)
	})
	api.BaseRoutes.APIRoot.Method("GET", "/issuesGetter/count", countHandler)
	api.BaseRoutes.APIRoot.Method("POST", "/issuesGetter/count", countHandler)

	getterHandler := api.APISessionRequiredWithPermission(app.PermissionIssueRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Getter(w, r)
	})
	api.BaseRoutes.APIRoot.Method("GET", "/issuesGetter", getterHandler)
	api.BaseRoutes.APIRoot.Method("POST", "/issuesGetter", getterHandler)

	api.BaseRoutes.Issues.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+}", api.APISessionRequiredWithPermission(app.PermissionIssueRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetByID(w, r)
	}))

	api.BaseRoutes.Issues.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+}/comments", api.APISessionRequiredWithPermission(app.PermissionIssueRead, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.Comments(w, r)
	}))
}

// List يعيد قائمة القضايا (مع بحث اختياري).
func (h *IssueHandler) List(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	req := ListIssuesRequest{Query: r.URL.Query().Get("query")}
	issues, err := h.app.ListIssues(c.AppContext, req.Query)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, IssueListResponse{Issues: issues, Count: len(issues)})
}

// GetByID يعيد قضية واحدة.
func (h *IssueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	issueID := chi.URLParam(r, "id")
	if issueID == "" {
		issueID = r.PathValue("id")
	}

	issue, err := h.app.GetIssue(c.AppContext, issueID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, IssueItemResponse{Issue: issue})
}

// Create ينشئ قضية جديدة.
func (h *IssueHandler) Create(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	var req CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.NewBadRequestError("Issue.Create", "invalid request body"))
		return
	}

	summary := ""
	if req.Summary != nil {
		summary = *req.Summary
	}

	issue, err := h.app.CreateIssue(c.AppContext, app.CreateIssueRequest{
		Summary:     summary,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		ReporterID:  c.UserID(),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, IssueItemResponse{Issue: issue})
}

// Comments يعيد تعليقات قضية.
func (h *IssueHandler) Comments(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	issueID := chi.URLParam(r, "id")
	if issueID == "" {
		issueID = r.PathValue("id")
	}

	comments, err := h.app.GetIssueComments(c.AppContext, issueID)
	if err != nil {
		writeError(w, err)
		return
	}
	list := make([]CommentListItem, 0, len(comments))
	for _, comm := range comments {
		list = append(list, CommentListItem{Comment: comm})
	}
	writeModel(w, list)
}

// Count يعيد عدد المشاكل في مجلد معيّن (مطابق لـ request17.txt).
func (h *IssueHandler) Count(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	tree := c.FieldsTree

	folderID := r.URL.Query().Get("folderId")
	if folderID == "" {
		folderID = r.URL.Query().Get("folder")
	}
	query := r.URL.Query().Get("query")
	unresolvedOnly := false
	if raw := r.URL.Query().Get("unresolvedOnly"); raw != "" {
		unresolvedOnly, _ = strconv.ParseBool(raw)
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			if folderID == "" {
				folderID = r.Form.Get("folderId")
			}
			if query == "" {
				query = r.Form.Get("query")
			}
			if raw := r.Form.Get("unresolvedOnly"); raw != "" {
				unresolvedOnly, _ = strconv.ParseBool(raw)
			}
		}
	}

	resp, err := h.app.GetIssueCount(c.AppContext, folderID, query, unresolvedOnly)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, issueCountToMap(resp, tree))
}

func issueCountToMap(resp *model.IssueCountResponse, tree *fields.FieldTree) map[string]any {
	out := map[string]any{"$type": "IssueCountResponse"}
	if resp == nil {
		return out
	}
	if tree == nil || tree.IsEmpty() || tree.Has("count") {
		out["count"] = resp.Count
	}
	if tree == nil || tree.IsEmpty() || tree.Has("folder") {
		if resp.Folder != nil {
			folderTree := tree.Child("folder")
			folder := map[string]any{"$type": resp.Folder.Type}
			addIfRequested(folder, "id", resp.Folder.ID, folderTree)
			addIfRequested(folder, "name", resp.Folder.Name, folderTree)
			addIfRequested(folder, "shortName", resp.Folder.ShortName, folderTree)
			addIfRequested(folder, "query", resp.Folder.Query, folderTree)
			if folderTree == nil || folderTree.IsEmpty() || folderTree.Has("pinned") {
				if resp.Folder.Pinned != nil {
					folder["pinned"] = *resp.Folder.Pinned
				} else {
					folder["pinned"] = nil
				}
			}
			if folderTree == nil || folderTree.IsEmpty() || folderTree.Has("pinnedInHelpdesk") {
				if resp.Folder.PinnedInHelpdesk != nil {
					folder["pinnedInHelpdesk"] = *resp.Folder.PinnedInHelpdesk
				} else {
					folder["pinnedInHelpdesk"] = nil
				}
			}
			out["folder"] = folder
		}
	}
	return out
}

func addIfRequested(out map[string]any, name string, value any, tree *fields.FieldTree) {
	if tree == nil || tree.IsEmpty() || tree.Has(name) {
		if s, ok := value.(string); ok {
			out[name] = nullOr(s)
			return
		}
		out[name] = value
	}
}

// GetSortedIssues يعيد قائمة القضايا المرتبة لمجلد أو استعلام بحث (مطابق لـ request14.txt).
func (h *IssueHandler) GetSortedIssues(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	folderID := r.URL.Query().Get("folderId")
	query := r.URL.Query().Get("query")

	topStr := r.URL.Query().Get("topRoot")
	skipStr := r.URL.Query().Get("skipRoot")

	top := 50
	skip := 0
	if topStr != "" {
		if val, err := strconv.Atoi(topStr); err == nil {
			top = val
		}
	}
	if skipStr != "" {
		if val, err := strconv.Atoi(skipStr); err == nil {
			skip = val
		}
	}

	fieldTree := c.FieldsTree
	resp, err := h.app.GetSortedIssues(c.AppContext, folderID, query, top, skip, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, resp)
}

// Getter يعيد قائمة المشاكل مع الحقول المخصصة (مطابق لـ request28.txt).
func (h *IssueHandler) Getter(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	var refs []string
	if r.Method == http.MethodPost {
		var bodyRefs []struct {
			ID         string `json:"id"`
			IDReadable string `json:"idReadable"`
		}
		if err := json.NewDecoder(r.Body).Decode(&bodyRefs); err == nil {
			for _, ref := range bodyRefs {
				if ref.ID != "" {
					refs = append(refs, ref.ID)
				} else if ref.IDReadable != "" {
					refs = append(refs, ref.IDReadable)
				}
			}
		}
	}

	query := r.URL.Query().Get("query")
	topStr := r.URL.Query().Get("$top")
	skipStr := r.URL.Query().Get("$skip")

	top := 50
	if topStr == "-1" {
		top = -1
	} else if topStr != "" {
		if val, err := strconv.Atoi(topStr); err == nil {
			top = val
		}
	}

	skip := 0
	if skipStr != "" {
		if val, err := strconv.Atoi(skipStr); err == nil {
			skip = val
		}
	}

	fieldTree := c.FieldsTree
	issues, err := h.app.GetIssuesGetter(c.AppContext, refs, query, top, skip, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(issues))
	for _, i := range issues {
		result = append(result, issueGetterToMap(i, fieldTree))
	}
	writeJSON(w, http.StatusOK, result)
}

func issueGetterToMap(i *model.IssueGetterIssue, tree *fields.FieldTree) map[string]any {
	out := map[string]any{"$type": "Issue"}
	if i == nil {
		return out
	}

	addIfRequested(out, "id", i.ID, tree)
	addIfRequested(out, "idReadable", i.IDReadable, tree)
	addIfRequested(out, "summary", i.Summary, tree)

	if tree == nil || tree.IsEmpty() || tree.Has("resolved") {
		if i.Resolved != nil {
			out["resolved"] = *i.Resolved
		} else {
			out["resolved"] = nil
		}
	}

	if tree == nil || tree.IsEmpty() || tree.Has("fields") {
		fieldTree := tree.Child("fields")
		fList := make([]map[string]any, 0, len(i.Fields))
		for _, f := range i.Fields {
			fList = append(fList, issueCustomFieldToMap(f, fieldTree))
		}
		out["fields"] = fList
	}

	return out
}

func issueCustomFieldToMap(f *model.IssueCustomField, tree *fields.FieldTree) map[string]any {
	out := map[string]any{"$type": f.Type}
	if f == nil {
		return out
	}

	addIfRequested(out, "id", f.ID, tree)
	addIfRequested(out, "name", f.Name, tree)

	if tree == nil || tree.IsEmpty() || tree.Has("value") {
		valTree := tree.Child("value")
		out["value"] = fieldValueToAny(f.Value, valTree)
	}

	if tree == nil || tree.IsEmpty() || tree.Has("projectCustomField") {
		pcfTree := tree.Child("projectCustomField")
		out["projectCustomField"] = projectCustomFieldToMap(f.ProjectCustomField, pcfTree)
	}

	return out
}

func fieldValueToAny(v any, tree *fields.FieldTree) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case *model.IssueFieldValue:
		return fieldValueToMap(val, tree)
	case []*model.IssueFieldValue:
		list := make([]map[string]any, 0, len(val))
		for _, item := range val {
			list = append(list, fieldValueToMap(item, tree))
		}
		return list
	case string:
		return val
	case int, int64, float64:
		return val
	default:
		return val
	}
}

func fieldValueToMap(v *model.IssueFieldValue, tree *fields.FieldTree) map[string]any {
	if v == nil {
		return nil
	}
	out := map[string]any{"$type": v.Type}
	addIfRequested(out, "id", v.ID, tree)
	addIfRequested(out, "name", v.Name, tree)
	addIfRequested(out, "localizedName", v.LocalizedName, tree)
	addIfRequested(out, "login", v.Login, tree)
	addIfRequested(out, "avatarUrl", v.AvatarURL, tree)
	addIfRequested(out, "presentation", v.Presentation, tree)

	if tree == nil || tree.IsEmpty() || tree.Has("minutes") {
		out["minutes"] = v.Minutes
	}

	if (tree == nil || tree.IsEmpty() || tree.Has("color")) && v.Color != nil {
		cTree := tree.Child("color")
		cMap := map[string]any{"$type": v.Color.Type}
		addIfRequested(cMap, "id", v.Color.ID, cTree)
		addIfRequested(cMap, "foreground", v.Color.Foreground, cTree)
		addIfRequested(cMap, "background", v.Color.Background, cTree)
		out["color"] = cMap
	}

	return out
}

func projectCustomFieldToMap(pcf *model.ProjectCustomField, tree *fields.FieldTree) map[string]any {
	if pcf == nil {
		return nil
	}
	out := map[string]any{"$type": pcf.Type}
	addIfRequested(out, "id", pcf.ID, tree)

	if tree == nil || tree.IsEmpty() || tree.Has("bundle") {
		bTree := tree.Child("bundle")
		bMap := map[string]any{"$type": pcf.Bundle.Type}
		addIfRequested(bMap, "id", pcf.Bundle.ID, bTree)
		out["bundle"] = bMap
	}

	if tree == nil || tree.IsEmpty() || tree.Has("field") {
		fTree := tree.Child("field")
		fMap := map[string]any{"$type": pcf.Field.Type}
		addIfRequested(fMap, "id", pcf.Field.ID, fTree)
		addIfRequested(fMap, "name", pcf.Field.Name, fTree)
		addIfRequested(fMap, "localizedName", pcf.Field.LocalizedName, fTree)

		if fTree == nil || fTree.IsEmpty() || fTree.Has("fieldType") {
			ftTree := fTree.Child("fieldType")
			ftMap := map[string]any{"$type": pcf.Field.FieldType.Type}
			addIfRequested(ftMap, "id", pcf.Field.FieldType.ID, ftTree)
			addIfRequested(ftMap, "valueType", pcf.Field.FieldType.ValueType, ftTree)
			fMap["fieldType"] = ftMap
		}
		out["field"] = fMap
	}

	return out
}
