package api

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/cors"
)

// CORS يعيد middleware للسماح بالطلبات من تطبيقات الواجهة.
func CORS() func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link", "Content-Length", "Content-Range"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	return c.Handler
}

// RequestID يضيف معرّفًا فريدًا لكل طلب.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger يسجل كل طلب.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		reqID, _ := r.Context().Value(requestIDContextKey{}).(string)
		log.Printf("[%s] %s %s (%s)", reqID, r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// Recoverer يلتقط الـ panic ويمنع انهيار الخادم.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC RECOVERED] %v", rec)
				writeError(w, errInternal)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// VersionedAPI accepts /api/v1 and /api/v2 while retaining the existing /api routes.
func VersionedAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for _, version := range []string{"/api/v1", "/api/v2"} {
			if path == version || strings.HasPrefix(path, version+"/") {
				r.URL.Path = "/api" + strings.TrimPrefix(path, version)
				w.Header().Set("X-API-Version", version[5:])
				defer func() { r.URL.Path = path }()
				break
			}
		}
		next.ServeHTTP(w, r)
	})
}

// JWTAuth يتحقق من صحة رمز Bearer ويضع هوية المستخدم في الـ context.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			parts := strings.Fields(auth)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Authorization header required"})
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errInvalidSigning
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Invalid or expired token"})
				return
			}

			userID, _ := claims["sub"].(string)
			if userID == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Invalid token claims"})
				return
			}

			ctx := withUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
