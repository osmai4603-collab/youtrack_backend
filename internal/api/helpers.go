package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"youtrack_backend/internal/model"
)

var (
	errInternal       = errors.New("Internal Server Error")
	errInvalidSigning = errors.New("unexpected signing method")
)

// writeJSON يكتب استجابة JSON مع كود الحالة.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeModel يكتب استجابة JSON بحالة 200.
func writeModel(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

// writeError يكتب رسالة خطأ بصيغة JSON.
func writeError(w http.ResponseWriter, err error) {
	if ce, ok := err.(*model.CustomError); ok {
		writeJSON(w, ce.Status, map[string]any{
			"error": ce.Message,
			"type":  ce.Type,
			"param": ce.Param,
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
}

// parsePagination يستخرج معاملات $top و $skip من الرابط.
func parsePagination(r *http.Request) (top int, skip int) {
	topStr := r.URL.Query().Get("$top")
	skipStr := r.URL.Query().Get("$skip")
	top = 50
	skip = 0
	if topStr != "" {
		if val, err := strconv.Atoi(topStr); err == nil {
			top = val
		}
	}
	if skipStr != "" {
		if val, err := strconv.Atoi(skipStr); err == nil {
			skip = val
		}
	}
	return top, skip
}

// requestIDContextKey مفتاح معرّف الطلب في الـ context.
type requestIDContextKey struct{}
