package apperror

import (
	"errors"
	"fmt"
)

// Code represents an application-level error classification.
type Code string

const (
	CodeValidation          Code = "VALIDATION"
	CodeNotFound            Code = "NOT_FOUND"
	CodeConflict            Code = "CONFLICT"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeForbidden           Code = "FORBIDDEN"
	CodeInternal            Code = "INTERNAL"
	CodeUnprocessableEntity Code = "UNPROCESSABLE_ENTITY"
)

// AppError is the shared application error type. Services return these;
// the transport layer (HTTP, gRPC) maps them to protocol-specific status codes.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"` // unwrappable root cause, never serialized to clients
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// Wrap attaches a root-cause error for structured logging and errors.Unwrap.
func (e *AppError) Wrap(err error) *AppError {
	e.Err = err
	return e
}

// --- Constructors --------------------------------------------------------

// Validation creates a 400-class error for bad input, missing fields, etc.
func Validation(msg string, err ...error) *AppError {
	appErr := &AppError{Code: CodeValidation, Message: msg}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

// NotFound creates a 404-class error for missing resources.
func NotFound(msg string, err ...error) *AppError {
	appErr := &AppError{Code: CodeNotFound, Message: msg}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

// Conflict creates a 409-class error for duplicate resources, constraint violations.
func Conflict(msg string, err ...error) *AppError {
	appErr := &AppError{Code: CodeConflict, Message: msg}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

// Unauthorized creates a 401-class error for missing/invalid credentials.
func Unauthorized(msg string, err ...error) *AppError {
	appErr := &AppError{Code: CodeUnauthorized, Message: msg}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

// Forbidden creates a 403-class error for insufficient permissions.
func Forbidden(msg string, err ...error) *AppError {
	appErr := &AppError{Code: CodeForbidden, Message: msg}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

// Internal creates a 500-class error for unexpected failures.
func Internal(msg string, err ...error) *AppError {
	appErr := &AppError{Code: CodeInternal, Message: msg}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

func UnprocessableEntity(msg, details string, err ...error) *AppError {
	appErr := &AppError{Code: CodeUnprocessableEntity, Message: msg, Details: details}
	if len(err) > 0 {
		appErr.Err = err[0]
	}
	return appErr
}

// --- Inspection helpers --------------------------------------------------

// IsCode checks whether err (or any wrapped error) is an AppError with the given code.
func IsCode(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// AsAppError extracts an *AppError from an error chain, if present.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
