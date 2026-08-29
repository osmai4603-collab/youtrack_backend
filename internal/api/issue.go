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
