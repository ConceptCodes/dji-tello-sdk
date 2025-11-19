package tracking

import (
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
)

// TrackingFactory creates Tracking processors
type TrackingFactory struct{}

// NewTrackingFactory creates a new Tracking factory
func NewTrackingFactory() *TrackingFactory {
	return &TrackingFactory{}
}

// CreateProcessor creates a new Tracking processor
func (f *TrackingFactory) CreateProcessor(config map[string]interface{}) (processors.MLProcessor, error) {
	processor := NewTrackingProcessor("tracking")

	if err := processor.Configure(config); err != nil {
		return nil, err
	}

	return processor, nil
}

// GetProcessorType returns the processor type
func (f *TrackingFactory) GetProcessorType() ml.ProcessorType {
	return ml.ProcessorTypeTracking
}

// GetDefaultConfig returns default configuration for Tracking processor
func (f *TrackingFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"input_processor": "yolo",
		"tracker_config": map[string]interface{}{
			"max_distance":      50.0,
			"max_age":           30,
			"min_hits":          3,
			"max_iou_distance":  0.7,
			"use_kalman_filter": true,
			"enable_prediction": true,
		},
	}
}
