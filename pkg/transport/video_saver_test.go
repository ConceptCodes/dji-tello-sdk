package transport

import (
	"os"
	"testing"
	"time"
)

func TestVideoSaver(t *testing.T) {
	// Create a temporary file for testing
	tempFile := "test_video.h264"
	defer os.Remove(tempFile)

	saver := NewVideoSaver(tempFile)

	// Test starting recording
	if err := saver.StartRecording(); err != nil {
		t.Errorf("Failed to start recording: %v", err)
	}

	// Test is recording
	if !saver.IsRecording() {
		t.Error("Expected IsRecording to return true")
	}

	// Create test video frame
	testFrame := VideoFrame{
		Data:      []byte{0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1E}, // Mock H.264 data
		Timestamp: time.Now(),
		Size:      7,
		SeqNum:    1,
	}

	// Test saving frame
	if err := saver.SaveFrame(testFrame); err != nil {
		t.Errorf("Failed to save frame: %v", err)
	}

	// Test getting stats
	stats := saver.GetStats()
	if stats["frame_count"] != 1 {
		t.Errorf("Expected frame count 1, got %v", stats["frame_count"])
	}
	if stats["is_recording"] != true {
		t.Error("Expected is_recording to be true")
	}

	// Test stopping recording
	if err := saver.StopRecording(); err != nil {
		t.Errorf("Failed to stop recording: %v", err)
	}

	// Test is not recording
	if saver.IsRecording() {
		t.Error("Expected IsRecording to return false")
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Expected video file to be created")
	}

	// Test cleanup
	if err := saver.Close(); err != nil {
		t.Errorf("Failed to close saver: %v", err)
	}
}

func TestVideoSaverDoubleStart(t *testing.T) {
	tempFile := "test_video_double.h264"
	defer os.Remove(tempFile)

	saver := NewVideoSaver(tempFile)

	// Start recording
	if err := saver.StartRecording(); err != nil {
		t.Errorf("Failed to start recording: %v", err)
	}

	// Try to start again (should fail)
	if err := saver.StartRecording(); err == nil {
		t.Error("Expected error when starting recording twice")
	}

	// Cleanup
	saver.StopRecording()
	saver.Close()
}

func TestVideoSaverStopWithoutStart(t *testing.T) {
	tempFile := "test_video_stop.h264"
	defer os.Remove(tempFile)

	saver := NewVideoSaver(tempFile)

	// Try to stop without starting (should fail)
	if err := saver.StopRecording(); err == nil {
		t.Error("Expected error when stopping recording without starting")
	}
}

func TestVideoRecorder(t *testing.T) {
	tempFile := "test_recorder.h264"
	defer os.Remove(tempFile)

	recorder, err := NewVideoRecorder(":11115", tempFile)
	if err != nil {
		t.Errorf("Failed to create video recorder: %v", err)
	}
	defer recorder.Close()

	// Test getting stats before starting
	stats := recorder.GetStats()
	if stats["is_running"] != false {
		t.Error("Expected is_running to be false initially")
	}

	// Note: We can't fully test the recording functionality without
	// actual video data, but we can test the setup and stats
	if recorder.saver == nil {
		t.Error("Expected saver to be initialized")
	}

	if recorder.listener == nil {
		t.Error("Expected listener to be initialized")
	}
}

func TestVideoRecorderStats(t *testing.T) {
	tempFile := "test_recorder_stats.h264"
	defer os.Remove(tempFile)

	recorder, err := NewVideoRecorder(":11116", tempFile)
	if err != nil {
		t.Errorf("Failed to create video recorder: %v", err)
	}
	defer recorder.Close()

	stats := recorder.GetStats()

	// Check required stats fields
	expectedFields := []string{"is_running", "is_recording", "frame_count", "total_bytes", "file_path"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected stats to contain field '%s'", field)
		}
	}

	if stats["file_path"] != tempFile {
		t.Errorf("Expected file_path to be '%s', got '%v'", tempFile, stats["file_path"])
	}
}
