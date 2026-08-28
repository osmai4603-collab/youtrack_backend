package handler

import (
	"encoding/json"
	"net/http"
	"youtrack_backend/internal/usecase"
)

type ConfigHandler struct {
	configUC *usecase.ConfigUseCase
}

func NewConfigHandler(configUC *usecase.ConfigUseCase) *ConfigHandler {
	return &ConfigHandler{
		configUC: configUC,
	}
}

// GetConfig يعالج طلبات /api/config
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.configUC.GetFrontendConfig()
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cfg)
}
