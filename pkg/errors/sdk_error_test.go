package errors

import (
	"fmt"
	"testing"
	"time"
)

func TestSDKError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SDKError
		expected string
	}{
		{
			name: "basic error without cause",
			err: &SDKError{
				Code:      ErrConnectionFailed,
				Component: "commander",
				Operation: "connect",
				Message:   "failed to establish connection",
				Timestamp: time.Now(),
			},
			expected: "[CONNECTION_FAILED] commander.connect: failed to establish connection",
		},
		{
			name: "error with cause",
			err: &SDKError{
				Code:      ErrCommandFailed,
				Component: "commander",
				Operation: "takeoff",
				Message:   "command execution failed",
				Timestamp: time.Now(),
				cause:     fmt.Errorf("timeout"),
			},
			expected: "[COMMAND_FAILED] commander.takeoff: command execution failed (caused by: timeout)",
		},
		{
			name: "error with metadata",
			err: &SDKError{
				Code:      ErrConfigValidation,
				Component: "config",
				Operation: "validate",
				Message:   "invalid configuration",
				Timestamp: time.Now(),
				Metadata: map[string]string{
					"field": "speed",
					"value": "200",
				},
			},
			expected: "[CONFIG_VALIDATION] config.validate: invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSDKError_Cause(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := &SDKError{
		Code:    ErrInternal,
		cause:   cause,
		Message: "wrapped error",
	}

	if err.Cause() != cause {
		t.Errorf("Cause() = %v, want %v", err.Cause(), cause)
	}
}

func TestSDKError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := &SDKError{
		Code:    ErrInternal,
		cause:   cause,
		Message: "wrapped error",
	}

	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestSDKError_WithMetadata(t *testing.T) {
	err := NewSDKError(ErrConfigValidation, "config", "validation failed")
	err.WithMetadata("field", "speed").WithMetadata("limit", "100")

	if len(err.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(err.Metadata))
	}

	if err.Metadata["field"] != "speed" {
		t.Errorf("Metadata['field'] = %v, want 'speed'", err.Metadata["field"])
	}

	if err.Metadata["limit"] != "100" {
		t.Errorf("Metadata['limit'] = %v, want '100'", err.Metadata["limit"])
	}
}

func TestSDKError_WithOperation(t *testing.T) {
	err := NewSDKError(ErrConnectionFailed, "commander", "connection failed")
	err.WithOperation("connect")

	if err.Operation != "connect" {
		t.Errorf("Operation = %v, want 'connect'", err.Operation)
	}
}

func TestSDKError_StackTrace(t *testing.T) {
	// Test that StackTrace returns a StackTrace type (even if empty)
	err := NewSDKError(ErrInternal, "test", "test error")
	stack := err.StackTrace()

	// StackTrace should return a pkgerrors.StackTrace type
	// It might be empty if debugMode is false, but should not panic
	_ = stack // Use stack to avoid unused variable warning
	// Just verify the method doesn't panic
}
