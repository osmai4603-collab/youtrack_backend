package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriterRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger يسجل كل طلب HTTP مع زمن الاستجابة، كود الحالة، و RequestID
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseWriterRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		reqID, _ := r.Context().Value(RequestIDKey).(string)
		if reqID != "" {
			log.Printf("[%s] [%s] %d %s (%s)", reqID, r.Method, recorder.statusCode, r.URL.Path, time.Since(start))
		} else {
			log.Printf("[%s] %d %s (%s)", r.Method, recorder.statusCode, r.URL.Path, time.Since(start))
		}
	})
}

