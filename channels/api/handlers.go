package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// HandlerFunc يمثل توقيع دوال معالجة الطلبات في طبقة الـ API.
type HandlerFunc func(*Context, http.ResponseWriter, *http.Request)

// Handler هو المغلف التنفيذي لخط أنابيب معالجة طلبات HTTP.
type Handler struct {
	App               *app.YouTrackApp
	HandleFunc        HandlerFunc
	RequireSession    bool
	RequirePermission string
	TrustRequester    bool
}

// ServeHTTP ينفذ خط الأنابيب: التحقق، إعداد السياق، الاستدعاء، ومعالجة الأخطاء.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.NewString()
	}
	w.Header().Set("X-Request-ID", reqID)

	appCtx := request.NewContext(
		r.Context(),
		reqID,
		r.RemoteAddr,
		r.URL.Path,
		r.UserAgent(),
	)

	fieldsStr := r.URL.Query().Get("fields")
	var tree *fields.FieldTree
	if fieldsStr != "" {
		tree = fields.Parse(fieldsStr)
	}

	c := &Context{
		App:        h.App,
		AppContext: appCtx,
		FieldsTree: tree,
		Params:     ParamsFromRequest(r),
	}

	// التحقق من المصادقة إذا كان المسار يتطلب جلسة
	if h.RequireSession {
		auth := r.Header.Get("Authorization")
		parts := strings.Fields(auth)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.Err = model.NewUnauthorizedError("Handler", "Authorization header required")
		} else {
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				jwtSecret := h.App.JWTSecret()
				if jwtSecret == "" {
					jwtSecret = "secret"
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				c.Err = model.NewUnauthorizedError("Handler", "Invalid or expired token")
			} else {
				userID, _ := claims["sub"].(string)
				if userID == "" {
					c.Err = model.NewUnauthorizedError("Handler", "Invalid token claims")
				} else {
					var roles []string
					if rawRoles, ok := claims["roles"]; ok {
						switch v := rawRoles.(type) {
						case []any:
							for _, item := range v {
								if s, ok := item.(string); ok {
									roles = append(roles, s)
								}
							}
						case []string:
							roles = v
						case string:
							roles = []string{v}
						}
					}
					c.AppContext = c.AppContext.WithUserID(userID).WithUserRoles(roles).WithSessionToken(parts[1])
					// نقل الهوية والأدوار إلى سياق الطلب حتى تصل إلى المعالجات التي
					// تعيد بناء سياق التطبيق عبر ContextFromRequest.
					r = r.WithContext(withUserRoles(withUserID(r.Context(), userID), roles))
				}
			}
		}
	}

	if c.Err == nil && h.RequirePermission != "" {
		if !h.App.SessionHasPermission(c.AppContext, h.RequirePermission) {
			c.Err = model.NewForbiddenError("Handler", "Permission denied")
		}
	}

	if c.Err == nil {
		h.HandleFunc(c, w, r)
	}

	if c.Err != nil {
		h.handleError(c, w, r)
	}
}

func (h Handler) handleError(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.Err.StatusCode == 0 {
		c.Err.StatusCode = http.StatusInternalServerError
	}
	c.Err.RequestId = c.AppContext.RequestId()

	log.Printf("[%s] ERROR %s %s: status=%d msg=%s details=%s",
		c.Err.RequestId, r.Method, r.URL.Path, c.Err.StatusCode, c.Err.Message, c.Err.DetailedError)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c.Err.StatusCode)
	w.Write([]byte(c.Err.ToJSON()))
}
