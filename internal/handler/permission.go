package handler

import (
	"encoding/json"
	"net/http"
	"youtrack_backend/internal/usecase"
)

type PermissionHandler struct {
	configUC *usecase.ConfigUseCase
}

func NewPermissionHandler(configUC *usecase.ConfigUseCase) *PermissionHandler {
	return &PermissionHandler{
		configUC: configUC,
	}
}

// GetPermissionsCache يعالج طلبات /api/permissions/cache
func (h *PermissionHandler) GetPermissionsCache(w http.ResponseWriter, r *http.Request) {
	cachedPermissions := h.configUC.GetPermissionsCache()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cachedPermissions)
}
