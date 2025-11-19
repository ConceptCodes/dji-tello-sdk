package transport

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// VideoSaver handles saving video frames to files
type VideoSaver struct {
	filePath    string
	file        *os.File
	mutex       sync.Mutex
	isRecording bool
	frameCount  int
	totalBytes  int64
	startTime   time.Time
}

// NewVideoSaver creates a new video saver
func NewVideoSaver(filePath string) *VideoSaver {
	return &VideoSaver{
		filePath: filePath,
	}
}

// StartRecording begins saving video frames to the specified file
func (vs *VideoSaver) StartRecording() error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	if vs.isRecording {
		return fmt.Errorf("recording is already in progress")
	}

	file, err := os.Create(vs.filePath)
	if err != nil {
		return fmt.Errorf("failed to create video file '%s': %w", vs.filePath, err)
	}

	vs.file = file
	vs.isRecording = true
	vs.frameCount = 0
	vs.totalBytes = 0
	vs.startTime = time.Now()

	utils.Logger.Infof("Started recording video to: %s", vs.filePath)
	return nil
}

// StopRecording stops saving video frames and closes the file
func (vs *VideoSaver) StopRecording() error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	if !vs.isRecording {
		return fmt.Errorf("no recording is in progress")
	}

	if vs.file != nil {
		err := vs.file.Close()
		vs.file = nil
		if err != nil {
			return fmt.Errorf("failed to close video file: %w", err)
		}
	}

	vs.isRecording = false
	duration := time.Since(vs.startTime)

	utils.Logger.Infof("Stopped recording. Frames: %d, Size: %.2f MB, Duration: %v",
		vs.frameCount, float64(vs.totalBytes)/(1024*1024), duration)

	return nil
}

// SaveFrame writes a video frame to the file
func (vs *VideoSaver) SaveFrame(frame VideoFrame) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	if !vs.isRecording {
		return fmt.Errorf("not currently recording")
	}

	if vs.file == nil {
		return fmt.Errorf("video file is not open")
	}

	// Write frame data to file
	bytesWritten, err := vs.file.Write(frame.Data)
	if err != nil {
		return fmt.Errorf("failed to write frame to file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := vs.file.Sync(); err != nil {
		utils.Logger.Warnf("Failed to sync video file: %v", err)
	}

	vs.frameCount++
	vs.totalBytes += int64(bytesWritten)

	// Log progress every 100 frames
	if vs.frameCount%100 == 0 {
		elapsed := time.Since(vs.startTime)
		fps := float64(vs.frameCount) / elapsed.Seconds()
		utils.Logger.Debugf("Recording progress: %d frames, %.2f FPS, %.2f MB written",
			vs.frameCount, fps, float64(vs.totalBytes)/(1024*1024))
	}

	return nil
}

// IsRecording returns whether the saver is currently recording
func (vs *VideoSaver) IsRecording() bool {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	return vs.isRecording
}

// GetStats returns recording statistics
func (vs *VideoSaver) GetStats() map[string]interface{} {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	stats := make(map[string]interface{})
	stats["is_recording"] = vs.isRecording
	stats["frame_count"] = vs.frameCount
	stats["total_bytes"] = vs.totalBytes
	stats["file_path"] = vs.filePath

	if vs.isRecording && !vs.startTime.IsZero() {
		stats["duration"] = time.Since(vs.startTime)
		stats["fps"] = float64(vs.frameCount) / time.Since(vs.startTime).Seconds()
	}

	return stats
}

// Close stops recording and cleans up resources
func (vs *VideoSaver) Close() error {
	if vs.IsRecording() {
		if err := vs.StopRecording(); err != nil {
			return err
		}
	}
	return nil
}

// VideoRecorder combines video stream listening and saving
type VideoRecorder struct {
	listener    *VideoStreamListener
	saver       *VideoSaver
	mp4Recorder *MP4VideoRecorder
	frameChan   <-chan VideoFrame
	stopChan    chan bool
	isRunning   bool
	mutex       sync.Mutex
	format      VideoFormat
}

// NewVideoRecorder creates a new video recorder
func NewVideoRecorder(listenAddr, savePath string) (*VideoRecorder, error) {
	return NewVideoRecorderWithFormat(listenAddr, savePath, FormatH264)
}

// NewVideoRecorderFromChannel creates a video recorder from existing frame channel
func NewVideoRecorderFromChannel(frameChan <-chan VideoFrame, savePath string) (*VideoRecorder, error) {
	return NewVideoRecorderWithFormatAndChannel(frameChan, savePath, FormatH264)
}

// NewVideoRecorderWithFormatAndChannel creates a video recorder with specified format and existing frame channel
func NewVideoRecorderWithFormatAndChannel(frameChan <-chan VideoFrame, savePath string, format VideoFormat) (*VideoRecorder, error) {
	recorder := &VideoRecorder{
		listener:  nil, // No listener needed when using existing channel
		frameChan: frameChan,
		stopChan:  make(chan bool),
		isRunning: false,
		format:    format,
	}

	// Initialize appropriate saver based on format
	switch format {
	case FormatH264:
		recorder.saver = NewVideoSaver(savePath)
	case FormatMP4:
		recorder.mp4Recorder = NewMP4VideoRecorder(savePath)
	default:
		return nil, fmt.Errorf("unsupported video format: %s", format)
	}

	return recorder, nil
}

