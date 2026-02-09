package transport

import (
	"fmt"
	"net"
	"reflect"
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

func TestVideoStreamListener_Close(t *testing.T) {
	listener, err := NewVideoStreamListener("127.0.0.1:11145")
	if err != nil {
		t.Fatalf("Failed to create video stream listener: %v", err)
	}

	// Test Close method
	err = listener.Close()
	if err != nil {
		t.Errorf("Expected no error from Close(), got %v", err)
	}

	// Verify that Stop was called (FrameChan should be nil after Close)
	if listener.FrameChan != nil {
		t.Error("Expected FrameChan to be nil after Close()")
	}

	// Test closing again (should not panic)
	err = listener.Close()
	if err != nil {
		t.Errorf("Expected no error from second Close(), got %v", err)
	}
}

func TestVideoStreamListener_GetFrameChannel(t *testing.T) {
	listener, err := NewVideoStreamListener("127.0.0.1:11146")
	if err != nil {
		t.Fatalf("Failed to create video stream listener: %v", err)
	}
	defer listener.Stop()

	// Test GetFrameChannel returns a read-only channel
	frameChan := listener.GetFrameChannel()

	// Verify it's a read-only channel by checking the type
	if reflect.TypeOf(frameChan).String() != "<-chan transport.VideoFrame" {
		t.Errorf("Expected read-only channel type, got %s", reflect.TypeOf(frameChan).String())
	}

	// Test that it returns the same channel instance
	frameChan2 := listener.GetFrameChannel()
	if frameChan != frameChan2 {
		t.Error("Expected GetFrameChannel to return the same channel instance")
	}

	// Test that we can receive from the channel (it's not nil)
	select {
	case <-frameChan:
		// Channel is readable, good
	default:
		// Channel is empty but readable, also good
	}
}

func TestVideoFrame_toRGBImage(t *testing.T) {
	// Test with empty data - should return image with base color
	vf := &VideoFrame{
		Data:       []byte{},
		IsKeyFrame: false,
	}
	img := vf.toRGBImage()
	if img == nil {
		t.Fatal("Expected image to be created")
	}

	// Check dimensions
	bounds := img.Bounds()
	if bounds.Dx() != defaultVideoFrameWidth || bounds.Dy() != defaultVideoFrameHeight {
		t.Errorf("Expected dimensions %dx%d, got %dx%d", defaultVideoFrameWidth, defaultVideoFrameHeight, bounds.Dx(), bounds.Dy())
	}

	// Check base color for non-keyframe (should be gray: R=30, G=30, B=30)
	r, g, b, _ := img.At(0, 0).RGBA()
	if r>>8 != 30 || g>>8 != 30 || b>>8 != 30 {
		t.Errorf("Expected base color (30,30,30) for non-keyframe, got (%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Test with keyframe - should have different base color
	vf.IsKeyFrame = true
	imgKeyframe := vf.toRGBImage()
	r2, g2, b2, _ := imgKeyframe.At(0, 0).RGBA()
	if r2>>8 != 30 || g2>>8 != 90 || b2>>8 != 160 {
		t.Errorf("Expected base color (30,90,160) for keyframe, got (%d,%d,%d)", r2>>8, g2>>8, b2>>8)
	}

	// Test with data - should extract RGB values and fill blocks
	// Create data that will give predictable colors: R=255, G=128, B=64, repeated
	testData := []byte{255, 128, 64, 255, 128, 64} // Two sets of RGB
	vf = &VideoFrame{
		Data:       testData,
		IsKeyFrame: false,
	}
	imgWithData := vf.toRGBImage()

	// Check that first block (0,0 to 7,7) has the expected color
	r3, g3, b3, _ := imgWithData.At(0, 0).RGBA()
	if r3>>8 != 255 || g3>>8 != 128 || b3>>8 != 64 {
		t.Errorf("Expected color (255,128,64) in first block, got (%d,%d,%d)", r3>>8, g3>>8, b3>>8)
	}

	// Check that second block (8,0 to 15,7) also has the same color (data cycles)
	r4, g4, b4, _ := imgWithData.At(8, 0).RGBA()
	if r4>>8 != 255 || g4>>8 != 128 || b4>>8 != 64 {
		t.Errorf("Expected color (255,128,64) in second block, got (%d,%d,%d)", r4>>8, g4>>8, b4>>8)
	}

	// Check that a pixel in the middle of a block has the same color
	r5, g5, b5, _ := imgWithData.At(4, 4).RGBA()
	if r5>>8 != 255 || g5>>8 != 128 || b5>>8 != 64 {
		t.Errorf("Expected color (255,128,64) in middle of block, got (%d,%d,%d)", r5>>8, g5>>8, b5>>8)
	}
}
