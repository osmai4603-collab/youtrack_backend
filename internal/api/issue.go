package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
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
	app *app.App
}

func NewIssueHandler(a *app.App) *IssueHandler {
	return &IssueHandler{app: a}
}

// List يعيد قائمة القضايا (مع بحث اختياري).
func (h *IssueHandler) List(w http.ResponseWriter, r *http.Request) {
	req := ListIssuesRequest{Query: r.URL.Query().Get("query")}
	issues, err := h.app.ListIssues(r.Context(), req.Query)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, IssueListResponse{Issues: issues, Count: len(issues)})
}

// GetByID يعيد قضية واحدة.
func (h *IssueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	req := GetIssueRequest{IssueID: r.PathValue("id")}
	issue, err := h.app.GetIssue(r.Context(), req.IssueID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, IssueItemResponse{Issue: issue})
}

// Create ينشئ قضية جديدة.
func (h *IssueHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.BadRequest("invalid request body"))
		return
	}

	reporterID, _ := CurrentUserID(r)
	summary := ""
	if req.Summary != nil {
		summary = *req.Summary
	}

	issue, err := h.app.CreateIssue(r.Context(), app.CreateIssueRequest{
		Summary:     summary,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		ReporterID:  reporterID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, IssueItemResponse{Issue: issue})
}

// Comments يعيد تعليقات قضية.
func (h *IssueHandler) Comments(w http.ResponseWriter, r *http.Request) {
	req := GetIssueRequest{IssueID: r.PathValue("id")}
	comments, err := h.app.GetIssueComments(r.Context(), req.IssueID)
	if err != nil {
		writeError(w, err)
		return
	}
	list := make([]CommentListItem, 0, len(comments))
	for _, c := range comments {
		list = append(list, CommentListItem{Comment: c})
	}
	writeModel(w, list)
}

// Count يعيد عدد المشاكل في مجلد معيّن (مطابق لـ request17.txt).
// يدعم GET (query params) و POST (form body) احتياطياً.
func (h *IssueHandler) Count(w http.ResponseWriter, r *http.Request) {
	tree := fields.Parse(r.URL.Query().Get("fields"))

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

	resp, err := h.app.GetIssueCount(r.Context(), folderID, query, unresolvedOnly)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, issueCountToMap(resp, tree))
}

// issueCountToMap يحوّل استجابة العد إلى map وفق شجرة الحقول مع إدراج $type دائماً.
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

// addIfRequested يضيف حقلاً إلى الخريطة فقط عندما يكون مطلوباً في شجرة الحقول.
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
	folderID := r.URL.Query().Get("folderId")
	query := r.URL.Query().Get("query")

	// topRoot و skipRoot تُستخدم لترقيم الصفحات في هذا الطلب
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

	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	resp, err := h.app.GetSortedIssues(r.Context(), folderID, query, top, skip, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, resp)
}

// Getter يعيد قائمة المشاكل مع الحقول المخصصة (مطابق لـ request28.txt).
// يدعم كلاً من GET و POST (حيث يمكن تمرير قائمة معرّفات في جسم الطلب).
func (h *IssueHandler) Getter(w http.ResponseWriter, r *http.Request) {
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

	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	issues, err := h.app.GetIssuesGetter(r.Context(), refs, query, top, skip, fieldTree)
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

// issueGetterToMap يحوّل IssueGetterIssue إلى خريطة تحترم شجرة الحقول وتضمّن $type.
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
