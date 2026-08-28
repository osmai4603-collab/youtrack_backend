package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// CORS ينشئ Middleware لتمكين الطلبات من تطبيقات الـ Frontend (Web / Flutter / Desktop)
func CORS() func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"*", // يمكن تحديد نطاقات محددة في بيئة الإنتاج
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Link",
			"Content-Length",
			"Content-Range",
		},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes cache for preflight
	})

	return c.Handler
}
