package api

import (
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// GetProjectRequest يمثّل مخطط طلب جلب مشروع.
type GetProjectRequest struct {
	ProjectID string // من المسار {id}
}

// ProjectListResponse يمثّل مخطط استجابة قائمة المشاريع.
type ProjectListResponse struct {
	Projects []*model.Project `json:"projects"`
	Count    int              `json:"count"`
}

// ProjectItemResponse يمثّل مخطط استجابة مشروع واحد.
type ProjectItemResponse struct {
	Project *model.Project `json:"project"`
}

// ProjectHandler يعالج طلبات المشاريع.
type ProjectHandler struct {
	app *app.App
}

func NewProjectHandler(a *app.App) *ProjectHandler {
	return &ProjectHandler{app: a}
}

// List يعيد قائمة المشاريع.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.app.GetAllProjects(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, ProjectListResponse{Projects: projects, Count: len(projects)})
}

// GetByID يعيد مشروعًا بمعرّفه أو رمزه مع دعم جلب الحقول المخصصة.
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	fieldsParam := r.URL.Query().Get("fields")
	tree := fields.Parse(fieldsParam)

	project, err := h.app.GetProject(r.Context(), projectID, tree)
	if err != nil {
		writeError(w, err)
		return
	}
	// نرسل المشروع مباشرة بدون تغليف ليتوافق مع YouTrack API
	writeModel(w, project)
}
