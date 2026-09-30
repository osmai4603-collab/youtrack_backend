package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Params يجمّع كل المعاملات المستخرجة من الطلب مركزياً.
type Params struct {
	UserID    string
	ProjectID string
	IssueID   string
	Page      int
	PerPage   int
	Query     string
	Fields    string
	Top       int
	Skip      int
}

const (
	DefaultPerPage = 60
	MaxPerPage     = 200
)

// ParamsFromRequest يستخرج كل المعاملات من الطلب.
func ParamsFromRequest(r *http.Request) *Params {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		userID = r.PathValue("id")
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		projectID = r.PathValue("projectId")
	}

	issueID := chi.URLParam(r, "issueId")
	if issueID == "" {
		issueID = r.PathValue("issueId")
	}

	q := r.URL.Query()
	p := &Params{
		UserID:    userID,
		ProjectID: projectID,
		IssueID:   issueID,
		Query:     q.Get("query"),
		Fields:    q.Get("fields"),
	}

	if page, err := strconv.Atoi(q.Get("$skip")); err == nil {
		p.Skip = page
	}
	if perPage, err := strconv.Atoi(q.Get("$top")); err == nil && perPage > 0 {
		if perPage > MaxPerPage {
			perPage = MaxPerPage
		}
		p.Top = perPage
	} else {
		p.Top = DefaultPerPage
	}

	return p
}
