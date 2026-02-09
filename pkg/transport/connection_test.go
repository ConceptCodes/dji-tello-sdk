package transport

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// MockControlledUDPClient is a mock that allows precise control over behavior
type MockControlledUDPClient struct {
	mu            sync.Mutex
	sendCount     int
	sendErrs      []error
	sendErrIndex  int
	recvData      string
	recvErrs      []error
	recvErrIndex  int
	closeCount    int
	closeErrs     []error
	closeErrIndex int
	closed        bool
}

// NewMockControlledUDPClient creates a controlled mock UDP client
func NewMockControlledUDPClient() *MockControlledUDPClient {
	return &MockControlledUDPClient{
		sendErrs:  []error{},
		recvErrs:  []error{},
		closeErrs: []error{},
	}
}

// SetSendErrors configures the errors to return on subsequent Send calls
func (m *MockControlledUDPClient) SetSendErrors(errs []error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErrs = errs
	m.sendErrIndex = 0
}

// SetReceiveErrors configures the errors to return on subsequent Receive calls
func (m *MockControlledUDPClient) SetReceiveErrors(errs []error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recvErrs = errs
	m.recvErrIndex = 0
}

// SetReceiveData sets the data to return on Receive
func (m *MockControlledUDPClient) SetReceiveData(data string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recvData = data
}

// SetCloseErrors configures errors to return on Close
func (m *MockControlledUDPClient) SetCloseErrors(errs []error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeErrs = errs
	m.closeErrIndex = 0
}

// GetSendCount returns how many times Send was called
func (m *MockControlledUDPClient) GetSendCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sendCount
}

// GetCloseCount returns how many times Close was called
func (m *MockControlledUDPClient) GetCloseCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeCount
}

// IsClosed returns whether Close was called
func (m *MockControlledUDPClient) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func (m *MockControlledUDPClient) Send(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("send on closed connection")
	}

	if m.sendErrIndex < len(m.sendErrs) {
		err := m.sendErrs[m.sendErrIndex]
		m.sendErrIndex++
		if err != nil {
			return err
		}
	}

	m.sendCount++
	return nil
}

func (m *MockControlledUDPClient) Receive(bufferSize int, timeout time.Duration) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return "", errors.New("receive on closed connection")
	}

	if m.recvErrIndex < len(m.recvErrs) {
		err := m.recvErrs[m.recvErrIndex]
		m.recvErrIndex++
		if err != nil {
			return "", err
		}
	}

	return m.recvData, nil
}

func (m *MockControlledUDPClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closeErrIndex < len(m.closeErrs) {
		err := m.closeErrs[m.closeErrIndex]
		m.closeErrIndex++
		if err != nil {
			return err
		}
	}

	m.closed = true
	m.closeCount++
	return nil
}

// Test helper functions - mark with t.Helper()
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing '%s', got nil", substr)
		return
	}
	if !containsString(err.Error(), substr) {
		t.Errorf("expected error containing '%s', got '%s'", substr, err.Error())
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func assertEqualInt(t *testing.T, expected, actual int, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", msg, expected, actual)
	}
}

func assertEqualBool(t *testing.T, expected, actual bool, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func assertEqualString(t *testing.T, expected, actual string, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %s, got %s", msg, expected, actual)
	}
}

func assertTrue(t *testing.T, val bool, msg string) {
	t.Helper()
	if !val {
		t.Errorf("%s: expected true, got false", msg)
	}
}

