package integration

import (
	"image"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/stretchr/testify/assert"
)

// MockVideoFrame represents a simplified video frame for testing
type MockVideoFrame struct {
	Data      []byte
	Timestamp time.Time
	SeqNum    int
	Size      int
}

// MockMLProcessor simulates an ML processor for testing
type MockMLProcessor struct {
	Name      string
	Enabled   bool
	ProcessFn func(frame *MockVideoFrame) (ml.MLResult, error)
}

// MockVideoOverlay simulates video overlay rendering
type MockVideoOverlay struct {
	Results []ml.MLResult
	Enabled bool
}

// MockVideoMLPipeline integrates video and ML components
type MockVideoMLPipeline struct {
	VideoChan   chan *MockVideoFrame
	MLResults   chan ml.MLResult
	Processors  []*MockMLProcessor
	Overlay     *MockVideoOverlay
	FrameCount  int
	ResultCount int
}

// NewMockVideoMLPipeline creates a new mock pipeline
func NewMockVideoMLPipeline() *MockVideoMLPipeline {
	return &MockVideoMLPipeline{
		VideoChan:   make(chan *MockVideoFrame, 10),
		MLResults:   make(chan ml.MLResult, 10),
		Processors:  make([]*MockMLProcessor, 0),
		Overlay:     &MockVideoOverlay{Results: make([]ml.MLResult, 0), Enabled: true},
		FrameCount:  0,
		ResultCount: 0,
	}
}

// AddProcessor adds an ML processor to the pipeline
func (p *MockVideoMLPipeline) AddProcessor(processor *MockMLProcessor) {
	p.Processors = append(p.Processors, processor)
}

// ProcessFrame processes a video frame through the ML pipeline
func (p *MockVideoMLPipeline) ProcessFrame(frame *MockVideoFrame) ([]ml.MLResult, error) {
	p.FrameCount++
	results := make([]ml.MLResult, 0)

	for _, processor := range p.Processors {
		if processor.Enabled {
			result, err := processor.ProcessFn(frame)
			if err != nil {
				continue
			}
			results = append(results, result)
			p.ResultCount++
		}
	}

	return results, nil
}

// RenderOverlay renders ML results onto video overlay
func (p *MockVideoMLPipeline) RenderOverlay(results []ml.MLResult) {
	if p.Overlay.Enabled {
		p.Overlay.Results = append(p.Overlay.Results, results...)
	}
}

// TestVideoMLIntegration_FrameFlow tests that frames flow from video to ML pipeline
func TestVideoMLIntegration_FrameFlow(t *testing.T) {
	// Create mock pipeline
	pipeline := NewMockVideoMLPipeline()

	// Create a mock ML processor that detects objects
	detectionProcessor := &MockMLProcessor{
		Name:    "yolo_detector",
		Enabled: true,
		ProcessFn: func(frame *MockVideoFrame) (ml.MLResult, error) {
			// Simulate detection result
			return &ml.DetectionResult{
				Processor: "yolo",
				Timestamp: time.Now(),
				Detections: []ml.Detection{
					{
						ClassID:    0,
						ClassName:  "person",
						Confidence: 0.85,
						Box:        image.Rect(100, 100, 200, 200),
						Timestamp:  time.Now(),
					},
				},
			}, nil
		},
	}
	pipeline.AddProcessor(detectionProcessor)

	// Create test frames
	frames := []*MockVideoFrame{
		{Data: []byte("frame1"), Timestamp: time.Now(), SeqNum: 1, Size: 100},
		{Data: []byte("frame2"), Timestamp: time.Now(), SeqNum: 2, Size: 100},
		{Data: []byte("frame3"), Timestamp: time.Now(), SeqNum: 3, Size: 100},
	}

	// Process frames through pipeline
	for _, frame := range frames {
		results, err := pipeline.ProcessFrame(frame)
		assert.NoError(t, err)
		assert.Len(t, results, 1, "Each frame should produce one ML result")
	}

	// Verify frame flow
	assert.Equal(t, 3, pipeline.FrameCount, "All 3 frames should be processed")
	assert.Equal(t, 3, pipeline.ResultCount, "All 3 frames should produce results")
}

// TestVideoMLIntegration_ResultFlow tests that ML results flow back to video overlay
func TestVideoMLIntegration_ResultFlow(t *testing.T) {
	// Create mock pipeline
	pipeline := NewMockVideoMLPipeline()

	// Create mock processors
	yoloProcessor := &MockMLProcessor{
		Name:    "yolo",
		Enabled: true,
		ProcessFn: func(frame *MockVideoFrame) (ml.MLResult, error) {
			return &ml.DetectionResult{
				Processor: "yolo",
				Timestamp: time.Now(),
				Detections: []ml.Detection{
					{ClassName: "person", Confidence: 0.90},
				},
			}, nil
		},
	}

	gestureProcessor := &MockMLProcessor{
		Name:    "gesture",
		Enabled: true,
		ProcessFn: func(frame *MockVideoFrame) (ml.MLResult, error) {
			return &ml.GestureResult{
				Processor:  "gesture",
				Timestamp:  time.Now(),
				Gesture:    "thumbs_up",
				Confidence: 0.95,
			}, nil
		},
	}

	pipeline.AddProcessor(yoloProcessor)
	pipeline.AddProcessor(gestureProcessor)

	// Process a frame
	frame := &MockVideoFrame{Data: []byte("test"), Timestamp: time.Now(), SeqNum: 1, Size: 50}
	results, err := pipeline.ProcessFrame(frame)
	assert.NoError(t, err)
	assert.Len(t, results, 2, "Should have results from both processors")

	// Render results to overlay
	pipeline.RenderOverlay(results)

	// Verify results in overlay
	assert.Len(t, pipeline.Overlay.Results, 2, "Overlay should contain both ML results")

	// Verify result types
	detectionResult, ok := pipeline.Overlay.Results[0].(*ml.DetectionResult)
	assert.True(t, ok, "First result should be DetectionResult")
	assert.Equal(t, "yolo", detectionResult.Processor)
	assert.Len(t, detectionResult.Detections, 1)
	assert.Equal(t, "person", detectionResult.Detections[0].ClassName)

	gestureResult, ok := pipeline.Overlay.Results[1].(*ml.GestureResult)
	assert.True(t, ok, "Second result should be GestureResult")
	assert.Equal(t, "gesture", gestureResult.Processor)
	assert.Equal(t, "thumbs_up", gestureResult.Gesture)
}
