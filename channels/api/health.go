package api

import (
	"net/http"
)

// HealthHandler يخدم فحص صحة الخدمة.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (api *API) InitHealth() {
	health := NewHealthHandler()
	api.BaseRoutes.Root.Handle("/health", api.APIHandler(func(c *Context, w http.ResponseWriter, r *http.Request) {
		health.Check(w, r)
	})).Methods("GET")
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
