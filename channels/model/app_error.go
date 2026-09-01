package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const maxErrorLength = 1024

// AppError يمثل هيكل الأخطاء المعياري والشامل للمنصة، متوافقاً مع معمارية Mattermost.
type AppError struct {
	Id            string         `json:"id"`
	Message       string         `json:"message"`               // رسالة موجهة للمستخدم
	DetailedError string         `json:"detailed_error"`        // تفاصيل الخطأ التقنية للمطورين
	RequestId     string         `json:"request_id,omitempty"`  // معرّف الطلب لربطه بالسجلات
	StatusCode    int            `json:"status_code,omitempty"` // رمز حالة HTTP
	Where         string         `json:"-"`                     // الدالة أو المكان الذي نشأ منه الخطأ (Struct.Func)
	params        map[string]any `json:"-"`
	wrapped       error          `json:"-"`
}

func (er *AppError) Error() string {
	var sb strings.Builder

	if er.Where != "" {
		sb.WriteString(er.Where)
		sb.WriteString(": ")
	}

	if er.Message != "" {
		sb.WriteString(er.Message)
	}

	if er.DetailedError != "" {
		if er.Message != "" {
			sb.WriteString(", ")
		}
		sb.WriteString(er.DetailedError)
	}

	if er.wrapped != nil {
		sb.WriteString(", ")
		sb.WriteString(er.wrapped.Error())
	}

	res := sb.String()
	if len(res) > maxErrorLength {
		res = res[:maxErrorLength] + "..."
	}
	return res
}

func (er *AppError) Unwrap() error {
	return er.wrapped
}

func (er *AppError) Wrap(err error) *AppError {
	er.wrapped = err
	return er
}

func (er *AppError) ToJSON() string {
	detailed := er.DetailedError
	defer func() {
		er.DetailedError = detailed
	}()

	if er.wrapped != nil {
		if er.DetailedError != "" {
			er.DetailedError += ", "
		}
		er.DetailedError += er.wrapped.Error()
	}

	b, _ := json.Marshal(er)
	return string(b)
}

// NewAppError ينشئ كائن AppError جديد مع كافة الحقول المحددة.
func NewAppError(where, id string, params map[string]any, details string, status int) *AppError {
	msg := id
	if details != "" && msg == "" {
		msg = details
	}
	return &AppError{
		Id:            id,
		Message:       msg,
		Where:         where,
		DetailedError: details,
		StatusCode:    status,
		params:        params,
	}
}

// NewNotFoundError ينشئ خطأ 404 Not Found.
func NewNotFoundError(where, details string) *AppError {
	return NewAppError(where, "app.not_found", nil, details, http.StatusNotFound)
}

// NewBadRequestError ينشئ خطأ 400 Bad Request.
func NewBadRequestError(where, details string) *AppError {
	return NewAppError(where, "app.bad_request", nil, details, http.StatusBadRequest)
}

// NewUnauthorizedError ينشئ خطأ 401 Unauthorized.
func NewUnauthorizedError(where, details string) *AppError {
	return NewAppError(where, "app.unauthorized", nil, details, http.StatusUnauthorized)
}

// NewForbiddenError ينشئ خطأ 403 Forbidden.
func NewForbiddenError(where, details string) *AppError {
	return NewAppError(where, "app.forbidden", nil, details, http.StatusForbidden)
}

// NewInternalError ينشئ خطأ 500 Internal Server Error مع إمكانية تغليف خطأ أساسي.
func NewInternalError(where, details string, err error) *AppError {
	appErr := NewAppError(where, "app.internal_error", nil, details, http.StatusInternalServerError)
	if err != nil {
		appErr.Wrap(err)
	}
	return appErr
}

// دوال مساعدة للتوافق المباشر
func NotFound(format string, args ...any) *AppError {
	return NewNotFoundError("", fmt.Sprintf(format, args...))
}

func BadRequest(format string, args ...any) *AppError {
	return NewBadRequestError("", fmt.Sprintf(format, args...))
}

func Unauthorized(format string, args ...any) *AppError {
	return NewUnauthorizedError("", fmt.Sprintf(format, args...))
}

func Forbidden(format string, args ...any) *AppError {
	return NewForbiddenError("", fmt.Sprintf(format, args...))
}

func Internal(format string, args ...any) *AppError {
	return NewInternalError("", fmt.Sprintf(format, args...), nil)
}
