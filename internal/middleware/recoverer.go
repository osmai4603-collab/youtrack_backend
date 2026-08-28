package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recoverer يلتقط أي panic مفاجئ ويمنع انهيار السيرفر مع إرجاع 500 Internal Server Error
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack trace:\n%s", rvr, debug.Stack())
				http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
