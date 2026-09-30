package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
)

// GetProjectRequest يمثّل مخطط طلب جلب مشروع.
type GetProjectRequest struct {
	ProjectID string
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
	app *app.YouTrackApp
}

func NewProjectHandler(a *app.YouTrackApp) *ProjectHandler {
	return &ProjectHandler{app: a}
}

func (api *API) InitProject() {
	handler := NewProjectHandler(api.newApp())

	api.BaseRoutes.Projects.Method("GET", "/", api.APISessionRequiredWithPermission("project.read", func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.List(w, r)
	}))

	api.BaseRoutes.Projects.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+}", api.APISessionRequiredWithPermission("project.read", func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetByID(w, r)
	}))

	api.BaseRoutes.Admin.Method("GET", "/projects", api.APISessionRequiredWithPermission("system.admin", func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.List(w, r)
	}))

	api.BaseRoutes.Admin.Method("GET", "/projects/{id:[A-Za-z0-9_\\-\\.]+}", api.APISessionRequiredWithPermission("system.admin", func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetByID(w, r)
	}))
}

// List يعيد قائمة المشاريع.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	projects, err := h.app.GetAllProjects(c.AppContext)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, ProjectListResponse{Projects: projects, Count: len(projects)})
}

// GetByID يعيد مشروعًا بمعرّفه أو رمزه مع دعم جلب الحقول المخصصة.
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		projectID = r.PathValue("id")
	}

	tree := c.FieldsTree

	if isRequest27(tree) {
		res, err := h.app.GetProjectTeamAndLeader(c.AppContext, projectID, tree.Child("leader"), tree.Child("team"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeModel(w, project27ToMap(res, tree.Child("leader"), tree.Child("team")))
		return
	}

	project, err := h.app.GetProject(c.AppContext, projectID, tree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, project)
}
