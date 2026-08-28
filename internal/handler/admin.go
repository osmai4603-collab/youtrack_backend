package handler

import (
	"encoding/json"
	"net/http"
	"youtrack_backend/internal/usecase"
)

type AdminHandler struct {
	adminUC *usecase.AdminUseCase
}

func NewAdminHandler(adminUC *usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{
		adminUC: adminUC,
	}
}

// GetWorkTimeSettings يعالج طلبات /api/admin/timeTrackingSettings/workTimeSettings
func (h *AdminHandler) GetWorkTimeSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.adminUC.GetWorkTimeSettings()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(settings)
}

// GetGlobalSettings يعالج طلبات /api/admin/globalSettings
func (h *AdminHandler) GetGlobalSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.adminUC.GetGlobalSettings()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(settings)
}

// GetGeneralWidgets يعالج طلبات /api/admin/widgets/general
func (h *AdminHandler) GetGeneralWidgets(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.adminUC.GetWidgets()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(widgets)
}
