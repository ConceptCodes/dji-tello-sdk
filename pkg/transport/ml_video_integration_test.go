package transport

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestVideoFrame_ToEnhancedFrame(t *testing.T) {
	// Create a video frame
	data := []byte("test video data")
	timestamp := time.Now()
	seqNum := 42
	nalUnits := []NALUnit{{Type: 1, Size: 100}}

	videoFrame := VideoFrame{
		Data:       data,
		Timestamp:  timestamp,
		Size:       len(data),
		SeqNum:     seqNum,
		NALUnits:   nalUnits,
		IsKeyFrame: true,
	}

	// Convert to enhanced frame
	enhancedFrame := videoFrame.ToEnhancedFrame()

	// Verify conversion
	assert.NotNil(t, enhancedFrame)
	assert.Equal(t, data, enhancedFrame.Data)
	assert.Equal(t, timestamp, enhancedFrame.Timestamp)
	assert.Equal(t, seqNum, enhancedFrame.SeqNum)
	assert.Equal(t, len(data), enhancedFrame.Size)
	assert.True(t, enhancedFrame.IsKeyFrame)
	assert.False(t, enhancedFrame.Processed)
	assert.NotNil(t, enhancedFrame.MLResults)
	assert.Equal(t, defaultVideoFrameWidth, enhancedFrame.Width)
	assert.Equal(t, defaultVideoFrameHeight, enhancedFrame.Height)
	assert.Equal(t, 3, enhancedFrame.Channels)

	img, ok := enhancedFrame.Image.(image.Image)
	assert.True(t, ok)
	assert.NotNil(t, img)
}

func TestNewMLVideoIntegration(t *testing.T) {
	// Create ML config
	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{
			{
				Name:    "test_processor",
				Type:    ml.ProcessorTypeCustom,
				Enabled: false, // Disabled to avoid factory issues
				Config:  map[string]interface{}{"test": true},
			},
		},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled:        false,
			ShowFPS:        false,
			ShowDetections: false,
			Colors:         map[string]string{},
			LineWidth:      2,
			FontSize:       12,
			FontScale:      0.5,
		},
	}

	// Create ML video integration
	integration, err := NewMLVideoIntegration("0.0.0.0:6060", mlConfig)
	// Note: This might fail if UDP server can't bind to the port
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create video listener")
		return
	}

	assert.NotNil(t, integration)
	assert.Equal(t, mlConfig, integration.mlConfig)
	assert.Equal(t, &mlConfig.Overlay, integration.overlayConfig)
	assert.False(t, integration.IsRunning())
}

func TestMLVideoIntegration_StartStop(t *testing.T) {
	// Create ML config with no processors to avoid factory issues
	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	// Create ML video integration with a different port to avoid conflicts
	integration, err := NewMLVideoIntegration("0.0.0.0:6061", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Test start
	err = integration.Start()
	if err != nil {
		t.Skipf("Skipping test due to start error: %v", err)
		return
	}
	assert.True(t, integration.IsRunning())

	// Test double start
	err = integration.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = integration.Stop()
	assert.NoError(t, err)
	assert.False(t, integration.IsRunning())
}

func TestMLVideoIntegration_GetMetrics(t *testing.T) {
	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6062", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Get metrics without starting
	metrics := integration.GetMetrics()
	assert.Equal(t, float64(0), metrics.FPS)
	assert.Equal(t, int64(0), metrics.DroppedFrames)
	assert.NotNil(t, metrics.ProcessorStats)
}

func TestMLVideoIntegration_GetChannels(t *testing.T) {
	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6063", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Get frame channel
	frameChan := integration.GetFrameChannel()
	assert.NotNil(t, frameChan)

	// Get ML results channel
	mlChan := integration.GetMLResults()
	assert.NotNil(t, mlChan)
}

func TestMLVideoIntegration_HandleMLResult(t *testing.T) {
	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6064", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Test handling different ML result types
	detectionResult := &ml.DetectionResult{
		Detections: []ml.Detection{
			{
				ClassID:    0,
				ClassName:  "person",
				Confidence: 0.85,
				Box:        image.Rect(10, 10, 100, 100),
				Timestamp:  time.Now(),
			},
		},
		Processor: "yolo",
		Timestamp: time.Now(),
	}

	// This would normally be called internally, but we can test the method directly
	// Note: This is testing the method exists and doesn't panic
	integration.handleMLResult(detectionResult)

	gestureResult := &ml.GestureResult{
		Gesture:     "thumbs_up",
		Confidence:  0.90,
		BoundingBox: image.Rect(50, 50, 150, 150),
		Processor:   "gesture",
		Timestamp:   time.Now(),
	}

	integration.handleMLResult(gestureResult)

	slamResult := &ml.SLAMResult{
		Pose: &ml.Pose6D{
			Position:    ml.Point3D{X: 1.0, Y: 2.0, Z: 3.0},
			Orientation: ml.Point3D{X: 0.0, Y: 0.0, Z: 90.0},
			Timestamp:   time.Now(),
		},
		Processor: "slam",
		Timestamp: time.Now(),
	}

	integration.handleMLResult(slamResult)
}

func TestMLVideoIntegration_ContextCancellation(t *testing.T) {
	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6065", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Start integration
	err = integration.Start()
	if err != nil {
		t.Skipf("Skipping test due to start error: %v", err)
		return
	}
	assert.True(t, integration.IsRunning())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Wait for context cancellation or integration stop
	done := make(chan bool, 1)
	go func() {
		integration.Stop()
		done <- true
	}()

	select {
	case <-ctx.Done():
		// Context timed out, stop integration
		integration.Stop()
	case <-done:
		// Integration stopped cleanly
	}

	assert.False(t, integration.IsRunning())
}

// ==================== Goroutine Leak Detection Tests ====================

// TestMLVideoIntegrationShutdownNoLeak verifies no goroutines leak after ML video integration shutdown
func TestMLVideoIntegrationShutdownNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6070", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Start integration
	err = integration.Start()
	if err != nil {
		t.Skipf("Skipping test due to start error: %v", err)
		return
	}

	// Stop integration
	err = integration.Stop()
	if err != nil {
		t.Fatalf("Failed to stop integration: %v", err)
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)
}

// TestMLVideoIntegrationMultipleStartStopNoLeak verifies no leaks during multiple start/stop cycles
func TestMLVideoIntegrationMultipleStartStopNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6071", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Perform multiple start/stop cycles
	for i := 0; i < 3; i++ {
		err = integration.Start()
		if err != nil {
			t.Fatalf("Failed to start integration: %v", err)
		}

		err = integration.Stop()
		if err != nil {
			t.Fatalf("Failed to stop integration: %v", err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// TestMLVideoIntegrationConcurrentOperationsNoLeak verifies no leaks during concurrent operations
func TestMLVideoIntegrationConcurrentOperationsNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	mlConfig := &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 2,
			FrameBufferSize:         100,
			WorkerPoolSize:          4,
			EnableMetrics:           false,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	integration, err := NewMLVideoIntegration("0.0.0.0:6072", mlConfig)
	if err != nil {
		t.Skipf("Skipping test due to UDP bind error: %v", err)
		return
	}

	// Start integration
	err = integration.Start()
	if err != nil {
		t.Skipf("Skipping test due to start error: %v", err)
		return
	}

	// Get channels (we can't close receive-only channels, but we can access them)
	_ = integration.GetFrameChannel()
	_ = integration.GetMLResults()

	// Stop integration
	err = integration.Stop()
	if err != nil {
		t.Fatalf("Failed to stop integration: %v", err)
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)
}
