package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"youtrack_backend/channels/app"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// GetUserRequest يمثّل مخطط طلب جلب مستخدم (معرّف مستخدم من المسار).
type GetUserRequest struct {
	UserID string // من المسار {id}
}

type UserHandler struct {
	app *app.YouTrackApp
}

func NewUserHandler(a *app.YouTrackApp) *UserHandler { return &UserHandler{app: a} }

func (h *UserHandler) context(r *http.Request) *Context { return ContextFromRequest(h.app, r) }

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	getMe(h.context(r), w, r)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	getUsers(h.context(r), w, r)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	getUserByID(h.context(r), w, r)
}

func (h *UserHandler) GetGrazieProfile(w http.ResponseWriter, r *http.Request) {
	getGrazieProfile(h.context(r), w, r)
}

func (h *UserHandler) GetGeneralProfile(w http.ResponseWriter, r *http.Request) {
	getGeneralProfile(h.context(r), w, r)
}

func (h *UserHandler) GetQuestionnaireProfile(w http.ResponseWriter, r *http.Request) {
	getQuestionnaireProfile(h.context(r), w, r)
}

func (h *UserHandler) GetRecentIssues(w http.ResponseWriter, r *http.Request) {
	getRecentIssues(h.context(r), w, r)
}

func (h *UserHandler) GetRecentArticles(w http.ResponseWriter, r *http.Request) {
	getRecentArticles(h.context(r), w, r)
}

func (h *UserHandler) GetHubMe(w http.ResponseWriter, r *http.Request) {
	getHubMe(h.context(r), w, r)
}

func (h *UserHandler) GetInboxFolders(w http.ResponseWriter, r *http.Request) {
	getInboxFolders(h.context(r), w, r)
}

func (api *API) InitUser() {
	api.BaseRoutes.Users.Method("GET", "/", api.APISessionRequiredWithPermission("system.admin", getUsers))
	api.BaseRoutes.Users.Method("GET", "/me", api.APISessionRequiredWithPermission("profile.read", getMe))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+}", api.APISessionRequiredWithPermission("profile.read", getUserByID))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/profiles/grazie", api.APISessionRequiredWithPermission("profile.read", getGrazieProfile))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/profiles/general", api.APISessionRequiredWithPermission("profile.read", getGeneralProfile))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/profiles/questionnaire", api.APISessionRequiredWithPermission("profile.read", getQuestionnaireProfile))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/recentIssues", api.APISessionRequiredWithPermission("profile.read", getRecentIssues))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/recentArticles", api.APISessionRequiredWithPermission("profile.read", getRecentArticles))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/hubMe", api.APISessionRequiredWithPermission("profile.read", getHubMe))
	api.BaseRoutes.Users.Method("GET", "/{id:[A-Za-z0-9_\\-\\.]+|me}/folders", api.APISessionRequiredWithPermission("profile.read", getInboxFolders))
}

// GetMe يعيد المستخدم الحالي مع كامل مرفقاته وإعداداته مباشرة في جذر الـ JSON (مطابق لـ request1.txt).
func getMe(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetMe", "user context missing"))
		return
	}

	fieldTree := c.FieldsTree
	user, err := c.App.GetCurrentUser(c.AppContext, c.UserID(), fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, user)
}

// List يعيد قائمة المستخدمين بصيغة مصفوفة JSON مباشرة (مطابق لـ request32.txt).
func getUsers(c *Context, w http.ResponseWriter, r *http.Request) {
	users, err := c.App.GetAllUsers(c.AppContext)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, users)
}

// GetByID يعيد مستخدمًا بمعرّفه مباشرة في جذر الـ JSON (مطابق لـ request33.txt).
func getUserByID(c *Context, w http.ResponseWriter, r *http.Request) {
	userID := ""
	if c.Params != nil {
		userID = c.Params.UserID
	}
	if userID == "" {
		userID = chi.URLParam(r, "id")
	}
	if userID == "" {
		userID = r.PathValue("id")
	}
	if userID == "me" {
		getMe(c, w, r)
		return
	}
	user, err := c.App.GetUserByID(c.AppContext, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, user)
}

// GetGrazieProfile يعيد إعدادات الذكاء الاصطناعي والتدقيق اللغوي Grazie (مطابق لـ request4.txt).
func getGrazieProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetGrazieProfile", "user context missing"))
		return
	}
	grazie, err := c.App.GetGrazieProfile(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, grazie)
}

