package transport

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// VideoFormat represents the video output format
type VideoFormat string

const (
	FormatH264 VideoFormat = "h264"
	FormatMP4  VideoFormat = "mp4"
)

// MP4VideoRecorder handles saving video frames to MP4 files using FFmpeg
type MP4VideoRecorder struct {
	filePath    string
	tempFile    string
	ffmpegCmd   *exec.Cmd
	ffmpegStdin *bufio.Writer
	mutex       sync.Mutex
	isRecording bool
	frameCount  int
	totalBytes  int64
	startTime   time.Time
	width       int
	height      int
	fps         int
	bitrate     int
}

// NewMP4VideoRecorder creates a new MP4 video recorder
func NewMP4VideoRecorder(filePath string) *MP4VideoRecorder {
	return &MP4VideoRecorder{
		filePath: filePath,
		width:    960,     // Tello video width
		height:   720,     // Tello video height
		fps:      30,      // Tello video FPS
		bitrate:  2500000, // 2.5 Mbps bitrate
	}
}

// SetVideoParams sets video parameters for the MP4 output
func (mvr *MP4VideoRecorder) SetVideoParams(width, height, fps, bitrate int) {
	mvr.width = width
	mvr.height = height
	mvr.fps = fps
	mvr.bitrate = bitrate
}

// checkFFmpeg checks if FFmpeg is available
func (mvr *MP4VideoRecorder) checkFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("FFmpeg not found. Please install FFmpeg: %w", err)
	}
	return nil
}

// StartRecording begins saving video frames to MP4 file
func (mvr *MP4VideoRecorder) StartRecording() error {
	mvr.mutex.Lock()
	defer mvr.mutex.Unlock()

	if mvr.isRecording {
		return fmt.Errorf("MP4 recording is already in progress")
	}

	// Check FFmpeg availability
	if err := mvr.checkFFmpeg(); err != nil {
		return err
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(mvr.filePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// FFmpeg command to convert H.264 to MP4
	ffmpegArgs := []string{
		"-f", "h264",
		"-r", fmt.Sprintf("%d", mvr.fps),
		"-i", "pipe:0",
		"-c:v", "copy",
		"-f", "mp4",
		"-movflags", "frag_keyframe+empty_moov",
		"-r", fmt.Sprintf("%d", mvr.fps),
		"-b:v", fmt.Sprintf("%d", mvr.bitrate),
		"-s", fmt.Sprintf("%dx%d", mvr.width, mvr.height),
		mvr.filePath,
	}

	cmd := exec.Command("ffmpeg")
	cmd.Args = append(cmd.Args, ffmpegArgs...)

	// Create pipe for stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start FFmpeg process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	mvr.ffmpegCmd = cmd
	mvr.ffmpegStdin = bufio.NewWriter(stdin)
	mvr.isRecording = true
	mvr.frameCount = 0
	mvr.totalBytes = 0
	mvr.startTime = time.Now()

	utils.Logger.Infof("Started MP4 recording to: %s (%dx%d @ %d fps)", mvr.filePath, mvr.width, mvr.height, mvr.fps)
	return nil
}

// SaveFrame writes a video frame to the MP4 file
func (mvr *MP4VideoRecorder) SaveFrame(frame VideoFrame) error {
	mvr.mutex.Lock()
	defer mvr.mutex.Unlock()

	if !mvr.isRecording {
		return fmt.Errorf("not currently recording MP4")
	}

	if mvr.ffmpegStdin == nil {
		return fmt.Errorf("FFmpeg stdin not available")
	}

	// Write frame data to FFmpeg stdin
	bytesWritten, err := mvr.ffmpegStdin.Write(frame.Data)
	if err != nil {
		return fmt.Errorf("failed to write frame to FFmpeg: %w", err)
	}

	// Flush the buffer to ensure data is sent to FFmpeg
	if err := mvr.ffmpegStdin.Flush(); err != nil {
		utils.Logger.Warnf("Failed to flush FFmpeg stdin: %v", err)
	}

	mvr.frameCount++
	mvr.totalBytes += int64(bytesWritten)

	// Log progress every 100 frames
	if mvr.frameCount%100 == 0 {
		elapsed := time.Since(mvr.startTime)
		fps := float64(mvr.frameCount) / elapsed.Seconds()
		utils.Logger.Debugf("MP4 recording progress: %d frames, %.2f FPS, %.2f MB written",
			mvr.frameCount, fps, float64(mvr.totalBytes)/(1024*1024))
	}

	return nil
}

// StopRecording stops saving video frames and finalizes the MP4 file
func (mvr *MP4VideoRecorder) StopRecording() error {
	mvr.mutex.Lock()
	defer mvr.mutex.Unlock()

	if !mvr.isRecording {
		return fmt.Errorf("no MP4 recording is in progress")
	}

	// Close stdin to signal EOF to FFmpeg
	if mvr.ffmpegStdin != nil {
		if err := mvr.ffmpegStdin.Flush(); err != nil {
			utils.Logger.Warnf("Failed to flush FFmpeg stdin on stop: %v", err)
		}
		mvr.ffmpegStdin = nil
	}

	// Wait for FFmpeg to finish
	if mvr.ffmpegCmd != nil {
		if err := mvr.ffmpegCmd.Wait(); err != nil {
			if !strings.Contains(err.Error(), "exit status") {
				utils.Logger.Errorf("FFmpeg process error: %v", err)
			}
		}
		mvr.ffmpegCmd = nil
	}

	mvr.isRecording = false
	duration := time.Since(mvr.startTime)

	utils.Logger.Infof("Stopped MP4 recording. Frames: %d, Size: %.2f MB, Duration: %v",
		mvr.frameCount, float64(mvr.totalBytes)/(1024*1024), duration)

	return nil
}

// IsRecording returns whether the MP4 recorder is currently recording
func (mvr *MP4VideoRecorder) IsRecording() bool {
	mvr.mutex.Lock()
	defer mvr.mutex.Unlock()
	return mvr.isRecording
}

// GetStats returns recording statistics
func (mvr *MP4VideoRecorder) GetStats() map[string]interface{} {
	mvr.mutex.Lock()
	defer mvr.mutex.Unlock()

	stats := make(map[string]interface{})
	stats["is_recording"] = mvr.isRecording
	stats["frame_count"] = mvr.frameCount
	stats["total_bytes"] = mvr.totalBytes
	stats["file_path"] = mvr.filePath
	stats["format"] = "mp4"
	stats["width"] = mvr.width
	stats["height"] = mvr.height
	stats["fps"] = mvr.fps
	stats["bitrate"] = mvr.bitrate

	if mvr.isRecording && !mvr.startTime.IsZero() {
		stats["duration"] = time.Since(mvr.startTime)
		stats["fps_actual"] = float64(mvr.frameCount) / time.Since(mvr.startTime).Seconds()
	}

	return stats
}

// Close stops recording and cleans up resources
func (mvr *MP4VideoRecorder) Close() error {
	if mvr.IsRecording() {
		if err := mvr.StopRecording(); err != nil {
			return err
		}
	}
	return nil
}