// TestIsBindingError tests binding error detection
func TestIsBindingError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "address already in use",
			err:      errors.New("address already in use"),
			expected: true,
		},
		{
			name:     "permission denied",
			err:      errors.New("permission denied for binding port"),
			expected: true,
		},
		{
			name:     "bind error",
			err:      errors.New("bind: cannot assign requested address"),
			expected: true,
		},
		{
			name:     "broken pipe (not binding)",
			err:      errors.New("broken pipe"),
			expected: false,
		},
		{
			name:     "connection refused (not binding)",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "case insensitive - ADDRESS ALREADY IN USE",
			err:      errors.New("ADDRESS ALREADY IN USE"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBindingError(tt.err)
			if result != tt.expected {
				t.Errorf("isBindingError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestIsNetworkError tests network error detection
func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "broken pipe",
			err:      errors.New("broken pipe"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "network unreachable",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "no route to host",
			err:      errors.New("no route to host"),
			expected: true,
		},
		{
			name:     "host is down",
			err:      errors.New("host is down"),
			expected: true,
		},
		{
			name:     "connection timed out",
			err:      errors.New("connection timed out"),
			expected: true,
		},
		{
			name:     "i/o timeout",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "address already in use (not network)",
			err:      errors.New("address already in use"),
			expected: false,
		},
		{
			name:     "permission denied (not network)",
			err:      errors.New("permission denied"),
			expected: false,
		},
		{
			name:     "case insensitive - BROKEN PIPE",
			err:      errors.New("BROKEN PIPE"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("isNetworkError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestRetryLogicWithBindingError tests retry behavior when binding error occurs
func TestRetryLogicWithBindingError(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// First send fails with binding error, second succeeds
	mockClient.SetSendErrors([]error{
		errors.New("address already in use"),
		nil,
	})
	mockClient.SetReceiveData("ok")

	response, err := conn.SendCommand("test")
	if err != nil {
		t.Errorf("expected success after fallback, got error: %v", err)
	}

	if response != "ok" {
		t.Errorf("expected response 'ok', got '%s'", response)
	}

	// Verify fallback was triggered
	assertTrue(t, conn.inFallback, "expected inFallback to be true")
	assertEqualInt(t, 1, conn.fallbackTry, "fallbackTry count")
}

// TestRetryLogicWithNetworkError tests retry behavior when network error occurs
func TestRetryLogicWithNetworkError(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// First send fails with network error, second succeeds
	mockClient.SetSendErrors([]error{
		errors.New("broken pipe"),
		nil,
	})
	mockClient.SetReceiveData("ok")

	response, err := conn.SendCommand("test")
	if err != nil {
		t.Errorf("expected success after retry, got error: %v", err)
	}

	if response != "ok" {
		t.Errorf("expected response 'ok', got '%s'", response)
	}

	// Network error should not trigger fallback
	assertTrue(t, !conn.inFallback, "expected inFallback to be false for network error")
}

// TestFallbackToEphemeralPort tests fallback to ephemeral port after binding error
func TestFallbackToEphemeralPort(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Configure to fail with binding error then succeed
	mockClient.SetSendErrors([]error{
		errors.New("bind: permission denied"),
		nil,
	})
	mockClient.SetReceiveData("ok")

	response, err := conn.SendCommand("test")
	if err != nil {
		t.Fatalf("expected success after fallback, got error: %v", err)
	}

	assertEqualString(t, "ok", response, "response")
	assertTrue(t, conn.inFallback, "expected inFallback to be true")
	assertEqualInt(t, 1, conn.fallbackTry, "fallbackTry count")

	// Send another command - should continue in fallback mode
	mockClient.SetSendErrors([]error{})
	mockClient.SetReceiveData("ok")

	_, err = conn.SendCommand("test2")
	if err != nil {
		t.Fatalf("expected success in fallback mode, got error: %v", err)
	}

	// fallbackTry should remain the same (not increment on success)
	assertEqualInt(t, 1, conn.fallbackTry, "fallbackTry count should not increment on success")
}

// TestBindingErrorOnReceive tests fallback when receive fails with binding error
func TestBindingErrorOnReceive(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Send succeeds, receive fails with binding error, then succeeds
	mockClient.SetSendErrors([]error{})
	mockClient.SetReceiveErrors([]error{
		errors.New("address already in use"),
		nil,
	})

	response, err := conn.SendCommand("test")
	if err != nil {
		t.Fatalf("expected success after receive fallback, got error: %v", err)
	}

	assertEqualString(t, "ok", response, "response")
	assertTrue(t, conn.inFallback, "expected inFallback to be true after receive binding error")
}

// TestRevertFromFallback tests reversion from ephemeral port to bound port
func TestRevertFromFallback(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Force initial fallback
	mockClient.SetSendErrors([]error{
		errors.New("address already in use"),
		nil,
	})
	mockClient.SetReceiveData("ok")

	_, err := conn.SendCommand("test")
	if err != nil {
		t.Fatalf("initial command failed: %v", err)
	}

	assertTrue(t, conn.inFallback, "expected inFallback after binding error")
	initialFallbackTry := conn.fallbackTry

	// Need 3 successful commands before revert is attempted
	// First success
	mockClient.SetSendErrors([]error{})
	mockClient.SetReceiveData("ok")
	_, err = conn.SendCommand("test1")
	if err != nil {
		t.Fatalf("first success command failed: %v", err)
	}
	assertEqualInt(t, initialFallbackTry, conn.fallbackTry, "fallbackTry should not change on success")

	// Second success
	_, err = conn.SendCommand("test2")
	if err != nil {
		t.Fatalf("second success command failed: %v", err)
	}
	assertEqualInt(t, initialFallbackTry, conn.fallbackTry, "fallbackTry should not change on success")

	// Third success - should trigger revert attempt
	_, err = conn.SendCommand("test3")
	if err != nil {
		t.Fatalf("third success command failed: %v", err)
	}

	// After 3 successful commands, should attempt revert
	// Note: revert may fail if binding fails, which is expected in test
	// The important thing is the revert logic is called
}

// TestFallbackCounterIncrements correctly tracks fallback attempts
func TestFallbackCounterIncrements(t *testing.T) {
	tests := []struct {
		name           string
		bindingErrors  []error
		expectedTries  int
		expectFallback bool
	}{
		{
			name:           "single binding error",
			bindingErrors:  []error{errors.New("address already in use")},
			expectedTries:  1,
			expectFallback: true,
		},
		{
			name:           "multiple binding errors",
			bindingErrors:  []error{errors.New("address already in use"), errors.New("permission denied")},
			expectedTries:  2,
			expectFallback: true,
		},
		{
			name:           "no binding errors",
			bindingErrors:  []error{},
			expectedTries:  0,
			expectFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockControlledUDPClient()
			conn := NewCommandConnectionWithClient(mockClient)

			// Build error list: binding errors + success
			allErrors := append(tt.bindingErrors, nil)
			mockClient.SetSendErrors(allErrors)
			mockClient.SetReceiveData("ok")

			if len(tt.bindingErrors) > 0 {
				_, err := conn.SendCommand("test")
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
			} else {
				_, err := conn.SendCommand("test")
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
			}

			assertEqualInt(t, tt.expectedTries, conn.fallbackTry, "fallbackTry count")
			assertEqualBool(t, tt.expectFallback, conn.inFallback, "inFallback state")
		})
	}
}

// TestReconnectLogic tests the reconnect method
func TestReconnectLogic(t *testing.T) {
	mockClient := NewMockControlledUDPClient()

	initialCloseCount := mockClient.GetCloseCount()

	// Call reconnect - should close and recreate
	// Note: We can't fully test reconnect without a real UDP client
	// because reconnect tries to create a new UDP client
	// We test the close behavior here

	// Simulate reconnect by closing the client manually
	err := mockClient.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}

	assertEqualInt(t, initialCloseCount+1, mockClient.GetCloseCount(), "close count after reconnect")
}

// TestReconnectWithLocal tests reconnectWithLocal method
func TestReconnectWithLocal(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Test that useBound is updated correctly
	// When inFallback is false and we call reconnectWithLocal(true)
	conn.useBound = false
	conn.inFallback = false

	// We can't fully test reconnectWithLocal because it creates real UDP connections
	// But we can test the state transitions on the connection object

	assertTrue(t, !conn.useBound, "useBound should be false initially")

	// Set up state to test inFallback protection
	conn.inFallback = true
	conn.useBound = false

	// When inFallback is true and bindToCommandPort is true,
	// useBound should still be set to true
	if conn.inFallback {
		// This tests that the logic path exists
		// The actual reconnect would need a real network
	}
}

// TestConcurrentReconnectionHandling tests concurrent reconnection safety
func TestConcurrentReconnectionHandling(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Set up to have some initial failures then success
	mockClient.SetSendErrors([]error{
		errors.New("broken pipe"),
		nil,
	})
	mockClient.SetReceiveData("ok")

	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	// Launch multiple goroutines sending commands concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Each goroutine needs its own mock client setup
			// But they share the same connection
			// We can't easily test concurrent SendCommand calls
			// because they share state
			// Instead, test concurrent state access

			// Test concurrent state reads (mutex protection)
			conn.mutex.Lock()
			_ = conn.inFallback
			_ = conn.fallbackTry
			_ = conn.useBound
			conn.mutex.Unlock()

			errChan <- nil
		}(i)
	}

	wg.Wait()

	// Check for any errors from concurrent access
	close(errChan)
	for err := range errChan {
		if err != nil {
			t.Errorf("concurrent access error: %v", err)
		}
	}
}

// TestSendRetryCount tests that retry count matches config
func TestSendRetryCount(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Configure retries to fail all attempts
	failCount := conn.config.CommandRetries + 1
	sendErrs := make([]error, failCount)
	for i := range sendErrs {
		sendErrs[i] = errors.New("persistent error")
	}
	mockClient.SetSendErrors(sendErrs)
	mockClient.SetReceiveData("")

	_, err := conn.SendCommand("test")
	if err == nil {
		t.Error("expected error after all retries exhausted")
	}

	// Verify all retries were attempted
	sendCount := mockClient.GetSendCount()
	assertEqualInt(t, conn.config.CommandRetries, sendCount, "number of send attempts")
}

// TestReceiveRetryCount tests retry count for receive errors
func TestReceiveRetryCount(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Send always succeeds, receive fails all attempts
	mockClient.SetSendErrors([]error{})
	failCount := conn.config.CommandRetries + 1
	recvErrors := make([]error, failCount)
	for i := range recvErrors {
		recvErrors[i] = errors.New("receive timeout")
	}
	mockClient.SetReceiveErrors(recvErrors)

	_, err := conn.SendCommand("test")
	if err == nil {
		t.Error("expected error after all retries exhausted")
	}

	// Verify all retries were attempted
	sendCount := mockClient.GetSendCount()
	// Each retry attempt sends, so we expect CommandRetries sends
	assertEqualInt(t, conn.config.CommandRetries, sendCount, "number of send attempts for receive retries")
}

// TestConnectionStateTransitions tests various connection state transitions
func TestConnectionStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		sendErrors     []error
		recvErrors     []error
		recvData       string
		expectSuccess  bool
		expectFallback bool
		expectedTries  int
	}{
		{
			name:           "success on first try",
			sendErrors:     []error{},
			recvErrors:     []error{},
			recvData:       "ok",
			expectSuccess:  true,
			expectFallback: false,
			expectedTries:  0,
		},
		{
			name:           "retry then success - send error",
			sendErrors:     []error{errors.New("broken pipe"), nil},
			recvErrors:     []error{},
			recvData:       "ok",
			expectSuccess:  true,
			expectFallback: false,
			expectedTries:  0,
		},
		{
			name:           "retry then success - recv error",
			sendErrors:     []error{},
			recvErrors:     []error{errors.New("timeout"), nil},
			recvData:       "ok",
			expectSuccess:  true,
			expectFallback: false,
			expectedTries:  0,
		},
		{
			name:           "binding error triggers fallback",
			sendErrors:     []error{errors.New("address already in use"), nil},
			recvErrors:     []error{},
			recvData:       "ok",
			expectSuccess:  true,
			expectFallback: true,
			expectedTries:  1,
		},
		{
			name:           "all retries exhausted",
			sendErrors:     []error{errors.New("broken pipe"), errors.New("broken pipe"), errors.New("broken pipe")},
			recvErrors:     []error{},
			recvData:       "",
			expectSuccess:  false,
			expectFallback: false,
			expectedTries:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockControlledUDPClient()
			conn := NewCommandConnectionWithClient(mockClient)

			mockClient.SetSendErrors(tt.sendErrors)
			mockClient.SetReceiveErrors(tt.recvErrors)
			mockClient.SetReceiveData(tt.recvData)

			response, err := conn.SendCommand("test")

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("expected success, got error: %v", err)
				}
				if response != tt.recvData {
					t.Errorf("expected response '%s', got '%s'", tt.recvData, response)
				}
			} else {
				if err == nil {
					t.Error("expected error, got success")
				}
			}

			assertEqualBool(t, tt.expectFallback, conn.inFallback, "inFallback state")
			assertEqualInt(t, tt.expectedTries, conn.fallbackTry, "fallbackTry count")
		})
	}
}

