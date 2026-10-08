package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes returned in API responses.
const (
	CodeInvalidInput   = "INVALID_INPUT"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeConflict       = "CONFLICT"
	CodeRateLimited    = "RATE_LIMITED"
	CodeInternal       = "INTERNAL_ERROR"
	CodeUnavailable    = "SERVICE_UNAVAILABLE"
	CodeNotImplemented = "NOT_IMPLEMENTED"
)

// APIError is the canonical error type returned by handlers.
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Status  int               `json:"-"`
	Err     error             `json:"-"`
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *APIError) Unwrap() error { return e.Err }

func NewError(status int, code, message string) *APIError {
	return &APIError{Code: code, Message: message, Status: status}
}

func InvalidInput(message string) *APIError {
	return NewError(http.StatusBadRequest, CodeInvalidInput, message)
}

func Unauthorized(message string) *APIError {
	return NewError(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *APIError {
	return NewError(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *APIError {
	return NewError(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *APIError {
	return NewError(http.StatusConflict, CodeConflict, message)
}

func RateLimited(message string) *APIError {
	return NewError(http.StatusTooManyRequests, CodeRateLimited, message)
}

func Internal(message string, err error) *APIError {
	return &APIError{Code: CodeInternal, Message: message, Status: http.StatusInternalServerError, Err: err}
}

func Unavailable(message string) *APIError {
	return NewError(http.StatusServiceUnavailable, CodeUnavailable, message)
}

// AsAPIError converts an arbitrary error into an *APIError.
func AsAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return Internal("An unexpected error occurred", err)
}
