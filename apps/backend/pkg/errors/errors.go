// Package pkgerrors provides application error types and utilities
package pkgerrors

import (
	"errors"
	"fmt"
)

// Common application errors (DRY - defined once)
var (
	ErrVMNotFound       = errors.New("vm not found")
	ErrQuotaExceeded    = errors.New("quota exceeded")
	ErrInvalidTier      = errors.New("invalid tier")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrInternalServer   = errors.New("internal server error")
	ErrUserNotFound     = errors.New("user not found")
	ErrQuotaNotFound    = errors.New("quota not found")
	ErrVMNotStopped     = errors.New("VM must be stopped to start")
	ErrVMNotRunning     = errors.New("VM must be running to stop")
	ErrTierNotAvailable = errors.New("tier not available for your role")
)

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Is reports whether err matches the target error
func Is(err, target error) bool {
	return errors.Is(err, target)
}