// NewVideoRecorderWithFormat creates a new video recorder with specified format
func NewVideoRecorderWithFormat(listenAddr, savePath string, format VideoFormat) (*VideoRecorder, error) {
	listener, err := NewVideoStreamListener(listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create video stream listener: %w", err)
	}

	recorder := &VideoRecorder{
		listener:  listener,
		frameChan: listener.GetFrameChannel(),
		stopChan:  make(chan bool),
		isRunning: false,
		format:    format,
	}

	// Initialize appropriate saver based on format
	switch format {
	case FormatH264:
		recorder.saver = NewVideoSaver(savePath)
	case FormatMP4:
		recorder.mp4Recorder = NewMP4VideoRecorder(savePath)
	default:
		return nil, fmt.Errorf("unsupported video format: %s", format)
	}

	return recorder, nil
}

// StartRecording starts both the video listener and recording
func (vr *VideoRecorder) StartRecording() error {
	vr.mutex.Lock()
	defer vr.mutex.Unlock()

	if vr.isRunning {
		return fmt.Errorf("video recorder is already running")
	}

	// Start the video listener only if we have one
	if vr.listener != nil {
		go func() {
			if err := vr.listener.Start(); err != nil {
				utils.Logger.Errorf("Video listener error: %v", err)
			}
		}()
	}

	// Start recording based on format
	var err error
	switch vr.format {
	case FormatH264:
		err = vr.saver.StartRecording()
	case FormatMP4:
		err = vr.mp4Recorder.StartRecording()
	default:
		return fmt.Errorf("unsupported video format: %s", vr.format)
	}

	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Start frame processing goroutine
	go vr.processFrames()

	vr.isRunning = true
	utils.Logger.Infof("Video recorder started (format: %s)", vr.format)
	return nil
}

// StopRecording stops the video recorder
func (vr *VideoRecorder) StopRecording() error {
	vr.mutex.Lock()
	defer vr.mutex.Unlock()

	if !vr.isRunning {
		return fmt.Errorf("video recorder is not running")
	}

	// Signal stop
	close(vr.stopChan)

	// Stop recording based on format
	switch vr.format {
	case FormatH264:
		if err := vr.saver.StopRecording(); err != nil {
			utils.Logger.Errorf("Failed to stop recording: %v", err)
		}
	case FormatMP4:
		if err := vr.mp4Recorder.StopRecording(); err != nil {
			utils.Logger.Errorf("Failed to stop recording: %v", err)
		}
	}

	// Stop listener
	vr.listener.Stop()

	vr.isRunning = false
	vr.stopChan = make(chan bool) // Recreate stop channel for next use

	utils.Logger.Info("Video recorder stopped")
	return nil
}

// processFrames processes incoming video frames and saves them
func (vr *VideoRecorder) processFrames() {
	for {
		select {
		case frame, ok := <-vr.frameChan:
			if !ok {
				utils.Logger.Info("Video frame channel closed")
				return
			}

			var err error
			switch vr.format {
			case FormatH264:
				err = vr.saver.SaveFrame(frame)
			case FormatMP4:
				err = vr.mp4Recorder.SaveFrame(frame)
			default:
				utils.Logger.Errorf("Unsupported video format: %s", vr.format)
				continue
			}

			if err != nil {
				utils.Logger.Errorf("Failed to save frame: %v", err)
			}

		case <-vr.stopChan:
			utils.Logger.Info("Received stop signal, stopping frame processing")
			return
		}
	}
}

// IsRecording returns whether the recorder is currently recording to file
func (vr *VideoRecorder) IsRecording() bool {
	vr.mutex.Lock()
	defer vr.mutex.Unlock()

	if !vr.isRunning {
		return false
	}

	switch vr.format {
	case FormatH264:
		if vr.saver != nil {
			return vr.saver.IsRecording()
		}
	case FormatMP4:
		if vr.mp4Recorder != nil {
			return vr.mp4Recorder.IsRecording()
		}
	}
	return false
}

// GetStats returns combined statistics from listener and saver
func (vr *VideoRecorder) GetStats() map[string]interface{} {
	var stats map[string]interface{}

	switch vr.format {
	case FormatH264:
		stats = vr.saver.GetStats()
	case FormatMP4:
		stats = vr.mp4Recorder.GetStats()
	default:
		stats = make(map[string]interface{})
	}

	stats["is_running"] = vr.isRunning
	stats["format"] = string(vr.format)
	return stats
}

// Close stops recording and cleans up resources
func (vr *VideoRecorder) Close() error {
	if vr.isRunning {
		if err := vr.StopRecording(); err != nil {
			return err
		}
	}

	// Close format-specific resources
	switch vr.format {
	case FormatH264:
		if vr.saver != nil {
			vr.saver.Close()
		}
	case FormatMP4:
		if vr.mp4Recorder != nil {
			vr.mp4Recorder.Close()
		}
	}

	return nil
}

// NewVideoRecorderMP4 creates a new MP4 video recorder with listener
func NewVideoRecorderMP4(listenAddr, savePath string) (*VideoRecorder, error) {
	return NewVideoRecorderWithFormat(listenAddr, savePath, FormatMP4)
}
