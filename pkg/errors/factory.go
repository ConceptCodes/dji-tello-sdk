package errors

import (
	"fmt"
	"runtime/debug"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// debugMode controls whether stack traces are captured
// Set via build tags: go build -tags=debug
var debugMode = false

// NewSDKError creates a new SDKError with stack trace (in debug builds only)
func NewSDKError(code ErrorCode, component, message string) *SDKError {
	err := &SDKError{
		Code:      code,
		Component: component,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}

	if debugMode {
		// Capture stack trace in debug builds
		err.stack = captureStackTrace()
	}

	return err
}

// WrapSDKError wraps an existing error with SDKError structure
func WrapSDKError(err error, code ErrorCode, component, message string) *SDKError {
	sdkErr := NewSDKError(code, component, message)
	sdkErr.cause = err

	// Preserve stack trace from pkg/errors if available
	if stackTracer, ok := err.(interface{ StackTrace() pkgerrors.StackTrace }); ok {
		sdkErr.stack = stackTracer.StackTrace()
	} else if debugMode && sdkErr.stack == nil {
		// Capture new stack trace if not already captured
		sdkErr.stack = captureStackTrace()
	}

	return sdkErr
}

// Wrap wraps an error with additional context using pkg/errors
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return pkgerrors.Wrap(err, message)
}

// Wrapf wraps an error with formatted context using pkg/errors
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return pkgerrors.Wrapf(err, format, args...)
}

// WithMessage adds context to an error without stack trace
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return pkgerrors.WithMessage(err, message)
}

// WithMessagef adds formatted context to an error without stack trace
func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return pkgerrors.WithMessagef(err, format, args...)
}

// Helper functions for common error patterns

// ConnectionError creates a connection-related error
func ConnectionError(component, operation string, cause error) error {
	if cause == nil {
		return NewSDKError(ErrConnectionFailed, component,
			fmt.Sprintf("failed to %s", operation))
	}
	return WrapSDKError(cause, ErrConnectionFailed, component,
		fmt.Sprintf("failed to %s", operation))
}

// CommandError creates a command execution error
func CommandError(cmd string, cause error) error {
	if cause == nil {
		return NewSDKError(ErrCommandFailed, "commander",
			fmt.Sprintf("command '%s' failed", cmd))
	}
	return WrapSDKError(cause, ErrCommandFailed, "commander",
		fmt.Sprintf("command '%s' failed", cmd))
}

// ConfigError creates a configuration-related error
func ConfigError(component, configPath string, cause error) error {
	if cause == nil {
		return NewSDKError(ErrConfigValidation, component,
			fmt.Sprintf("invalid configuration at '%s'", configPath))
	}
	return WrapSDKError(cause, ErrConfigValidation, component,
		fmt.Sprintf("invalid configuration at '%s'", configPath))
}

// ValidationError creates a validation error
func ValidationError(component, field, reason string) *SDKError {
	return NewSDKError(ErrConfigValidation, component,
		fmt.Sprintf("validation failed for field '%s': %s", field, reason))
}

// TimeoutError creates a timeout error
func TimeoutError(component, operation string, timeout time.Duration) *SDKError {
	return NewSDKError(ErrTimeout, component,
		fmt.Sprintf("operation '%s' timed out after %v", operation, timeout))
}

// NotImplementedError creates a not implemented error
func NotImplementedError(component, feature string) *SDKError {
	return NewSDKError(ErrNotImplemented, component,
		fmt.Sprintf("feature '%s' is not implemented", feature))
}

// InvalidArgumentError creates an invalid argument error
func InvalidArgumentError(component, arg, reason string) *SDKError {
	return NewSDKError(ErrInvalidArgument, component,
		fmt.Sprintf("invalid argument '%s': %s", arg, reason))
}

// InternalError creates an internal error (should be used sparingly)
func InternalError(component, message string, cause error) error {
	if cause == nil {
		return NewSDKError(ErrInternal, component, message)
	}
	return WrapSDKError(cause, ErrInternal, component, message)
}

// captureStackTrace captures the current stack trace
func captureStackTrace() pkgerrors.StackTrace {
	// Use pkg/errors to capture stack trace
	err := pkgerrors.New("")
	if stackTracer, ok := err.(interface{ StackTrace() pkgerrors.StackTrace }); ok {
		return stackTracer.StackTrace()
	}

	// Fallback to debug.Stack() if pkg/errors fails
	stack := debug.Stack()
	_ = stack // Use stack to avoid unused variable warning
	// Convert to pkgerrors.StackTrace format (simplified)
	return pkgerrors.StackTrace{}
}
