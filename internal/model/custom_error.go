package model

import (
	"fmt"
	"net/http"
)

// CustomError يمثّل خطأً موحّدًا في التطبيق، يحمل رسالة وكود HTTP اختياري.
type CustomError struct {
	Message string `json:"error"`
	Status  int    `json:"-"`
	Type    string `json:"type,omitempty"`
	Param   string `json:"param,omitempty"`
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewError(status int, format string, args ...any) *CustomError {
	return &CustomError{
		Message: fmt.Sprintf(format, args...),
		Status:  status,
	}
}

func NotFound(format string, args ...any) *CustomError {
	return NewError(http.StatusNotFound, format, args...)
}

func BadRequest(format string, args ...any) *CustomError {
	return NewError(http.StatusBadRequest, format, args...)
}

func Unauthorized(format string, args ...any) *CustomError {
	return NewError(http.StatusUnauthorized, format, args...)
}

func Internal(format string, args ...any) *CustomError {
	return NewError(http.StatusInternalServerError, format, args...)
}
