package errors

import (
	"fmt"
	"runtime"
)

// Recover converts a panic to an error
func Recover(component, operation string) error {
	if r := recover(); r != nil {
		// Create error from panic
		var panicErr error
		switch v := r.(type) {
		case error:
			panicErr = v
		case string:
			panicErr = fmt.Errorf("panic: %s", v)
		default:
			panicErr = fmt.Errorf("panic: %v", v)
		}

		// Wrap with SDKError
		return WrapSDKError(panicErr, ErrInternal, component,
			fmt.Sprintf("panic during %s", operation)).
			WithMetadata("panic", "true").
			WithMetadata("recovered", "true")
	}
	return nil
}

// RecoverWithHandler converts a panic to an error with custom handler
func RecoverWithHandler(component, operation string, handler func(error) error) error {
	if r := recover(); r != nil {
		// Create error from panic
		var panicErr error
		switch v := r.(type) {
		case error:
			panicErr = v
		case string:
			panicErr = fmt.Errorf("panic: %s", v)
		default:
			panicErr = fmt.Errorf("panic: %v", v)
		}

		// Wrap with SDKError
		sdkErr := WrapSDKError(panicErr, ErrInternal, component,
			fmt.Sprintf("panic during %s", operation)).
			WithMetadata("panic", "true").
			WithMetadata("recovered", "true")

		// Call custom handler if provided
		if handler != nil {
			return handler(sdkErr)
		}

		return sdkErr
	}
	return nil
}

// SafeExecute executes a function with panic recovery
func SafeExecute(component, operation string, fn func() error) (err error) {
	defer func() {
		if recovered := Recover(component, operation); recovered != nil {
			if err != nil {
				// Combine original error with panic error
				err = WrapSDKError(recovered, ErrInternal, component,
					fmt.Sprintf("panic after error during %s", operation))
			} else {
				err = recovered
			}
		}
	}()

	return fn()
}

// SafeExecuteWithResult executes a function with panic recovery and returns a result
func SafeExecuteWithResult[T any](component, operation string, fn func() (T, error)) (result T, err error) {
	defer func() {
		if recovered := Recover(component, operation); recovered != nil {
			if err != nil {
				// Combine original error with panic error
				err = WrapSDKError(recovered, ErrInternal, component,
					fmt.Sprintf("panic after error during %s", operation))
			} else {
				err = recovered
			}
		}
	}()

	return fn()
}

// WithRecovery wraps a function with panic recovery
func WithRecovery(component, operation string, fn func() error) func() error {
	return func() error {
		return SafeExecute(component, operation, fn)
	}
}

// WithRecoveryAndResult wraps a function with panic recovery and result
func WithRecoveryAndResult[T any](component, operation string, fn func() (T, error)) func() (T, error) {
	return func() (T, error) {
		return SafeExecuteWithResult(component, operation, fn)
	}
}

// StackInfo captures stack information at point of creation
type StackInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// CaptureStack captures current stack information
func CaptureStack(skip int) StackInfo {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return StackInfo{}
	}

	fn := runtime.FuncForPC(pc)
	var fnName string
	if fn != nil {
		fnName = fn.Name()
	}

	return StackInfo{
		File:     file,
		Line:     line,
		Function: fnName,
	}
}

// WithStackInfo adds stack information to an error
func WithStackInfo(err error, skip int) error {
	if err == nil {
		return nil
	}

	stack := CaptureStack(skip + 1)

	// Add stack info as metadata if it's an SDKError
	if sdkErr, ok := err.(*SDKError); ok {
		sdkErr.WithMetadata("stack_file", stack.File).
			WithMetadata("stack_line", fmt.Sprintf("%d", stack.Line)).
			WithMetadata("stack_function", stack.Function)
		return sdkErr
	}

	// For non-SDKError, wrap it
	return WrapSDKError(err, ErrInternal, "recovery",
		fmt.Sprintf("error at %s:%d in %s", stack.File, stack.Line, stack.Function))
}

// PanicToError converts a panic value to an error
func PanicToError(r interface{}, component, operation string) error {
	if r == nil {
		return nil
	}

	var panicErr error
	switch v := r.(type) {
	case error:
		panicErr = v
	case string:
		panicErr = fmt.Errorf("panic: %s", v)
	default:
		panicErr = fmt.Errorf("panic: %v", v)
	}

	return WrapSDKError(panicErr, ErrInternal, component,
		fmt.Sprintf("panic during %s", operation)).
		WithMetadata("panic", "true")
}

// Must returns the value if err is nil, otherwise returns the zero value and the error.
// Unlike MustWithMessage, this does not panic - it returns errors gracefully for production safety.
func Must[T any](value T, err error) (T, error) {
	if err != nil {
		var zero T
		return zero, err
	}
	return value, nil
}

// MustWithMessage returns the value if err is nil, otherwise wraps the error with the provided message.
// This is a non-panicking version that returns errors gracefully for production safety.
func MustWithMessage[T any](value T, err error, message string) (T, error) {
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%s: %w", message, err)
	}
	return value, nil
}
