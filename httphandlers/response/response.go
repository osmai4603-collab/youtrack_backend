package response

import (
	"encoding/json"
	"net/http"
	"youtrack_backend/internal/platform/errors"
)

type Envelope struct {
	Success bool          `json:"success"`
	Data    any           `json:"data,omitempty"`
	Meta    any           `json:"meta,omitempty"`
	Error   *ErrorPayload `json:"error,omitempty"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"error,omitmpty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

func Error(w http.ResponseWriter, args ...any) {
	var req *http.Request
	var err error
	for _, arg := range args {
		switch v := arg.(type) {
		case *http.Request:
			req = v
		case error:
			err = v
		}
	}
	if err == nil {
		err = errors.Internal("error", "internal server error", nil)
	}
	status := errors.HTTPStatusCode(err)
	translate := func(msg string) string {
		if req == nil {
			return msg
		}
		return msg
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var appErr *errors.Error
	if ok := isAppError(err, &appErr); ok {
		_ = json.NewEncoder(w).Encode(Envelope{
			Success: false,
			Error: &ErrorPayload{
				Code:    appErr.Code,
				Message: translate(appErr.Message),
				Details: appErr.Err,
			},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Error: &ErrorPayload{
			Code:    errors.CodeInternal,
			Message: translate("internal server error"),
		},
	})
}

func isAppError(err error, target **errors.Error) bool {
	var appErr *errors.Error
	if errorsAs(err, &appErr) {
		*target = appErr
		return true
	}
	return false
}

var errorsAs = func(err error, target any) bool {
	return platformerrorsAs(err, target)
}

func platformerrorsAs(err error, target any) bool {
	type unwrapper interface {
		Unwrap() error
	}

	for err != nil {
		if t, ok := target.(**errors.Error); ok {
			if e, ok := err.(*errors.Error); ok {
				*t = e
				return true
			}
		}
		u, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}