// GetGeneralProfile يعيد الإعدادات العامة لملف المستخدم (مطابق لـ request29.txt).
func getGeneralProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetGeneralProfile", "user context missing"))
		return
	}
	gen, err := c.App.GetGeneralProfile(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	fieldTree := c.FieldsTree
	if fieldTree == nil || fieldTree.IsEmpty() {
		writeModel(w, gen)
		return
	}
	writeModel(w, generalProfileToMap(gen, fieldTree))
}

func generalProfileToMap(p *model.GeneralUserProfile, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if p == nil {
		p = &model.GeneralUserProfile{}
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = p.ID
		result["timezone"] = p.Timezone
		result["dateFieldFormat"] = p.DateFormat
		result["locale"] = p.Locale
		result["semanticSearchForArticles"] = p.SemanticSearchForArticles
		result["lastCreatedIssue"] = p.LastCreatedIssue
		result["searchContext"] = p.SearchContext
		result["helpdeskContext"] = p.HelpdeskContext
	} else {
		if tree.Has("id") {
			result["id"] = p.ID
		}
		if tree.Has("timezone") {
			result["timezone"] = p.Timezone
		}
		if tree.Has("dateFieldFormat") {
			result["dateFieldFormat"] = p.DateFormat
		}
		if tree.Has("locale") {
			result["locale"] = p.Locale
		}
		if tree.Has("semanticSearchForArticles") {
			result["semanticSearchForArticles"] = p.SemanticSearchForArticles
		}
		if tree.Has("lastCreatedIssue") {
			result["lastCreatedIssue"] = p.LastCreatedIssue
		}
		if tree.Has("searchContext") {
			result["searchContext"] = p.SearchContext
		}
		if tree.Has("helpdeskContext") {
			result["helpdeskContext"] = p.HelpdeskContext
		}
	}
	result["$type"] = "GeneralUserProfile"
	return result
}

// GetQuestionnaireProfile يعيد إعدادات الاستبيانات للمستخدم.
func getQuestionnaireProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetQuestionnaireProfile", "user context missing"))
		return
	}
	qp, err := c.App.GetQuestionnaireProfile(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, qp)
}

// GetRecentIssues يعيد قائمة المشاكل المشاهدة مؤخراً للمستخدم (مطابق لـ request67.txt).
func getRecentIssues(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetRecentIssues", "user context missing"))
		return
	}
	top, skip := c.Params.Top, c.Params.Skip
	issues, err := c.App.GetRecentIssues(c.AppContext, c.UserID(), top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, issues)
}

// GetRecentArticles يعيد قائمة المقالات المشاهدة مؤخراً للمستخدم (مطابق لـ request68.txt).
func getRecentArticles(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetRecentArticles", "user context missing"))
		return
	}
	top, skip := c.Params.Top, c.Params.Skip
	articles, err := c.App.GetRecentArticles(c.AppContext, c.UserID(), top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, articles)
}

// GetHubMe يعيد ملف المستخدم الحالي في خدمة Hub (مطابق لـ request38.txt).
func getHubMe(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetHubMe", "user context missing"))
		return
	}
	hubUser, err := c.App.GetHubCurrentUser(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, hubUser)
}

// GetInboxFolders يعيد مجلدات صندوق الوارد للمستخدم مع احترام معامل fields (مطابق لـ request8.txt).
func getInboxFolders(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetInboxFolders", "user context missing"))
		return
	}
	fieldTree := c.FieldsTree
	folders, err := c.App.GetInboxFolders(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	result := make([]map[string]any, 0, len(folders))
	for _, f := range folders {
		result = append(result, inboxFolderToMap(f, fieldTree))
	}
	writeModel(w, result)
}

func inboxFolderToMap(f *model.InboxFolder, tree *fields.FieldTree) map[string]any {
	result := make(map[string]any)
	if f == nil {
		f = &model.InboxFolder{}
	}
	if tree == nil || tree.IsEmpty() {
		result["id"] = f.ID
		result["lastNotified"] = f.LastNotified
		result["lastSeen"] = f.LastSeen
		result["enabled"] = f.Enabled
	} else {
		if tree.Has("id") {
			result["id"] = f.ID
		}
		if tree.Has("lastNotified") {
			result["lastNotified"] = f.LastNotified
		}
		if tree.Has("lastSeen") {
			result["lastSeen"] = f.LastSeen
		}
		if tree.Has("enabled") {
			result["enabled"] = f.Enabled
		}
	}
	result["$type"] = "InboxFolder"
	return result
}
