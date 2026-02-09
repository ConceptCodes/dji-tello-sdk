package errors

import (
	"fmt"
	"testing"
)

func TestIs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     ErrorCode
		expected bool
	}{
		{
			name:     "SDKError with matching code",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			code:     ErrConnectionFailed,
			expected: true,
		},
		{
			name:     "SDKError with different code",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			code:     ErrCommandFailed,
			expected: false,
		},
		{
			name:     "wrapped SDKError",
			err:      WrapSDKError(fmt.Errorf("cause"), ErrConnectionFailed, "commander", "failed"),
			code:     ErrConnectionFailed,
			expected: true,
		},
		{
			name:     "non-SDKError",
			err:      fmt.Errorf("regular error"),
			code:     ErrConnectionFailed,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			code:     ErrConnectionFailed,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Is(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("Is(%v, %v) = %v, want %v", tt.err, tt.code, result, tt.expected)
			}
		})
	}
}

func TestAs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "SDKError",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: true,
		},
		{
			name:     "wrapped SDKError",
			err:      WrapSDKError(fmt.Errorf("cause"), ErrConnectionFailed, "commander", "failed"),
			expected: true,
		},
		{
			name:     "non-SDKError",
			err:      fmt.Errorf("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target *SDKError
			result := As(tt.err, &target)
			if result != tt.expected {
				t.Errorf("As(%v) = %v, want %v", tt.err, result, tt.expected)
			}
			if result && target == nil {
				t.Error("As() returned true but target is nil")
			}
		})
	}
}

func TestExtractMetadata(t *testing.T) {
	err1 := NewSDKError(ErrConnectionFailed, "commander", "failed").
		WithMetadata("ip", "192.168.10.1").
		WithMetadata("port", "8889")

	err2 := WrapSDKError(err1, ErrCommandFailed, "commander", "command failed").
		WithMetadata("command", "takeoff")

	metadata := ExtractMetadata(err2)

	if len(metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(metadata))
	}

	expected := map[string]string{
		"ip":      "192.168.10.1",
		"port":    "8889",
		"command": "takeoff",
	}

	for key, expectedValue := range expected {
		if value, ok := metadata[key]; !ok || value != expectedValue {
			t.Errorf("metadata[%s] = %v, want %v", key, value, expectedValue)
		}
	}
}

