package api

import (
	"net/http"

	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/model"
)

// GetUserRequest يمثّل مخطط طلب جلب مستخدم (معرّف مستخدم من المسار).
type GetUserRequest struct {
	UserID string // من المسار {id}
}

// UserHandler يعالج طلبات المستخدمين.
type UserHandler struct {
	app *app.App
}

func NewUserHandler(a *app.App) *UserHandler {
	return &UserHandler{app: a}
}

// GetMe يعيد المستخدم الحالي مع كامل مرفقاته وإعداداته مباشرة في جذر الـ JSON (مطابق لـ request1.txt).
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}

	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	user, err := h.app.GetCurrentUser(r.Context(), userID, fieldTree)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, user)
}

// List يعيد قائمة المستخدمين بصيغة مصفوفة JSON مباشرة (مطابق لـ request32.txt).
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.app.GetAllUsers(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, users)
}

// GetByID يعيد مستخدمًا بمعرّفه مباشرة في جذر الـ JSON (مطابق لـ request33.txt).
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	req := GetUserRequest{UserID: r.PathValue("id")}
	if req.UserID == "me" {
		h.GetMe(w, r)
		return
	}
	user, err := h.app.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, user)
}

// GetGrazieProfile يعيد إعدادات الذكاء الاصطناعي والتدقيق اللغوي Grazie (مطابق لـ request4.txt).
func (h *UserHandler) GetGrazieProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	grazie, err := h.app.GetGrazieProfile(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, grazie)
}

// GetGeneralProfile يعيد الإعدادات العامة لملف المستخدم (مطابق لـ request29.txt).
func (h *UserHandler) GetGeneralProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	gen, err := h.app.GetGeneralProfile(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	if fieldTree == nil || fieldTree.IsEmpty() {
		writeModel(w, gen)
		return
	}
	writeModel(w, generalProfileToMap(gen, fieldTree))
}

// generalProfileToMap يحوّل GeneralUserProfile إلى خريطة تحترم معامل fields
// وتضمن $type دائماً كما في استجابة YouTrack الأصلية.
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
			dff := tree.Child("dateFieldFormat")
			if dff != nil && !dff.IsEmpty() {
				dfMap := make(map[string]any)
				if p.DateFormat == nil {
					p.DateFormat = &model.DateFormatDescriptor{}
				}
				if dff.Has("pattern") {
					dfMap["pattern"] = p.DateFormat.Pattern
				}
				if dff.Has("datePattern") {
					dfMap["datePattern"] = p.DateFormat.DatePattern
				}
				if dff.Has("$type") {
					dfMap["$type"] = p.DateFormat.Type
				}
				result["dateFieldFormat"] = dfMap
			} else {
				result["dateFieldFormat"] = p.DateFormat
			}
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

// GetQuestionnaireProfile يعيد إعدادات واستطلاعات الرأي للمستخدم (مطابق لـ request66.txt).
func (h *UserHandler) GetQuestionnaireProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	qp, err := h.app.GetQuestionnaireProfile(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, qp)
}

// GetRecentIssues يعيد قائمة المشاكل المشاهدة مؤخراً للمستخدم (مطابق لـ request67.txt).
func (h *UserHandler) GetRecentIssues(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	top, skip := parsePagination(r)
	issues, err := h.app.GetRecentIssues(r.Context(), userID, top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, issues)
}

// GetRecentArticles يعيد قائمة المقالات المشاهدة مؤخراً للمستخدم (مطابق لـ request68.txt).
func (h *UserHandler) GetRecentArticles(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	top, skip := parsePagination(r)
	articles, err := h.app.GetRecentArticles(r.Context(), userID, top, skip)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, articles)
}

// GetHubMe يعيد ملف المستخدم الحالي في خدمة Hub (مطابق لـ request38.txt).
func (h *UserHandler) GetHubMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	hubUser, err := h.app.GetHubCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeModel(w, hubUser)
}

// GetInboxFolders يعيد مجلدات صندوق الوارد للمستخدم مع احترام معامل fields (مطابق لـ request8.txt).
func (h *UserHandler) GetInboxFolders(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r)
	if !ok {
		writeError(w, model.Unauthorized("user context missing"))
		return
	}
	fieldTree := fields.Parse(r.URL.Query().Get("fields"))
	folders, err := h.app.GetInboxFolders(r.Context(), userID)
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

// inboxFolderToMap يحوّل مجلد صندوق وارد إلى خريطة تحترم معامل fields
// وتضمّن $type دائماً كما في استجابة YouTrack الأصلية.
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
