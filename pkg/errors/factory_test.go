package errors

import (
	"fmt"
	"testing"
	"time"
)

func TestNewSDKError(t *testing.T) {
	err := NewSDKError(ErrConnectionFailed, "commander", "failed to connect")

	if err.Code != ErrConnectionFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrConnectionFailed)
	}

	if err.Component != "commander" {
		t.Errorf("Component = %v, want 'commander'", err.Component)
	}

	if err.Message != "failed to connect" {
		t.Errorf("Message = %v, want 'failed to connect'", err.Message)
	}

	if err.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestWrapSDKError(t *testing.T) {
	originalErr := fmt.Errorf("network timeout")
	wrappedErr := WrapSDKError(originalErr, ErrTimeout, "commander", "operation timed out")

	if wrappedErr.Code != ErrTimeout {
		t.Errorf("Code = %v, want %v", wrappedErr.Code, ErrTimeout)
	}

	if wrappedErr.Component != "commander" {
		t.Errorf("Component = %v, want 'commander'", wrappedErr.Component)
	}

	if wrappedErr.Message != "operation timed out" {
		t.Errorf("Message = %v, want 'operation timed out'", wrappedErr.Message)
	}

	if wrappedErr.Cause() != originalErr {
		t.Errorf("Cause() = %v, want %v", wrappedErr.Cause(), originalErr)
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, "additional context")

	if wrappedErr == nil {
		t.Error("Wrap() returned nil")
	}

	if wrappedErr.Error() == originalErr.Error() {
		t.Error("Wrap() should add context to error message")
	}
}

func TestWrapf(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrapf(originalErr, "error with %s", "format")

	if wrappedErr == nil {
		t.Error("Wrapf() returned nil")
	}

	if wrappedErr.Error() == originalErr.Error() {
		t.Error("Wrapf() should add formatted context to error message")
	}
}

func TestWithMessage(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	msgErr := WithMessage(originalErr, "with message")

	if msgErr == nil {
		t.Error("WithMessage() returned nil")
	}

	if msgErr.Error() == originalErr.Error() {
		t.Error("WithMessage() should add message to error")
	}
}

func TestWithMessagef(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	msgErr := WithMessagef(originalErr, "formatted %s", "message")

	if msgErr == nil {
		t.Error("WithMessagef() returned nil")
	}

	if msgErr.Error() == originalErr.Error() {
		t.Error("WithMessagef() should add formatted message to error")
	}
}

func TestConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		cause    error
		expected string
	}{
		{
			name:     "without cause",
			cause:    nil,
			expected: "[CONNECTION_FAILED] commander.: failed to connect",
		},
		{
			name:     "with cause",
			cause:    fmt.Errorf("network unreachable"),
			expected: "[CONNECTION_FAILED] commander.: failed to connect (caused by: network unreachable)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConnectionError("commander", "connect", tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("ConnectionError() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestCommandError(t *testing.T) {
	tests := []struct {
		name     string
		cause    error
		expected string
	}{
		{
			name:     "without cause",
			cause:    nil,
			expected: "[COMMAND_FAILED] commander.: command 'takeoff' failed",
		},
		{
			name:     "with cause",
			cause:    fmt.Errorf("timeout"),
			expected: "[COMMAND_FAILED] commander.: command 'takeoff' failed (caused by: timeout)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CommandError("takeoff", tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("CommandError() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		cause    error
		expected string
	}{
		{
			name:     "without cause",
			cause:    nil,
			expected: "[CONFIG_VALIDATION] config.: invalid configuration at '/path/to/config.json'",
		},
		{
			name:     "with cause",
			cause:    fmt.Errorf("invalid JSON"),
			expected: "[CONFIG_VALIDATION] config.: invalid configuration at '/path/to/config.json' (caused by: invalid JSON)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConfigError("config", "/path/to/config.json", tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("ConfigError() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("config", "speed", "must be between 10 and 100")
	expected := "[CONFIG_VALIDATION] config.: validation failed for field 'speed': must be between 10 and 100"

	if err.Error() != expected {
		t.Errorf("ValidationError() = %v, want %v", err.Error(), expected)
	}
}

func TestTimeoutError(t *testing.T) {
	err := TimeoutError("commander", "connect", 5*time.Second)
	expected := "[TIMEOUT] commander.: operation 'connect' timed out after 5s"

	if err.Error() != expected {
		t.Errorf("TimeoutError() = %v, want %v", err.Error(), expected)
	}
}

func TestNotImplementedError(t *testing.T) {
	err := NotImplementedError("ml", "face recognition")
	expected := "[NOT_IMPLEMENTED] ml.: feature 'face recognition' is not implemented"

	if err.Error() != expected {
		t.Errorf("NotImplementedError() = %v, want %v", err.Error(), expected)
	}
}

func TestInvalidArgumentError(t *testing.T) {
	err := InvalidArgumentError("commander", "speed", "must be positive")
	expected := "[INVALID_ARGUMENT] commander.: invalid argument 'speed': must be positive"

	if err.Error() != expected {
		t.Errorf("InvalidArgumentError() = %v, want %v", err.Error(), expected)
	}
}

func TestInternalError(t *testing.T) {
	tests := []struct {
		name     string
		cause    error
		expected string
	}{
		{
			name:     "without cause",
			cause:    nil,
			expected: "[INTERNAL_ERROR] commander.: unexpected error",
		},
		{
			name:     "with cause",
			cause:    fmt.Errorf("nil pointer"),
			expected: "[INTERNAL_ERROR] commander.: unexpected error (caused by: nil pointer)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InternalError("commander", "unexpected error", tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("InternalError() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}