func TestExtractComponent(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "SDKError",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: "commander",
		},
		{
			name:     "wrapped SDKError",
			err:      WrapSDKError(fmt.Errorf("cause"), ErrConnectionFailed, "commander", "failed"),
			expected: "commander",
		},
		{
			name:     "non-SDKError",
			err:      fmt.Errorf("regular error"),
			expected: "",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractComponent(tt.err)
			if result != tt.expected {
				t.Errorf("ExtractComponent(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestExtractOperation(t *testing.T) {
	err := NewSDKError(ErrConnectionFailed, "commander", "failed").
		WithOperation("connect")

	result := ExtractOperation(err)
	if result != "connect" {
		t.Errorf("ExtractOperation() = %v, want 'connect'", result)
	}
}

func TestErrorChain(t *testing.T) {
	root := fmt.Errorf("root cause")
	middle := WrapSDKError(root, ErrConnectionFailed, "commander", "connection failed")
	top := WrapSDKError(middle, ErrCommandFailed, "commander", "command failed")

	chain := ErrorChain(top)

	if len(chain) != 3 {
		t.Errorf("Expected chain length 3, got %d", len(chain))
	}

	if chain[0] != top {
		t.Error("First element should be top error")
	}

	if chain[2] != root {
		t.Error("Last element should be root cause")
	}
}

func TestFormatChain(t *testing.T) {
	root := fmt.Errorf("root cause")
	middle := WrapSDKError(root, ErrConnectionFailed, "commander", "connection failed")
	top := WrapSDKError(middle, ErrCommandFailed, "commander", "command failed")

	formatted := FormatChain(top)

	if formatted == "" {
		t.Error("FormatChain() returned empty string")
	}

	// Should contain all error messages
	if !contains(formatted, "command failed") {
		t.Error("Formatted chain should contain top error message")
	}
	if !contains(formatted, "connection failed") {
		t.Error("Formatted chain should contain middle error message")
	}
	if !contains(formatted, "root cause") {
		t.Error("Formatted chain should contain root cause")
	}
}

func TestRootCause(t *testing.T) {
	root := fmt.Errorf("root cause")
	middle := WrapSDKError(root, ErrConnectionFailed, "commander", "connection failed")
	top := WrapSDKError(middle, ErrCommandFailed, "commander", "command failed")

	result := RootCause(top)
	if result != root {
		t.Errorf("RootCause() = %v, want %v", result, root)
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection failed",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: true,
		},
		{
			name:     "command failed",
			err:      NewSDKError(ErrCommandFailed, "commander", "failed"),
			expected: true,
		},
		{
			name:     "timeout",
			err:      NewSDKError(ErrTimeout, "commander", "timeout"),
			expected: true,
		},
		{
			name:     "network error",
			err:      NewSDKError(ErrNetwork, "commander", "network"),
			expected: true,
		},
		{
			name:     "safety error",
			err:      NewSDKError(ErrSafetyViolation, "safety", "violation"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConnectionError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsSafetyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "safety violation",
			err:      NewSDKError(ErrSafetyViolation, "safety", "violation"),
			expected: true,
		},
		{
			name:     "emergency stop",
			err:      NewSDKError(ErrEmergencyStop, "safety", "emergency"),
			expected: true,
		},
		{
			name:     "altitude violation",
			err:      NewSDKError(ErrAltitudeViolation, "safety", "altitude"),
			expected: true,
		},
		{
			name:     "battery critical",
			err:      NewSDKError(ErrBatteryCritical, "safety", "battery"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSafetyError(tt.err)
			if result != tt.expected {
				t.Errorf("IsSafetyError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "config validation",
			err:      NewSDKError(ErrConfigValidation, "config", "validation"),
			expected: true,
		},
		{
			name:     "config loading",
			err:      NewSDKError(ErrConfigLoading, "config", "loading"),
			expected: true,
		},
		{
			name:     "config not found",
			err:      NewSDKError(ErrConfigNotFound, "config", "not found"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConfigError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConfigError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsMLError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ML pipeline",
			err:      NewSDKError(ErrMLPipeline, "ml", "pipeline"),
			expected: true,
		},
		{
			name:     "model loading",
			err:      NewSDKError(ErrModelLoading, "ml", "loading"),
			expected: true,
		},
		{
			name:     "inference failed",
			err:      NewSDKError(ErrInferenceFailed, "ml", "inference"),
			expected: true,
		},
		{
			name:     "processor failed",
			err:      NewSDKError(ErrProcessorFailed, "ml", "processor"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMLError(tt.err)
			if result != tt.expected {
				t.Errorf("IsMLError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      NewSDKError(ErrTimeout, "commander", "timeout"),
			expected: true,
		},
		{
			name:     "network error",
			err:      NewSDKError(ErrNetwork, "commander", "network"),
			expected: true,
		},
		{
			name:     "connection failed",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: true,
		},
		{
			name:     "error with 'temporary' in message",
			err:      fmt.Errorf("temporary failure"),
			expected: true,
		},
		{
			name:     "error with 'timeout' in message",
			err:      fmt.Errorf("operation timeout"),
			expected: true,
		},
		{
			name:     "safety error",
			err:      NewSDKError(ErrSafetyViolation, "safety", "violation"),
			expected: false,
		},
		{
			name:     "config error",
			err:      NewSDKError(ErrConfigValidation, "config", "validation"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRetry(tt.err)
			if result != tt.expected {
				t.Errorf("ShouldRetry(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "connection failed",
			err:      NewSDKError(ErrConnectionFailed, "commander", "failed"),
			expected: "Failed to connect to drone. Please check WiFi connection and try again.",
		},
		{
			name:     "battery critical",
			err:      NewSDKError(ErrBatteryCritical, "safety", "critical"),
			expected: "Drone battery is critically low. Please land immediately.",
		},
		{
			name:     "safety violation",
			err:      NewSDKError(ErrSafetyViolation, "safety", "violation"),
			expected: "Safety violation detected. Operation aborted for safety reasons.",
		},
		{
			name:     "config validation",
			err:      NewSDKError(ErrConfigValidation, "config", "validation"),
			expected: "Configuration error. Please check your settings and try again.",
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("something went wrong"),
			expected: "Operation failed: something went wrong",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UserFriendlyMessage(tt.err)
			if result != tt.expected {
				t.Errorf("UserFriendlyMessage(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
