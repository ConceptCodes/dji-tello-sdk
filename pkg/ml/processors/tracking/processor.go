package tracking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/tracking"
)

// TrackingProcessor implements object tracking
type TrackingProcessor struct {
	*processors.BaseProcessor
	tracker *tracking.ObjectTracker
	config  *TrackingConfig
	mu      sync.Mutex
}

// TrackingConfig defines configuration for Tracking processor
type TrackingConfig struct {
	InputProcessor string                 `json:"input_processor"` // Name of processor to get detections from
	TrackerConfig  tracking.TrackerConfig `json:"tracker_config"`
}

// NewTrackingProcessor creates a new Tracking processor
func NewTrackingProcessor(name string) *TrackingProcessor {
	return &TrackingProcessor{
		BaseProcessor: processors.NewBaseProcessor(name, ml.ProcessorTypeTracking),
		config: &TrackingConfig{
			InputProcessor: "yolo", // Default to yolo
			TrackerConfig: tracking.TrackerConfig{
				MaxDistance:      50.0,
				MaxAge:           30,
				MinHits:          3,
				MaxIOUDistance:   0.7,
				UseKalmanFilter:  true,
				EnablePrediction: true,
			},
		},
	}
}

// Configure configures the Tracking processor
func (tp *TrackingProcessor) Configure(config map[string]interface{}) error {
	if err := tp.BaseProcessor.Configure(config); err != nil {
		return err
	}

	// Parse configuration
	if inputProc, ok := config["input_processor"].(string); ok {
		tp.config.InputProcessor = inputProc
	}

	if trackerConfig, ok := config["tracker_config"].(map[string]interface{}); ok {
		// Parse tracker config manually or use mapstructure/json
		// For simplicity, we'll just check a few fields
		if maxDist, ok := trackerConfig["max_distance"].(float64); ok {
			tp.config.TrackerConfig.MaxDistance = maxDist
		}
		if maxAge, ok := trackerConfig["max_age"].(float64); ok {
			tp.config.TrackerConfig.MaxAge = int(maxAge)
		}
		if minHits, ok := trackerConfig["min_hits"].(float64); ok {
			tp.config.TrackerConfig.MinHits = int(minHits)
		}
		// ... other fields
	}

	return nil
}

// Start initializes the tracker
func (tp *TrackingProcessor) Start() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.tracker = tracking.NewObjectTracker(tp.config.TrackerConfig)
	return tp.BaseProcessor.Start()
}

// Stop stops the processor
func (tp *TrackingProcessor) Stop() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.tracker = nil
	return tp.BaseProcessor.Stop()
}

// Process processes a video frame
func (tp *TrackingProcessor) Process(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.tracker == nil {
		return nil, fmt.Errorf("tracker not initialized")
	}

	startTime := time.Now()

	// Get detections from input processor
	var detections []ml.Detection
	if result, ok := frame.GetResult(tp.config.InputProcessor); ok {
		if detResult, ok := result.(*ml.DetectionResult); ok {
			detections = detResult.Detections
		}
	}

	// Update tracker
	// Note: If detections are empty (either no objects or processor hasn't run),
	// the tracker will still predict and update ages.
	// Ideally we would want to wait for the input processor, but in a concurrent pipeline
	// without dependency management, we take what we have.
	tracks := tp.tracker.Update(detections)

	processTime := time.Since(startTime)
	tp.UpdateMetrics(processTime, true)

	// Convert tracks to MLResult
	// We need to convert tracking.Track to ml.Track
	// But wait, ml.Track IS tracking.Track?
	// No, ml.Track is defined in pkg/ml/types.go
	// tracking.Track is defined in pkg/ml/tracking/track.go
	// They are likely different structs or one imports the other.
	// Let's check pkg/ml/tracking/track.go

	mlTracks := make([]ml.Track, len(tracks))
	for i, t := range tracks {
		// Map tracking.Track to ml.Track
		// Assuming they have compatible fields
		mlTracks[i] = ml.Track{
			ID:         t.ID,
			Box:        t.GetBoundingBox(),
			ClassID:    0, // Tracker might not store ClassID directly if not passed
			ClassName:  t.Class,
			Confidence: t.Confidence,
			State:      ml.TrackState(t.State.String()), // Convert int enum to string enum
			Age:        t.Age,
			Hits:       t.HitStreak, // HitStreak vs Hits
			// ...
		}
	}

	return &ml.TrackingResult{
		Tracks:    mlTracks,
		Processor: tp.Name(),
		Timestamp: time.Now(),
	}, nil
}
