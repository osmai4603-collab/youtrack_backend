package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	platformctx "youtrack_backend/internal/platform/context"
)

// middlewareOptions holds the configurable behavior of the HTTP logging
// middleware.
type middlewareOptions struct {
	// suppressProbePaths contains exact URL paths whose successful (< 400)
	// completion logs are suppressed. Failing probes (>= 400) are never
	// suppressed so outages always surface.
	suppressProbePaths map[string]struct{}
}

// Option configures HTTPLoggingMiddleware behavior.
type Option func(*middlewareOptions)

// WithProbeSuppression silences completion logs for successful requests to the
// given paths. It is intended for health-check endpoints such as /livez and
// /readyz that are polled continuously; probe failures (>= 400) are always
// logged at WARN/ERROR so problems remain visible.
func WithProbeSuppression(paths ...string) Option {
	return func(o *middlewareOptions) {
		for _, p := range paths {
			if p = strings.TrimSpace(p); p != "" {
				o.suppressProbePaths[p] = struct{}{}
			}
		}
	}
}

// responseRecorder intercepts the response status code and the number of bytes
// written so they can be recorded in the request completion log after handling.
type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
	body   bytes.Buffer
}

const errorLogBodyLimit = 8 << 10
const errorLogMessageLimit = 512

// WriteHeader captures the status code and prevents double writes.
func (w *responseRecorder) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write captures the number of bytes sent to the client.
func (w *responseRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	if w.status >= http.StatusBadRequest && w.body.Len() < errorLogBodyLimit {
		remaining := errorLogBodyLimit - w.body.Len()
		if len(b) > remaining {
			b = b[:remaining]
		}
		_, _ = w.body.Write(b)
	}
	return n, err
}

// HTTPLoggingMiddleware creates an HTTP middleware that:
//
//  1. Extracts request_id and client_ip from the platform package context
//     (internal/platform/context), with the X-Request-ID header as a final fallback.
//  2. Builds a scoped logger carrying the request identifier and destination, and
//     injects it into the request context via WithContext.
//  3. Routes the log level according to the status code (5xx -> ERROR, 4xx -> WARN,
//     otherwise INFO), measures the duration, and records "http request completed"
//     via LogAttrs (a zero-allocation path for hot-path data).
//
// When WithProbeSuppression is passed, completion logs for the specified health
// check successes are ignored to avoid log bloat, while their failures always
// remain visible.
//
// This middleware must be mounted strictly after the platform package's context
// Middleware to guarantee request_id is available in the context before reaching
// this point.
func HTTPLoggingMiddleware(base *slog.Logger, opts ...Option) func(http.Handler) http.Handler {
	if base == nil {
		base = slog.Default()
	}

	options := middlewareOptions{suppressProbePaths: make(map[string]struct{})}
	for _, opt := range opts {
		opt(&options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Build a scoped logger carrying the current request attributes
			reqLogger := base.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			if meta, ok := platformctx.MetadataFromContext(r.Context()); ok {
				if meta.RequestID != "" {
					reqLogger = reqLogger.With(slog.String("request_id", meta.RequestID))
				}
				if meta.ClientIP != "" {
					reqLogger = reqLogger.With(slog.String("client_ip", meta.ClientIP))
				}
			} else if reqID := r.Header.Get(platformctx.HeaderRequestID); reqID != "" {
				reqLogger = reqLogger.With(slog.String("request_id", reqID))
			}

			ctx := WithContext(r.Context(), reqLogger)
			rec := &responseRecorder{ResponseWriter: w}

			// Execute the handler chain
			next.ServeHTTP(rec, r.WithContext(ctx))

			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}

			// Suppress completion logs for the specified healthy probes when requested
			if _, suppress := options.suppressProbePaths[r.URL.Path]; suppress && status < http.StatusBadRequest {
				return
			}

			// Select the log level based on the status code
			level := slog.LevelInfo
			switch {
			case status >= 500:
				level = slog.LevelError
			case status >= 400:
				level = slog.LevelWarn
			}

			// Log request completion after it finishes, using LogAttrs to avoid memory allocations
			attrs := []slog.Attr{
				slog.Int("status", status),
				slog.Duration("duration", time.Since(start)),
				slog.Int("bytes", rec.bytes),
			}
			if status >= http.StatusBadRequest {
				code, message := responseErrorDetails(status, rec.Header().Get("Content-Type"), rec.body.String())
				if code != "" {
					attrs = append(attrs, slog.String("error_code", code))
				}
				attrs = append(attrs, slog.String("error_message", message))
			}
			reqLogger.LogAttrs(ctx, level, "http request completed",
				attrs...,
			)
		})
	}
}

func responseErrorDetails(status int, contentType, body string) (string, string) {
	body = strings.TrimSpace(body)
	if strings.Contains(strings.ToLower(contentType), "json") {
		var payload map[string]any
		if json.Unmarshal([]byte(body), &payload) == nil {
			code, message := findErrorDetails(payload)
			if message != "" {
				return code, limitLogMessage(message)
			}
		}
	}

	if strings.HasPrefix(strings.ToLower(contentType), "text/plain") && body != "" {
		return "", limitLogMessage(body)
	}
	return "", "HTTP " + http.StatusText(status)
}

func findErrorDetails(value map[string]any) (string, string) {
	code, _ := value["code"].(string)
	message, _ := value["message"].(string)
	if message != "" {
		return code, message
	}
	switch nested := value["error"].(type) {
	case map[string]any:
		return findErrorDetails(nested)
	case string:
		return code, nested
	}
	return code, ""
}

func limitLogMessage(message string) string {
	message = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' {
			return ' '
		}
		return r
	}, strings.TrimSpace(message))
	runes := []rune(message)
	if len(runes) > errorLogMessageLimit {
		message = string(runes[:errorLogMessageLimit]) + "..."
	}
	return message
}
