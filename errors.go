package cloudflare_auth_sdk

import (
	"errors"
	"fmt"
)

// Common errors that can be returned by the SDK.
var (
	// Configuration errors
	ErrInvalidConfig         = errors.New("invalid configuration")
	ErrMissingRequiredConfig = errors.New("missing required configuration fields")
	ErrInvalidAuth           = errors.New("invalid authentication credentials")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")

	// Token errors
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenExpired = errors.New("token has expired")

	// Input errors
	ErrInvalidInput = errors.New("invalid input parameters")

	// KV errors
	ErrKVOperationFailed = errors.New("KV operation failed")
)

// AppError represents an application error with additional context.
//
// It wraps an underlying error with operation information, a user-friendly
// message, and an HTTP status code.
type AppError struct {
	Op      string // Operation that failed (e.g., "Client.Register")
	Err     error  // Underlying error
	Message string // User-friendly error message
	Code    int    // HTTP status code or error code
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error for error chain inspection.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError with the given parameters.
//
// Example:
//
//	err := NewAppError("Client.Register", ErrUserAlreadyExists,
//	    "A user with this email already exists", 409)
func NewAppError(op string, err error, message string, code int) *AppError {
	return &AppError{
		Op:      op,
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// IsUserNotFound checks if the error is a "user not found" error.
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}

// IsUserAlreadyExists checks if the error is a "user already exists" error.
func IsUserAlreadyExists(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists)
}

// IsInvalidCredentials checks if the error is an "invalid credentials" error.
func IsInvalidCredentials(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}

// IsInvalidToken checks if the error is an "invalid token" error.
func IsInvalidToken(err error) bool {
	return errors.Is(err, ErrInvalidToken)
}