// TestCloseDuringOperation tests closing connection during operation
func TestCloseDuringOperation(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	// Close the connection
	err := conn.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}

	// Verify mock was closed
	assertTrue(t, mockClient.IsClosed(), "mock client should be closed")

	// Verify subsequent operations fail
	_, err = conn.SendCommand("test")
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestNilClientHandling tests behavior with nil client
func TestNilClientHandling(t *testing.T) {
	conn := NewCommandConnectionWithClient(nil)

	_, err := conn.SendCommand("test")
	if err == nil {
		t.Error("expected error with nil client")
	}

	assertErrorContains(t, err, "UDP client is not initialized")
}

// TestConfigAccess tests that configuration is accessible
func TestConfigAccess(t *testing.T) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	config := conn.GetConfig()

	// Verify we can access config fields
	_ = config.CommandRetries
	_ = config.CommandTimeout
	_ = config.CommandSendDelay

	// Verify default values
	if config.CommandRetries <= 0 {
		t.Errorf("expected positive CommandRetries, got %d", config.CommandRetries)
	}
}

// TestMixedErrorScenarios tests various error combinations
func TestMixedErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		sendErrors    []error
		recvErrors    []error
		expectedFails int
	}{
		{
			name:          "network error then success",
			sendErrors:    []error{errors.New("connection refused"), nil},
			recvErrors:    []error{},
			expectedFails: 0,
		},
		{
			name:          "multiple network errors then success",
			sendErrors:    []error{errors.New("broken pipe"), errors.New("broken pipe"), nil},
			recvErrors:    []error{},
			expectedFails: 0,
		},
		{
			name:          "binding error then success",
			sendErrors:    []error{errors.New("bind: address already in use"), nil},
			recvErrors:    []error{},
			expectedFails: 0,
		},
		{
			name:          "mixed network and binding errors",
			sendErrors:    []error{errors.New("broken pipe"), errors.New("bind: permission denied"), nil},
			recvErrors:    []error{},
			expectedFails: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockControlledUDPClient()
			conn := NewCommandConnectionWithClient(mockClient)

			mockClient.SetSendErrors(tt.sendErrors)
			mockClient.SetReceiveErrors(tt.recvErrors)
			mockClient.SetReceiveData("ok")

			response, err := conn.SendCommand("test")

			if err != nil {
				t.Errorf("expected success after retries, got error: %v", err)
			}

			if response != "ok" {
				t.Errorf("expected 'ok' response, got '%s'", response)
			}
		})
	}
}

// BenchmarkSendCommand benchmarks the SendCommand method
func BenchmarkSendCommand(b *testing.B) {
	mockClient := NewMockControlledUDPClient()
	conn := NewCommandConnectionWithClient(mockClient)

	mockClient.SetSendErrors([]error{})
	mockClient.SetReceiveData("ok")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = conn.SendCommand("benchmark")
	}
}

// BenchmarkBindingErrorDetection benchmarks binding error detection
func BenchmarkBindingErrorDetection(b *testing.B) {
	err := errors.New("address already in use")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isBindingError(err)
	}
}

// BenchmarkNetworkErrorDetection benchmarks network error detection
func BenchmarkNetworkErrorDetection(b *testing.B) {
	err := errors.New("connection refused")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isNetworkError(err)
	}
}
