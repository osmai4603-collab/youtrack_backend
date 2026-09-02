package api

import (
	"net/http"

	"youtrack_backend/channels/app"
)

// HealthHandler يخدم فحص صحة الخدمة.
type HealthHandler struct {
	server *app.YouTrackServer
}

func NewHealthHandler(srv *app.YouTrackServer) *HealthHandler {
	return &HealthHandler{server: srv}
}

func (api *API) InitHealth() {
	health := NewHealthHandler(api.srv)
	api.BaseRoutes.Root.Handle("/health", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		health.Check(w, r)
	})).Methods("GET")
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	payload := map[string]any{"status": "healthy"}
	if h == nil || h.server == nil || !h.server.IsReady() {
		status = http.StatusServiceUnavailable
		payload = map[string]any{"status": "unavailable", "ready": false}
	} else {
		payload = map[string]any{"status": "healthy", "ready": true}
	}
	writeJSON(w, status, payload)
}
