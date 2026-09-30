package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// userIDKey مفتاح تخزين هوية المستخدم الحالي في الـ context.
type userIDKey struct{}

// userRolesKey مفتاح تخزين أدوار المستخدم الحالي في الـ context.
type userRolesKey struct{}

// CurrentUserID يقرأ معرّف المستخدم الحالي من الـ context.
func CurrentUserID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(userIDKey{}).(string)
	return id, ok
}

// withUserID يضع معرّف المستخدم في الـ context.
func withUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// CurrentUserRoles يقرأ أدوار المستخدم الحالي من الـ context.
func CurrentUserRoles(r *http.Request) ([]string, bool) {
	roles, ok := r.Context().Value(userRolesKey{}).([]string)
	return roles, ok
}

// withUserRoles يضع أدوار المستخدم في الـ context.
func withUserRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, userRolesKey{}, append([]string(nil), roles...))
}

// Context يحمل سياق الطلب الموحد لكل دالة معالجة، مستوحى من Mattermost web.Context.
type Context struct {
	App        *app.YouTrackApp
	AppContext request.CTX
	FieldsTree *fields.FieldTree
	Params     *Params
	Err        *model.AppError
}

// ContextFromRequest ينشئ Context مناسباً من الطلب الحالي (للتوافق مع الاستدعاء المباشر).
func ContextFromRequest(a *app.YouTrackApp, r *http.Request) *Context {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.NewString()
	}

	var appCtx request.CTX = request.NewContext(
		r.Context(),
		reqID,
		r.RemoteAddr,
		r.URL.Path,
		r.UserAgent(),
	)

	userID, ok := CurrentUserID(r)
	if ok && userID != "" {
		appCtx = appCtx.WithUserID(userID)
	}

	// الأدوار تُنقل أيضاً إلى سياق التطبيق حتى تتمكن طبقة app من تقييم
	// الصلاحيات الإدرارية (الوصول غير المقيّد بالمشاريع) من داخل دوالها.
	if roles, ok := CurrentUserRoles(r); ok && len(roles) > 0 {
		appCtx = appCtx.WithUserRoles(roles)
	}

	fieldsStr := r.URL.Query().Get("fields")
	var tree *fields.FieldTree
	if fieldsStr != "" {
		tree = fields.Parse(fieldsStr)
	}

	return &Context{
		App:        a,
		AppContext: appCtx,
		FieldsTree: tree,
		Params:     ParamsFromRequest(r),
	}
}

// SessionRequired يتأكد من وجود مستخدم مصادق عليه في السياق.
func (c *Context) SessionRequired() {
	if c.AppContext == nil || c.AppContext.UserID() == "" {
		c.Err = model.NewUnauthorizedError("Context.SessionRequired", "Authentication required")
	}
}

// SetError يضبط خطأ السياق.
func (c *Context) SetError(err *model.AppError) {
	c.Err = err
}

// UserID يعيد معرّف المستخدم الحالي من سياق الطلب.
func (c *Context) UserID() string {
	if c.AppContext != nil {
		return c.AppContext.UserID()
	}
	return ""
}
