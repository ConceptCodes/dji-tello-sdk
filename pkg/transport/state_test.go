package transport

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// MockUDPServerForState is a mock implementation for testing StateListener
type MockUDPServerForState struct {
	startCalled bool
	stopCalled  bool
	onDataFunc  func([]byte, *net.UDPAddr)
	onErrorFunc func(error)
	startError  error
}

func (m *MockUDPServerForState) Start() error {
	m.startCalled = true
	return m.startError
}

func (m *MockUDPServerForState) Stop() {
	m.stopCalled = true
}

func (m *MockUDPServerForState) SetOnData(onData func([]byte, *net.UDPAddr)) {
	m.onDataFunc = onData
}

func (m *MockUDPServerForState) SetOnError(onError func(error)) {
	m.onErrorFunc = onError
}

func TestNewStateListener(t *testing.T) {
	// Test with valid address
	listener, err := NewStateListener("127.0.0.1:8890")
	if err != nil {
		t.Errorf("Expected no error creating state listener, got %v", err)
	}

	if listener == nil {
		t.Error("Expected listener to be created, got nil")
	}

	// Clean up
	listener.Stop()
}

func TestNewStateListenerInvalidAddress(t *testing.T) {
	// Test with invalid address
	listener, err := NewStateListener("invalid_address")

	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}

	if listener != nil {
		t.Error("Expected listener to be nil for invalid address")
	}
}

func TestStateListenerStart(t *testing.T) {
	listener, err := NewStateListener("127.0.0.1:8891")
	if err != nil {
		t.Fatalf("Failed to create state listener: %v", err)
	}

	// Test starting - this will block, so run in goroutine
	startDone := make(chan error, 1)
	go func() {
		startDone <- listener.Start()
	}()

	// Wait a bit for start to be called
	time.Sleep(100 * time.Millisecond)

	// Stop the listener to unblock Start()
	listener.Stop()

	// Check if Start() returned
	select {
	case err := <-startDone:
		// Start() should return without error after Stop()
		if err != nil {
			t.Errorf("Expected no error from Start(), got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start() did not return after Stop()")
	}
}

func TestStateListenerStop(t *testing.T) {
	listener, err := NewStateListener("127.0.0.1:8892")
	if err != nil {
		t.Fatalf("Failed to create state listener: %v", err)
	}

	// Test stopping
	listener.Stop()

	// Test stopping again (should not panic)
	listener.Stop()
}

func TestStateListenerDataProcessing(t *testing.T) {
	// Test that StateListener can process data without panicking
	listener, err := NewStateListener("127.0.0.1:0") // Use port 0 for automatic assignment
	if err != nil {
		t.Fatalf("Failed to create state listener: %v", err)
	}
	defer listener.Stop()

	// Test data processing through the listener's internal methods
	// We can't directly test the private methods, but we can verify
	// the listener is created and can be started/stopped without errors
}

func TestOnStateError(t *testing.T) {
	// Test the onStateError function
	testError := fmt.Errorf("test error")

	// This function is called internally, we just need to ensure it doesn't panic
	onStateError(testError)
}
