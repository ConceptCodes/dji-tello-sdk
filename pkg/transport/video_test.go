package transport

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// MockUDPServerForVideo is a mock implementation for testing VideoStreamListener
type MockUDPServerForVideo struct {
	startCalled bool
	stopCalled  bool
	onDataFunc  func([]byte, *net.UDPAddr)
	onErrorFunc func(error)
	startError  error
}

func (m *MockUDPServerForVideo) Start() error {
	m.startCalled = true
	return m.startError
}

func (m *MockUDPServerForVideo) Stop() {
	m.stopCalled = true
}

func (m *MockUDPServerForVideo) SetOnData(onData func([]byte, *net.UDPAddr)) {
	m.onDataFunc = onData
}

func (m *MockUDPServerForVideo) SetOnError(onError func(error)) {
	m.onErrorFunc = onError
}

func TestNewVideoStreamListener(t *testing.T) {
	// Test with valid address
	listener, err := NewVideoStreamListener("127.0.0.1:11111")
	if err != nil {
		t.Errorf("Expected no error creating video stream listener, got %v", err)
	}
	
	if listener == nil {
		t.Error("Expected listener to be created, got nil")
	}
	
	// Clean up
	listener.Stop()
}

func TestNewVideoStreamListenerInvalidAddress(t *testing.T) {
	// Test with invalid address
	listener, err := NewVideoStreamListener("invalid_address")
	
	if err == nil {
		t.Error("Expected error for invalid address, got nil")
	}
	
	if listener != nil {
		t.Error("Expected listener to be nil for invalid address")
	}
}

func TestVideoStreamListenerStart(t *testing.T) {
	listener, err := NewVideoStreamListener("127.0.0.1:11112")
	if err != nil {
		t.Fatalf("Failed to create video stream listener: %v", err)
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

func TestVideoStreamListenerStop(t *testing.T) {
	listener, err := NewVideoStreamListener("127.0.0.1:11113")
	if err != nil {
		t.Fatalf("Failed to create video stream listener: %v", err)
	}
	
	// Test stopping
	listener.Stop()
	
	// Test stopping again (should not panic)
	listener.Stop()
}

func TestOnVideoStreamData(t *testing.T) {
	// Test onVideoStreamData function with video data
	testAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 11111}
	testData := []byte{0x00, 0x01, 0x02, 0x03} // Mock H.264 data
	
	// Create a video stream listener to test the method
	listener, err := NewVideoStreamListener("127.0.0.1:11114")
	if err != nil {
		t.Fatalf("Failed to create video stream listener: %v", err)
	}
	defer listener.Stop()
	
	// This function is called internally, we just need to ensure it doesn't panic
	// In a real scenario, this would be called by UDP server
	listener.onVideoStreamData(testData, testAddr)
	
	// Test with empty data
	emptyData := []byte{}
	listener.onVideoStreamData(emptyData, testAddr)
	
	// Test with larger data (simulating a video frame)
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	listener.onVideoStreamData(largeData, testAddr)
}

func TestOnVideoStreamError(t *testing.T) {
	// Test onVideoStreamError function
	testError := fmt.Errorf("video stream error")
	
	// This function is called internally, we just need to ensure it doesn't panic
	onVideoStreamError(testError)
	
	// Test with nil error (should not panic)
	onVideoStreamError(nil)
}