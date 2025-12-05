package errors

import (
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

// Is checks if an error is of a specific SDKError code
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	// Check if it's an SDKError
	var sdkErr *SDKError
	if As(err, &sdkErr) {
		return sdkErr.Code == code
	}

	// Check wrapped errors
	if cause := pkgerrors.Cause(err); cause != nil && cause != err {
		return Is(cause, code)
	}

	return false
}

// As attempts to convert an error to SDKError
func As(err error, target **SDKError) bool {
	if err == nil {
		return false
	}

	// Direct match
	if sdkErr, ok := err.(*SDKError); ok {
		*target = sdkErr
		return true
	}

	// Check wrapped errors using pkg/errors
	if cause := pkgerrors.Cause(err); cause != nil && cause != err {
		return As(cause, target)
	}

	return false
}

// ExtractMetadata extracts metadata from an error chain
func ExtractMetadata(err error) map[string]string {
	metadata := make(map[string]string)

	// Collect metadata from all SDKErrors in the chain
	for err != nil {
		if sdkErr, ok := err.(*SDKError); ok {
			// Merge metadata (later errors override earlier ones)
			for k, v := range sdkErr.Metadata {
				metadata[k] = v
			}
		}

		// Unwrap to next error in chain
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}

	return metadata
}

// ExtractComponent extracts the component from the first SDKError in chain
func ExtractComponent(err error) string {
	if sdkErr, ok := err.(*SDKError); ok {
		return sdkErr.Component
	}

	if cause := pkgerrors.Cause(err); cause != nil && cause != err {
		return ExtractComponent(cause)
	}

	return ""
}

// ExtractOperation extracts the operation from the first SDKError in chain
func ExtractOperation(err error) string {
	if sdkErr, ok := err.(*SDKError); ok {
		return sdkErr.Operation
	}

	if cause := pkgerrors.Cause(err); cause != nil && cause != err {
		return ExtractOperation(cause)
	}

	return ""
}

// ErrorChain returns a slice of all errors in the chain
func ErrorChain(err error) []error {
	var chain []error

	for err != nil {
		chain = append(chain, err)

		// Unwrap to next error
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}

	return chain
}

// FormatChain formats the error chain as a string
func FormatChain(err error) string {
	var builder strings.Builder
	chain := ErrorChain(err)

	for i, link := range chain {
		if i > 0 {
			builder.WriteString("\n  caused by: ")
		}
		builder.WriteString(link.Error())
	}

	return builder.String()
}

// HasStackTrace checks if any error in the chain has a stack trace
func HasStackTrace(err error) bool {
	for err != nil {
		if _, ok := err.(interface{ StackTrace() pkgerrors.StackTrace }); ok {
			return true
		}

		// Unwrap to next error
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}

	return false
}

// RootCause returns the deepest cause in the error chain
func RootCause(err error) error {
	if err == nil {
		return nil
	}

	// Use pkg/errors Cause for compatibility
	if cause := pkgerrors.Cause(err); cause != nil && cause != err {
		return RootCause(cause)
	}

	return err
}

// IsConnectionError checks if error is any connection-related error
func IsConnectionError(err error) bool {
	return Is(err, ErrConnectionFailed) ||
		Is(err, ErrCommandFailed) ||
		Is(err, ErrTimeout) ||
		Is(err, ErrNetwork)
}

// IsSafetyError checks if error is any safety-related error
func IsSafetyError(err error) bool {
	return Is(err, ErrSafetyViolation) ||
		Is(err, ErrEmergencyStop) ||
		Is(err, ErrAltitudeViolation) ||
		Is(err, ErrBatteryCritical)
}

// IsConfigError checks if error is any configuration-related error
func IsConfigError(err error) bool {
	return Is(err, ErrConfigValidation) ||
		Is(err, ErrConfigLoading) ||
		Is(err, ErrConfigNotFound)
}

// IsMLError checks if error is any ML-related error
func IsMLError(err error) bool {
	return Is(err, ErrMLPipeline) ||
		Is(err, ErrModelLoading) ||
		Is(err, ErrInferenceFailed) ||
		Is(err, ErrProcessorFailed)
}

// ShouldRetry determines if an operation should be retried based on error type
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Retry on transient errors
	switch {
	case Is(err, ErrTimeout):
		return true
	case Is(err, ErrNetwork):
		return true
	case Is(err, ErrConnectionFailed):
		return true
	case strings.Contains(strings.ToLower(err.Error()), "temporary"):
		return true
	case strings.Contains(strings.ToLower(err.Error()), "timeout"):
		return true
	}

	return false
}

// UserFriendlyMessage creates a user-friendly error message
func UserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	// Extract root cause for user message
	root := RootCause(err)

	// Check for specific error types
	if Is(root, ErrConnectionFailed) {
		return "Failed to connect to drone. Please check WiFi connection and try again."
	}

	if Is(root, ErrBatteryCritical) {
		return "Drone battery is critically low. Please land immediately."
	}

	if Is(root, ErrSafetyViolation) {
		return "Safety violation detected. Operation aborted for safety reasons."
	}

	if Is(root, ErrConfigValidation) {
		return "Configuration error. Please check your settings and try again."
	}

	// Generic message for other errors
	return fmt.Sprintf("Operation failed: %v", root)
}
