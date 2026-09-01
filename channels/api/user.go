package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
)

// GetUserRequest يمثّل مخطط طلب جلب مستخدم (معرّف مستخدم من المسار).
type GetUserRequest struct {
	UserID string // من المسار {id}
}

// UserHandler يعالج طلبات المستخدمين.
type UserHandler struct {
	app *app.YouTrackApp
}

func NewUserHandler(a *app.YouTrackApp) *UserHandler {
	return &UserHandler{app: a}
}

func (api *API) InitUser() {
	handler := NewUserHandler(app.New())

	api.BaseRoutes.Users.Handle("", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.List(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/me", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetMe(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+}", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetByID(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/profiles/grazie", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetGrazieProfile(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/profiles/general", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetGeneralProfile(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/profiles/questionnaire", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetQuestionnaireProfile(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/recentIssues", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetRecentIssues(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/recentArticles", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetRecentArticles(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/hubMe", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetHubMe(w, r)
	})).Methods("GET")

	api.BaseRoutes.Users.Handle("/{id:[A-Za-z0-9_\\-\\.]+|me}/folders", api.APISessionRequired(func(c *Context, w http.ResponseWriter, r *http.Request) {
		handler.GetInboxFolders(w, r)
	})).Methods("GET")
}

// GetMe يعيد المستخدم الحالي مع كامل مرفقاته وإعداداته مباشرة في جذر الـ JSON (مطابق لـ request1.txt).
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetMe", "user context missing"))
		return
	}

	fieldTree := c.FieldsTree
	user, err := h.app.GetCurrentUser(c.AppContext, c.UserID(), fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, user)
}

// List يعيد قائمة المستخدمين بصيغة مصفوفة JSON مباشرة (مطابق لـ request32.txt).
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	users, err := h.app.GetAllUsers(c.AppContext)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, users)
}

// GetByID يعيد مستخدمًا بمعرّفه مباشرة في جذر الـ JSON (مطابق لـ request33.txt).
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	userID := mux.Vars(r)["id"]
	if userID == "" {
		userID = r.PathValue("id")
	}
	if userID == "me" {
		h.GetMe(w, r)
		return
	}
	user, err := h.app.GetUserByID(c.AppContext, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, user)
}

// GetGrazieProfile يعيد إعدادات الذكاء الاصطناعي والتدقيق اللغوي Grazie (مطابق لـ request4.txt).
func (h *UserHandler) GetGrazieProfile(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetGrazieProfile", "user context missing"))
		return
	}
	grazie, err := h.app.GetGrazieProfile(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, grazie)
}

// GetGeneralProfile يعيد الإعدادات العامة لملف المستخدم (مطابق لـ request29.txt).
func (h *UserHandler) GetGeneralProfile(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetGeneralProfile", "user context missing"))
		return
	}
	gen, err := h.app.GetGeneralProfile(c.AppContext, c.UserID())
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
func (h *UserHandler) GetQuestionnaireProfile(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetQuestionnaireProfile", "user context missing"))
		return
	}
	qp, err := h.app.GetQuestionnaireProfile(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, qp)
}

// GetRecentIssues يعيد قائمة المشاكل المشاهدة مؤخراً للمستخدم (مطابق لـ request67.txt).
func (h *UserHandler) GetRecentIssues(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetRecentIssues", "user context missing"))
		return
	}
	top, skip := parsePagination(r)
	issues, err := h.app.GetRecentIssues(c.AppContext, c.UserID(), top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, issues)
}

// GetRecentArticles يعيد قائمة المقالات المشاهدة مؤخراً للمستخدم (مطابق لـ request68.txt).
func (h *UserHandler) GetRecentArticles(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetRecentArticles", "user context missing"))
		return
	}
	top, skip := parsePagination(r)
	articles, err := h.app.GetRecentArticles(c.AppContext, c.UserID(), top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, articles)
}

// GetHubMe يعيد ملف المستخدم الحالي في خدمة Hub (مطابق لـ request38.txt).
func (h *UserHandler) GetHubMe(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetHubMe", "user context missing"))
		return
	}
	hubUser, err := h.app.GetHubCurrentUser(c.AppContext, c.UserID())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, hubUser)
}

// GetInboxFolders يعيد مجلدات صندوق الوارد للمستخدم مع احترام معامل fields (مطابق لـ request8.txt).
func (h *UserHandler) GetInboxFolders(w http.ResponseWriter, r *http.Request) {
	c := ContextFromRequest(h.app, r)
	if c.UserID() == "" {
		writeError(w, model.NewUnauthorizedError("User.GetInboxFolders", "user context missing"))
		return
	}
	fieldTree := c.FieldsTree
	folders, err := h.app.GetInboxFolders(c.AppContext, c.UserID())
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
