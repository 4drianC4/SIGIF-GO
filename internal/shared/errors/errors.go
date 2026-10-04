package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeValidation         ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeConflict           ErrorCode = "CONFLICT"
	CodeBadRequest         ErrorCode = "BAD_REQUEST"
	CodeTooManyRequests    ErrorCode = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

var (
	ErrInternal           = New(CodeInternal, "internal server error", http.StatusInternalServerError)
	ErrNotFound           = New(CodeNotFound, "resource not found", http.StatusNotFound)
	ErrValidation         = New(CodeValidation, "validation failed", http.StatusBadRequest)
	ErrUnauthorized       = New(CodeUnauthorized, "unauthorized", http.StatusUnauthorized)
	ErrForbidden          = New(CodeForbidden, "forbidden", http.StatusForbidden)
	ErrConflict           = New(CodeConflict, "resource conflict", http.StatusConflict)
	ErrBadRequest         = New(CodeBadRequest, "bad request", http.StatusBadRequest)
	ErrTooManyRequests    = New(CodeTooManyRequests, "too many requests", http.StatusTooManyRequests)
	ErrServiceUnavailable = New(CodeServiceUnavailable, "service unavailable", http.StatusServiceUnavailable)
)

type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	StatusCode int       `json:"-"`
	Details    any       `json:"details,omitempty"`
	Cause      error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getStatusCode(code),
		Cause:      err,
	}
}

func Wrapf(err error, code ErrorCode, format string, args ...any) *AppError {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

func Is(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func getStatusCode(code ErrorCode) int {
	switch code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeConflict:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternal
}

func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}
