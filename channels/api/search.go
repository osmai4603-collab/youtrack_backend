package api

import (
	"encoding/json"
	"net/http"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
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
	app *app.YouTrackApp
}

func NewSearchHandler(a *app.YouTrackApp) *SearchHandler {
	return &SearchHandler{app: a}
}

func (api *API) InitSearch() {
	handler := NewSearchHandler(app.New())
	api.BaseRoutes.APIRoot.Handle("/search/assist", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetAssist(w, r)
	})).Methods("POST")
}

// GetAssist يعالج طلب POST /api/search/assist ويجلب اقتراحات البحث وتنسيقه.
func (h *SearchHandler) GetAssist(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	var req SearchAssistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, model.NewBadRequestError("Search.GetAssist", "invalid request body"))
		return
	}

	fieldTree := c.FieldsTree
	assist, err := h.app.GetSearchAssist(c.AppContext, req.Query, req.Caret, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}

	writeModel(w, assist)
}
