package api

import (
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// InboxHandler يعالج طلبات صندوق الوارد.
type InboxHandler struct {
	app *app.App
}

func NewInboxHandler(a *app.App) *InboxHandler {
	return &InboxHandler{app: a}
}

// GetThreads يعيد خيوط الرسائل في صندوق الوارد (مطابق لـ request9.txt).
func (h *InboxHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}

	top, skip := parsePagination(r)
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))

	threads, err := h.app.GetInboxThreads(r.Context(), userID, top, skip, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}

	// YouTrack يعيد مصفوفة مباشرة في هذا الطلب
	writeModel(w, threads)
}
