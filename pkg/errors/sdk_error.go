package errors

import (
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// ErrorCode represents specific error categories in the SDK
type ErrorCode string

const (
	// Connection errors
	ErrConnectionFailed ErrorCode = "CONNECTION_FAILED"
	ErrCommandFailed    ErrorCode = "COMMAND_FAILED"
	ErrTimeout          ErrorCode = "TIMEOUT"
	ErrNetwork          ErrorCode = "NETWORK_ERROR"

	// Video/Stream errors
	ErrVideoStreamFailed ErrorCode = "VIDEO_STREAM_FAILED"
	ErrFrameProcessing   ErrorCode = "FRAME_PROCESSING"
	ErrVideoDecode       ErrorCode = "VIDEO_DECODE"

	// Safety errors
	ErrSafetyViolation   ErrorCode = "SAFETY_VIOLATION"
	ErrEmergencyStop     ErrorCode = "EMERGENCY_STOP"
	ErrAltitudeViolation ErrorCode = "ALTITUDE_VIOLATION"
	ErrBatteryCritical   ErrorCode = "BATTERY_CRITICAL"

	// ML/Processing errors
	ErrMLPipeline      ErrorCode = "ML_PIPELINE"
	ErrModelLoading    ErrorCode = "MODEL_LOADING"
	ErrInferenceFailed ErrorCode = "INFERENCE_FAILED"
	ErrProcessorFailed ErrorCode = "PROCESSOR_FAILED"

	// Configuration errors
	ErrConfigValidation ErrorCode = "CONFIG_VALIDATION"
	ErrConfigLoading    ErrorCode = "CONFIG_LOADING"
	ErrConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"

	// Gamepad errors
	ErrGamepadConnection ErrorCode = "GAMEPAD_CONNECTION"
	ErrGamepadConfig     ErrorCode = "GAMEPAD_CONFIG"

	// Web/UI errors
	ErrWebServer         ErrorCode = "WEB_SERVER"
	ErrCSRFValidation    ErrorCode = "CSRF_VALIDATION"
	ErrRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// General errors
	ErrInvalidArgument   ErrorCode = "INVALID_ARGUMENT"
	ErrNotImplemented    ErrorCode = "NOT_IMPLEMENTED"
	ErrResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	ErrInternal          ErrorCode = "INTERNAL_ERROR"
)

// SDKError is the main error type for the Tello SDK
type SDKError struct {
	Code      ErrorCode         `json:"code"`
	Message   string            `json:"message"`
	Component string            `json:"component"`
	Operation string            `json:"operation,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`

	cause error
	stack pkgerrors.StackTrace
}

// Error implements the error interface
func (e *SDKError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s.%s: %s (caused by: %v)",
			e.Code, e.Component, e.Operation, e.Message, e.cause)
	}
	return fmt.Sprintf("[%s] %s.%s: %s", e.Code, e.Component, e.Operation, e.Message)
}

// Cause returns the underlying error
func (e *SDKError) Cause() error {
	return e.cause
}

// StackTrace returns the stack trace (only populated in debug builds)
func (e *SDKError) StackTrace() pkgerrors.StackTrace {
	return e.stack
}

// Unwrap implements Go 1.13 error unwrapping
func (e *SDKError) Unwrap() error {
	return e.cause
}

// WithMetadata adds metadata to the error
func (e *SDKError) WithMetadata(key, value string) *SDKError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// WithOperation sets the operation context
func (e *SDKError) WithOperation(op string) *SDKError {
	e.Operation = op
	return e
}
