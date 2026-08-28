package handler

import (
	"encoding/json"
	"net/http"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/middleware"
	"youtrack_backend/internal/usecase"
	"youtrack_backend/pkg/fields"
)

type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUC: userUC,
	}
}

// GetCurrentUser يعالج طلبات /api/users/me، ويعيد مخطط المستخدم الحالي
// (CurrentUser) المطابق للبنية المرجعية في request1.txt.
// يبدأ بتحليل جميع معاملات الاستعلام (query parameters): fields و $top،
// ثم يحدد حقول المستخدم المطلوبة من شجرة fields ويمررها إلى طبقة البيانات
// ليجلب من قاعدة البيانات الأقسام المطلوبة فقط بدلاً من جلب كامل الإعدادات.
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: user not found in context"})
		return
	}

	// 1) تحليل جميع معاملات الاستعلام أولاً
	top := getQueryInt(r, "$top", -1)
	tree := getFields(r)

	// $top = 0 يعني عدم إرجاع أي شيء — يمكن الاختصار قبل الاتصال بقاعدة البيانات
	if top == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
		return
	}

	// 2) تحديد حقول المستخدم المطلوبة (selection) من شجرة fields لتمريرها لقاعدة البيانات
	sel := buildUserFieldSelect(tree)

	user, err := h.userUC.GetCurrentUserDetail(userID, sel)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	// 3) تحويل كائن CurrentUser (struct) إلى map[string]any ليتمكن منتقي الحقول من تصفيته
	data := structToMap(user)
	if tree != nil {
		data = fields.Filter(data, tree)
	}

	// 4) $top=-1 (أو موجب) مع كائن وحيد: لا اقتصار إضافي
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

// buildUserFieldSelect يُحوّل شجرة fields إلى domain.UserFieldSelect يحدد الأقسام
// المطلوبة، بحيث تُنفَّذ استعلامات قاعدة البيانات للأقسام المطلوبة فقط.
// يُعيد nil عندما لا يوجد معامل fields (أي جلب كل شيء).
func buildUserFieldSelect(tree *fields.FieldNode) *domain.UserFieldSelect {
	if tree == nil {
		return nil
	}

	sel := &domain.UserFieldSelect{}
	sel.Profiles = fields.HasField(tree, "profiles")
	sel.Widgets = fields.HasField(tree, "widgets")
	sel.FeatureFlags = fields.HasField(tree, "featureFlags")
	sel.IssueRelatedGroup = fields.HasField(tree, "issueRelatedGroup")

	profilesTree := fields.GetChild(tree, "profiles")
	if sel.Profiles && (profilesTree == nil || !profilesTree.HasChildren()) {
		// profiles مطلوب ككل (بدون حقول فرعية) → كل الأقسام الفرعية
		sel.General = true
		sel.Articles = true
		sel.TimeTracking = true
		sel.Tips = true
		sel.Appearance = true
		sel.IssuesList = true
		sel.Helpdesk = true
		sel.AI = true
		sel.Notifications = true
	} else if profilesTree != nil {
		sel.General = fields.HasField(profilesTree, "general")
		sel.Articles = fields.HasField(profilesTree, "articles")
		sel.TimeTracking = fields.HasField(profilesTree, "timetracking")
		sel.Tips = fields.HasField(profilesTree, "tips")
		sel.Appearance = fields.HasField(profilesTree, "appearance")
		sel.IssuesList = fields.HasField(profilesTree, "issuesList")
		sel.Helpdesk = fields.HasField(profilesTree, "helpdesk")
		sel.AI = fields.HasField(profilesTree, "ai")
		sel.Notifications = fields.HasField(profilesTree, "notifications")
	}

	return sel
}

// structToMap يحول أي struct إلى map[string]any عبر JSON round-trip
// ليتمكن منتقي الحقول (fields.Filter) من التصفية على مستوى المفاتيح.
func structToMap(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return v
	}
	return m
}

// GetGrazieProfile يعالج طلبات /api/users/me/profiles/grazie
func (h *UserHandler) GetGrazieProfile(w http.ResponseWriter, r *http.Request) {
	grazie := h.userUC.GetGrazieProfile()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(grazie)
}
