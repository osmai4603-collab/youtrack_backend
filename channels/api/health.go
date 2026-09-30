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
	readyHandler := api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		health.Check(w, r)
	})
	api.BaseRoutes.Root.Method("GET", "/health", readyHandler)
	api.BaseRoutes.Root.Method("GET", "/ready", readyHandler)
	api.BaseRoutes.Root.Method("GET", "/live", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		health.Live(w, r)
	}))
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

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.server == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "live"})
}
