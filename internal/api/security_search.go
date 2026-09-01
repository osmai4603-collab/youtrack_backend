package api

import (
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// SecuritySearchHandler يعالج طلبات البحث المتعلقة بالأمان.
type SecuritySearchHandler struct {
	app *app.App
}

func NewSecuritySearchHandler(a *app.App) *SecuritySearchHandler {
	return &SecuritySearchHandler{app: a}
}

// GetFilterFields يعالج GET /api/securitySearch/filterFields ويجلب حقول التصفية بناءً على entityType.
func (h *SecuritySearchHandler) GetFilterFields(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entityType")
	if entityType == "" {
		writeError(w, model.BadRequest("entityType query parameter is required"))
		return
	}

	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	rawFields, err := h.app.GetSecurityFilterFields(r.Context(), entityType)
	if err != nil {
		writeError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(rawFields))
	for _, f := range rawFields {
		result = append(result, securityFilterFieldToMap(f, fieldTree))
	}

	writeModel(w, result)
}

// securityFilterFieldToMap يحوّل حقل تصفية أمان إلى خريطة تحترم معامل fields وتضمّن $type دائماً.
func securityFilterFieldToMap(f *model.SecurityFilterField, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if tree == nil || tree.IsEmpty() {
		result["id"] = f.ID
		result["name"] = f.Name
	} else {
		if tree.Has("id") {
			result["id"] = f.ID
		}
		if tree.Has("name") {
			result["name"] = f.Name
		}
	}
	result["$type"] = f.Type
	return result
}
