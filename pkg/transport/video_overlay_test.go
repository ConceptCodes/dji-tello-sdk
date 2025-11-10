package transport

import (
	"image"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/stretchr/testify/assert"
)

func TestVideoDisplayOverlayIntegration(t *testing.T) {
	// Create ML config with overlay enabled
	mlConfig := &ml.MLConfig{
		Overlay: ml.OverlayConfig{
			Enabled:        true,
			ShowFPS:        true,
			ShowDetections: true,
			ShowTracking:   true,
			ShowConfidence: true,
			Colors: map[string]string{
				"person":  "#FF0000",
				"car":     "#0000FF",
				"default": "#00FF00",
			},
			LineWidth: 2,
			FontSize:  12,
			FontScale: 0.5,
		},
	}

	// Create video display
	display := NewVideoDisplay(DisplayTypeWeb)
	display.SetMLConfig(mlConfig)

	// Verify overlay was created
	assert.NotNil(t, display.overlay)

	// Create test ML results
	detectionResult := &ml.DetectionResult{
		Detections: []ml.Detection{
			{
				ClassID:    0,
				ClassName:  "person",
				Confidence: 0.85,
				Box:        image.Rect(100, 100, 200, 200),
			},
		},
		Processor: "yolo-detector",
		Timestamp: time.Now(),
	}

	// Create test image
	testImg := image.NewRGBA(image.Rect(0, 0, 960, 720))

	// Add ML results to display
	display.mutex.Lock()
	display.lastMLResult["yolo-detector"] = detectionResult
	display.mutex.Unlock()

	// Test overlay rendering
	resultImg := display.overlay.Render(testImg, display.lastMLResult, ml.PipelineMetrics{
		FPS: 30.0,
	})

	// Verify result is not nil and has correct dimensions
	assert.NotNil(t, resultImg)
	assert.Equal(t, testImg.Bounds(), resultImg.Bounds())
}

func TestVideoDisplayWithoutOverlay(t *testing.T) {
	// Create ML config with overlay disabled
	mlConfig := &ml.MLConfig{
		Overlay: ml.OverlayConfig{
			Enabled: false,
		},
	}

	// Create video display
	display := NewVideoDisplay(DisplayTypeWeb)
	display.SetMLConfig(mlConfig)

	// Verify overlay was not created
	assert.Nil(t, display.overlay)
}

func TestVideoDisplayMLResultProcessing(t *testing.T) {
	// Create video display
	display := NewVideoDisplay(DisplayTypeWeb)

	// Create video channel (required for display to start)
	videoChan := make(chan VideoFrame, 10)
	display.SetVideoChannel(videoChan)

	// Create ML result channel
	mlResultChan := make(chan ml.MLResult, 10)
	display.SetMLResultChannel(mlResultChan)

	// Create test ML result
	detectionResult := &ml.DetectionResult{
		Detections: []ml.Detection{
			{
				ClassID:    0,
				ClassName:  "person",
				Confidence: 0.85,
				Box:        image.Rect(100, 100, 200, 200),
			},
		},
		Processor: "yolo-detector",
		Timestamp: time.Now(),
	}

	// Start display to enable ML result processing
	display.Start()
	defer display.Close()

	// Send ML result
	mlResultChan <- detectionResult

	// Give more time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify ML result was stored
	display.mutex.Lock()
	storedResult, exists := display.lastMLResult["yolo-detector"]
	display.mutex.Unlock()

	assert.True(t, exists)
	assert.Equal(t, detectionResult, storedResult)

	// Close channel
	close(mlResultChan)
}
