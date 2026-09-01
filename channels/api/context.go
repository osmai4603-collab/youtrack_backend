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

// CurrentUserID يقرأ معرّف المستخدم الحالي من الـ context.
func CurrentUserID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(userIDKey{}).(string)
	return id, ok
}

// withUserID يضع معرّف المستخدم في الـ context.
func withUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// Context يحمل سياق الطلب الموحد لكل دالة معالجة، مستوحى من Mattermost web.Context.
type Context struct {
	App        *app.YouTrackApp
	AppContext request.CTX
	FieldsTree *fields.FieldTree
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

	fieldsStr := r.URL.Query().Get("fields")
	var tree *fields.FieldTree
	if fieldsStr != "" {
		tree = fields.Parse(fieldsStr)
	}

	return &Context{
		App:        a,
		AppContext: appCtx,
		FieldsTree: tree,
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
