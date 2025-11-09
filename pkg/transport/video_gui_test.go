package transport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMP4VideoRecorder(t *testing.T) {
	recorder := NewMP4VideoRecorder("test.mp4")

	assert.NotNil(t, recorder)
	assert.Equal(t, "test.mp4", recorder.filePath)
	assert.Equal(t, 960, recorder.width)
	assert.Equal(t, 720, recorder.height)
	assert.Equal(t, 30, recorder.fps)
	assert.Equal(t, 2500000, recorder.bitrate)
}

func TestMP4VideoRecorderSetVideoParams(t *testing.T) {
	recorder := NewMP4VideoRecorder("test.mp4")

	recorder.SetVideoParams(1280, 720, 60, 5000000)

	assert.Equal(t, 1280, recorder.width)
	assert.Equal(t, 720, recorder.height)
	assert.Equal(t, 60, recorder.fps)
	assert.Equal(t, 5000000, recorder.bitrate)
}

func TestMP4VideoRecorderIsRecording(t *testing.T) {
	recorder := NewMP4VideoRecorder("test.mp4")

	// Initially not recording
	assert.False(t, recorder.IsRecording())
}

func TestMP4VideoRecorderGetStats(t *testing.T) {
	recorder := NewMP4VideoRecorder("test.mp4")

	stats := recorder.GetStats()

	assert.Contains(t, stats, "is_recording")
	assert.Contains(t, stats, "frame_count")
	assert.Contains(t, stats, "total_bytes")
	assert.Contains(t, stats, "file_path")
	assert.Contains(t, stats, "format")
	assert.Contains(t, stats, "width")
	assert.Contains(t, stats, "height")
	assert.Contains(t, stats, "fps")
	assert.Contains(t, stats, "bitrate")

	assert.Equal(t, false, stats["is_recording"])
	assert.Equal(t, "test.mp4", stats["file_path"])
	assert.Equal(t, "mp4", stats["format"])
	assert.Equal(t, 960, stats["width"])
	assert.Equal(t, 720, stats["height"])
	assert.Equal(t, 30, stats["fps"])
	assert.Equal(t, 2500000, stats["bitrate"])
}

func TestNewVideoDisplay(t *testing.T) {
	display := NewVideoDisplay(DisplayTypeWeb)

	assert.NotNil(t, display)
	assert.Equal(t, DisplayTypeWeb, display.displayType)
	assert.Equal(t, 8080, display.webPort)
}

func TestVideoDisplaySetWebPort(t *testing.T) {
	display := NewVideoDisplay(DisplayTypeWeb)

	display.SetWebPort(9000)

	assert.Equal(t, 9000, display.webPort)
}

func TestVideoDisplaySetVideoChannel(t *testing.T) {
	display := NewVideoDisplay(DisplayTypeWeb)

	frameChan := make(chan VideoFrame, 10)
	display.SetVideoChannel(frameChan)

	assert.Equal(t, (<-chan VideoFrame)(frameChan), display.frameChan)
}

func TestVideoDisplayIsRunning(t *testing.T) {
	display := NewVideoDisplay(DisplayTypeWeb)

	// Initially not running
	assert.False(t, display.IsRunning())
}

func TestVideoDisplayGetStats(t *testing.T) {
	display := NewVideoDisplay(DisplayTypeTerminal)

	stats := display.GetStats()

	assert.Contains(t, stats, "is_running")
	assert.Contains(t, stats, "frame_count")
	assert.Contains(t, stats, "display_type")
	assert.Contains(t, stats, "web_port")

	assert.Equal(t, false, stats["is_running"])
	assert.Equal(t, "terminal", stats["display_type"])
	assert.Equal(t, 8080, stats["web_port"])
}

func TestVideoRecorderWithFormat(t *testing.T) {
	// Test H.264 format
	recorderH264, err := NewVideoRecorderWithFormat(":11111", "test.h264", FormatH264)
	require.NoError(t, err)
	assert.NotNil(t, recorderH264)
	assert.Equal(t, FormatH264, recorderH264.format)

	// Test MP4 format
	recorderMP4, err := NewVideoRecorderWithFormat(":11111", "test.mp4", FormatMP4)
	require.NoError(t, err)
	assert.NotNil(t, recorderMP4)
	assert.Equal(t, FormatMP4, recorderMP4.format)

	// Test invalid format
	_, err = NewVideoRecorderWithFormat(":11111", "test.xyz", VideoFormat("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported video format")
}

func TestVideoRecorderGetStatsWithFormat(t *testing.T) {
	// Test H.264 recorder stats
	recorderH264, _ := NewVideoRecorderWithFormat(":11111", "test.h264", FormatH264)
	statsH264 := recorderH264.GetStats()
	assert.Equal(t, "h264", statsH264["format"])

	// Test MP4 recorder stats
	recorderMP4, _ := NewVideoRecorderWithFormat(":11111", "test.mp4", FormatMP4)
	statsMP4 := recorderMP4.GetStats()
	assert.Equal(t, "mp4", statsMP4["format"])
}

func TestVideoFrameGUI(t *testing.T) {
	now := time.Now()
	frame := VideoFrame{
		Data:       []byte{0x00, 0x01, 0x02, 0x03},
		Timestamp:  now,
		Size:       4,
		SeqNum:     42,
		IsKeyFrame: true,
	}

	assert.Equal(t, []byte{0x00, 0x01, 0x02, 0x03}, frame.Data)
	assert.Equal(t, now, frame.Timestamp)
	assert.Equal(t, 4, frame.Size)
	assert.Equal(t, 42, frame.SeqNum)
	assert.True(t, frame.IsKeyFrame)
}

func TestVideoFormatConstants(t *testing.T) {
	assert.Equal(t, VideoFormat("h264"), FormatH264)
	assert.Equal(t, VideoFormat("mp4"), FormatMP4)
}

func TestVideoDisplayTypeConstants(t *testing.T) {
	assert.Equal(t, VideoDisplayType("terminal"), DisplayTypeTerminal)
	assert.Equal(t, VideoDisplayType("web"), DisplayTypeWeb)
}
