package api

import (
	"net/http"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
)

// InboxHandler يعالج طلبات صندوق الوارد.
type InboxHandler struct {
	app *app.YouTrackApp
}

func NewInboxHandler(a *app.YouTrackApp) *InboxHandler {
	return &InboxHandler{app: a}
}

func (api *API) InitInbox() {
	handler := NewInboxHandler(app.New())
	api.BaseRoutes.Inbox.Handle("/threads", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetThreads(w, r)
	})).Methods("GET")
}

// GetThreads يعيد خيوط الرسائل في صندوق الوارد (مطابق لـ request9.txt).
func (h *InboxHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("Inbox.GetThreads", "user context missing"))
		return
	}

	top, skip := parsePagination(r)
	fieldTree := c.FieldsTree

	threads, err := h.app.GetInboxThreads(c.AppContext, c.UserID(), top, skip, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}

	writeModel(w, threads)
}
