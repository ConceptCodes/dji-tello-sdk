package errors

import (
	"fmt"
	"testing"
)

func TestRecover(t *testing.T) {
	// Test that Recover returns nil when no panic occurred
	t.Run("no panic", func(t *testing.T) {
		// Create a function that doesn't panic
		fn := func() {
			// Do nothing
		}

		// Call it with recover
		var recoveredErr error
		func() {
			defer func() {
				recoveredErr = Recover("test", "operation")
			}()
			fn()
		}()

		if recoveredErr != nil {
			t.Errorf("Recover() = %v, want nil", recoveredErr)
		}
	})
}

func TestRecoverWithHandler(t *testing.T) {
	t.Run("with handler", func(t *testing.T) {
		handlerCalled := false
		customHandler := func(err error) error {
			handlerCalled = true
			return fmt.Errorf("custom: %w", err)
		}

		// Test without panic
		err := func() error {
			defer func() {
				_ = RecoverWithHandler("test", "operation", customHandler)
			}()
			return nil
		}()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Handler shouldn't be called without panic
		if handlerCalled {
			t.Error("Handler should not be called without panic")
		}
	})
}

func TestSafeExecute(t *testing.T) {
	t.Run("function returns error", func(t *testing.T) {
		err := SafeExecute("test", "operation", func() error {
			return fmt.Errorf("function error")
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("function succeeds", func(t *testing.T) {
		err := SafeExecute("test", "operation", func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestSafeExecuteWithResult(t *testing.T) {
	t.Run("function returns value", func(t *testing.T) {
		value, err := SafeExecuteWithResult("test", "operation", func() (string, error) {
			return "success", nil
		})

		if value != "success" {
			t.Errorf("Value = %v, want 'success'", value)
		}

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("function returns error", func(t *testing.T) {
		value, err := SafeExecuteWithResult("test", "operation", func() (string, error) {
			return "", fmt.Errorf("function error")
		})

		if value != "" {
			t.Errorf("Value = %v, want empty string", value)
		}

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestWithRecovery(t *testing.T) {
	t.Run("wraps function", func(t *testing.T) {
		normalFn := func() error {
			return fmt.Errorf("normal error")
		}

		wrappedFn := WithRecovery("test", "operation", normalFn)
		err := wrappedFn()

		if err == nil {
			t.Error("Expected error from wrapped function")
		}
	})
}

func TestWithRecoveryAndResult(t *testing.T) {
	t.Run("wraps function with result", func(t *testing.T) {
		normalFn := func() (string, error) {
			return "result", fmt.Errorf("normal error")
		}

		wrappedFn := WithRecoveryAndResult("test", "operation", normalFn)
		value, err := wrappedFn()

		if value != "result" {
			t.Errorf("Value = %v, want 'result'", value)
		}

		if err == nil {
			t.Error("Expected error from wrapped function")
		}
	})
}

func TestCaptureStack(t *testing.T) {
	stack := CaptureStack(0)

	if stack.File == "" {
		t.Error("CaptureStack() should return file path")
	}

	if stack.Line <= 0 {
		t.Error("CaptureStack() should return line number")
	}

	if stack.Function == "" {
		t.Error("CaptureStack() should return function name")
	}
}

func TestWithStackInfo(t *testing.T) {
	t.Run("with SDKError", func(t *testing.T) {
		originalErr := NewSDKError(ErrConnectionFailed, "commander", "failed")
		wrappedErr := WithStackInfo(originalErr, 0)

		if wrappedErr == nil {
			t.Error("WithStackInfo() returned nil")
		}

		// Check if it's an SDKError with stack info
		var sdkErr *SDKError
		if As(wrappedErr, &sdkErr) {
			if sdkErr.Metadata["stack_file"] == "" {
				t.Error("Stack info should include file path")
			}
		}
	})

	t.Run("with non-SDKError", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		wrappedErr := WithStackInfo(originalErr, 0)

		if wrappedErr == nil {
			t.Error("WithStackInfo() returned nil")
		}

		// Should be wrapped as SDKError
		var sdkErr *SDKError
		if !As(wrappedErr, &sdkErr) {
			t.Error("WithStackInfo() should wrap non-SDKError as SDKError")
		}
	})
}

func TestPanicToError(t *testing.T) {
	tests := []struct {
		name     string
		panicVal interface{}
		expected string
	}{
		{
			name:     "error panic",
			panicVal: fmt.Errorf("panic error"),
			expected: "[INTERNAL_ERROR] test.: panic during operation (caused by: panic error)",
		},
		{
			name:     "string panic",
			panicVal: "panic string",
			expected: "[INTERNAL_ERROR] test.: panic during operation (caused by: panic: panic string)",
		},
		{
			name:     "nil panic",
			panicVal: nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PanicToError(tt.panicVal, "test", "operation")

			if tt.panicVal == nil {
				if err != nil {
					t.Errorf("PanicToError() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Error("PanicToError() returned nil, expected error")
				} else if err.Error() != tt.expected {
					t.Errorf("PanicToError() = %v, want %v", err.Error(), tt.expected)
				}
			}
		})
	}
}

func TestMust(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		value, err := Must("success", nil)
		if err != nil {
			t.Errorf("Must() unexpected error: %v", err)
		}
		if value != "success" {
			t.Errorf("Must() = %v, want 'success'", value)
		}
	})

	t.Run("with error", func(t *testing.T) {
		value, err := Must("", fmt.Errorf("test error"))
		if err == nil {
			t.Error("Must() should return error when error is not nil")
		}
		if value != "" {
			t.Errorf("Must() should return zero value on error, got: %v", value)
		}
	})
}

func TestMustWithMessage(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		value, err := MustWithMessage("success", nil, "custom message")
		if err != nil {
			t.Errorf("MustWithMessage() unexpected error: %v", err)
		}
		if value != "success" {
			t.Errorf("MustWithMessage() = %v, want 'success'", value)
		}
	})

	t.Run("with error", func(t *testing.T) {
		value, err := MustWithMessage("", fmt.Errorf("original error"), "custom message")
		if err == nil {
			t.Error("MustWithMessage() should return error when error is not nil")
		}
		if value != "" {
			t.Errorf("MustWithMessage() should return zero value on error, got: %v", value)
		}
		if !containsString(err.Error(), "custom message") {
			t.Errorf("Error message should contain custom message, got: %v", err.Error())
		}
	})
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}
