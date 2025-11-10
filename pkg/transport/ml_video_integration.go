package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/pipeline"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// MLVideoIntegration integrates ML processing with the video stream
type MLVideoIntegration struct {
	// Core components
	videoListener *VideoStreamListener
	mlPipeline    *pipeline.ConcurrentMLPipeline

	// Configuration
	mlConfig      *ml.MLConfig
	overlayConfig *ml.OverlayConfig

	// Synchronization
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// State management
	running bool
}

// NewMLVideoIntegration creates a new ML video integration
func NewMLVideoIntegration(listenAddr string, mlConfig *ml.MLConfig) (*MLVideoIntegration, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create video listener
	videoListener, err := NewVideoStreamListener(listenAddr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create video listener: %w", err)
	}

	// Create ML pipeline
	mlPipeline := pipeline.NewConcurrentMLPipeline(&mlConfig.Pipeline, mlConfig.Processors)

	return &MLVideoIntegration{
		videoListener: videoListener,
		mlPipeline:    mlPipeline,
		mlConfig:      mlConfig,
		overlayConfig: &mlConfig.Overlay,
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
	}, nil
}

// Start starts the ML video integration
func (mvi *MLVideoIntegration) Start() error {
	mvi.mu.Lock()
	defer mvi.mu.Unlock()

	if mvi.running {
		return fmt.Errorf("ML video integration is already running")
	}

	// Start ML pipeline
	if err := mvi.mlPipeline.Start(); err != nil {
		return fmt.Errorf("failed to start ML pipeline: %w", err)
	}

	// Start video listener
	if err := mvi.videoListener.Start(); err != nil {
		mvi.mlPipeline.Stop()
		return fmt.Errorf("failed to start video listener: %w", err)
	}

	// Start frame processing goroutine
	mvi.wg.Add(1)
	go mvi.processFrames()

	mvi.running = true
	utils.Logger.Info("ML video integration started successfully")

	return nil
}

// Stop stops the ML video integration
func (mvi *MLVideoIntegration) Stop() error {
	mvi.mu.Lock()
	defer mvi.mu.Unlock()

	if !mvi.running {
		return nil
	}

	// Signal cancellation
	mvi.cancel()

	// Stop video listener
	mvi.videoListener.Stop()

	// Stop ML pipeline
	if err := mvi.mlPipeline.Stop(); err != nil {
		utils.Logger.Errorf("Error stopping ML pipeline: %v", err)
	}

	// Wait for all goroutines to finish
	mvi.wg.Wait()

	mvi.running = false
	utils.Logger.Info("ML video integration stopped")

	return nil
}

// processFrames processes video frames through the ML pipeline
func (mvi *MLVideoIntegration) processFrames() {
	defer mvi.wg.Done()

	frameChan := mvi.videoListener.GetFrameChannel()

	for {
		select {
		case <-mvi.ctx.Done():
			return

		case frame, ok := <-frameChan:
			if !ok {
				utils.Logger.Info("Video frame channel closed")
				return
			}

			// Convert to enhanced frame for ML processing
			enhancedFrame := frame.ToEnhancedFrame()

			// Send to ML pipeline
			if err := mvi.mlPipeline.ProcessFrame(enhancedFrame); err != nil {
				utils.Logger.Debugf("Failed to process frame in ML pipeline: %v", err)
			}

		case result := <-mvi.mlPipeline.GetResults():
			// Handle ML results
			mvi.handleMLResult(result)
		}
	}
}

// handleMLResult handles ML processing results
func (mvi *MLVideoIntegration) handleMLResult(result ml.MLResult) {
	utils.Logger.Debugf("Received ML result from %s processor", result.GetProcessorName())

	// Process ML results based on type and configuration
	switch r := result.(type) {
	case *ml.DetectionResult:
		utils.Logger.Debugf("Detection result: %d objects detected", len(r.Detections))
		for _, detection := range r.Detections {
			utils.Logger.Debugf("  - %s: %.2f confidence at %v",
				detection.ClassName, detection.Confidence, detection.Box)
		}

	case *ml.GestureResult:
		utils.Logger.Debugf("Gesture result: %s with %.2f confidence",
			r.Gesture, r.Confidence)

	case *ml.SLAMResult:
		if r.Pose != nil {
			utils.Logger.Debugf("SLAM pose: X=%.2f, Y=%.2f, Z=%.2f",
				r.Pose.Position.X, r.Pose.Position.Y, r.Pose.Position.Z)
		}

	default:
		utils.Logger.Debugf("Unknown ML result type: %T", result)
	}
}

// GetMetrics returns current ML pipeline metrics
func (mvi *MLVideoIntegration) GetMetrics() ml.PipelineMetrics {
	return mvi.mlPipeline.GetMetrics()
}

// IsRunning returns whether the integration is currently running
func (mvi *MLVideoIntegration) IsRunning() bool {
	mvi.mu.RLock()
	defer mvi.mu.RUnlock()
	return mvi.running
}

// GetFrameChannel returns the video frame channel for external consumers
func (mvi *MLVideoIntegration) GetFrameChannel() <-chan VideoFrame {
	return mvi.videoListener.GetFrameChannel()
}

// GetMLResults returns the ML results channel for external consumers
func (mvi *MLVideoIntegration) GetMLResults() <-chan ml.MLResult {
	return mvi.mlPipeline.GetResults()
}

// CreateIntegratedVideoDisplay creates a video display with ML overlay integration
func (mvi *MLVideoIntegration) CreateIntegratedVideoDisplay(displayType VideoDisplayType) *VideoDisplay {
	display := NewVideoDisplay(displayType)

	// Set video channel
	display.SetVideoChannel(mvi.GetFrameChannel())

	// Set ML result channel for overlay
	display.SetMLResultChannel(mvi.GetMLResults())

	// Set ML configuration for overlay rendering
	display.SetMLConfig(mvi.mlConfig)

	return display
}
