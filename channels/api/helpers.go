package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"youtrack_backend/channels/model"
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

// writeError يكتب رسالة خطأ بصيغة JSON متوافقة مع AppError.
func writeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if appErr, ok := err.(*model.AppError); ok {
		if appErr != nil {
			writeJSON(w, appErr.StatusCode, appErr)
			return
		}
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
}

// nullOr يحوّل سلسلة فارغة إلى nil ليعرضها JSON كـ null كما في YouTrack.
func nullOr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullOrPtr يحوّل مؤشر سلسلة إلى nil إذا كان فارغًا.
func nullOrPtr(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
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
