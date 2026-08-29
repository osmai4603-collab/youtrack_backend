package api

import (
	"context"
	"net/http"
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
