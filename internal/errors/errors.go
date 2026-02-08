package errors

import (
	stderrors "errors"
	"fmt"
)

// Sentinel errors
var (
	ErrUserNotFound      = stderrors.New("user not found")
	ErrUserAlreadyExists = stderrors.New("user already exists")
	ErrInvalidCredentials = stderrors.New("invalid credentials")
	ErrInvalidToken      = stderrors.New("invalid token")
	ErrTokenExpired      = stderrors.New("token expired")
	ErrKVOperationFailed = stderrors.New("KV operation failed")
	ErrInvalidInput      = stderrors.New("invalid input")
)

// Is wraps errors.Is
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As wraps errors.As
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

// AppError represents an application error with context
type AppError struct {
	Op      string // Operation that failed
	Err     error  // Underlying error
	Message string // User-friendly message
	Code    int    // HTTP status code or error code
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(op string, err error, message string, code int) *AppError {
	return &AppError{
		Op:      op,
		Err:     err,
		Message: message,
		Code:    code,
	}
}
