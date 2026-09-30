package api

import (
	"net/http"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// SecuritySearchHandler يعالج طلبات البحث المتعلقة بالأمان.
type SecuritySearchHandler struct {
	app *app.YouTrackApp
}

func NewSecuritySearchHandler(a *app.YouTrackApp) *SecuritySearchHandler {
	return &SecuritySearchHandler{app: a}
}

func (api *API) InitSecuritySearch() {
	handler := NewSecuritySearchHandler(api.newApp())
	api.BaseRoutes.APIRoot.Method("GET", "/securitySearch/filterFields", api.APISessionRequiredWithPermission(app.PermissionSecuritySearch, func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetFilterFields(w, r)
	}))
}

// GetFilterFields يعالج GET /api/securitySearch/filterFields ويجلب حقول التصفية بناءً على entityType.
func (h *SecuritySearchHandler) GetFilterFields(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	entityType := r.URL.Query().Get("entityType")
	if entityType == "" {
		writeError(w, model.NewBadRequestError("SecuritySearch.GetFilterFields", "entityType query parameter is required"))
		return
	}

	fieldTree := c.FieldsTree
	rawFields, err := h.app.GetSecurityFilterFields(c.AppContext, entityType)
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
