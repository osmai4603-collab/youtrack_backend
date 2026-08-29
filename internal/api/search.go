package api

import (
	"encoding/json"
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// SearchAssistRequest يمثّل هيكل الطلب القادم لنقطة النهاية api/search/assist.
type SearchAssistRequest struct {
	Query    string `json:"query"`
	Caret    int    `json:"caret"`
	Type     string `json:"type"` // e.g., "Issue" or "Article"
	Helpdesk bool   `json:"helpdesk"`
}

// SearchHandler يعالج العمليات المتعلقة بالبحث.
type SearchHandler struct {
	app *app.App
}

func NewSearchHandler(a *app.App) *SearchHandler {
	return &SearchHandler{app: a}
}

// GetAssist يعالج طلب POST /api/search/assist ويجلب اقتراحات البحث وتنسيقه.
func (h *SearchHandler) GetAssist(w http.ResponseWriter, r *http.Request) {
	var req SearchAssistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.BadRequest("invalid request body: %v", err))
		return
	}

	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	assist, err := h.app.GetSearchAssist(r.Context(), req.Query, req.Caret, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}

	writeModel(w, assist)
}
