package api

import (
	"net/http"
	"time"
)

// HealthResponse يمثّل مخطط استجابة فحص الحالة.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// HealthHandler يعالج فحص حالة الخادم.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	writeModel(w, HealthResponse{
		Status:    "OK",
		Timestamp: time.Now().UTC(),
		Service:   "youtrack-backend",
	})
}
