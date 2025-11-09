package ml

import (
	"image"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetection_GetProcessorName(t *testing.T) {
	detection := Detection{
		ClassName: "person",
	}
	assert.Equal(t, "detection", detection.GetProcessorName())
}

func TestDetection_GetTimestamp(t *testing.T) {
	timestamp := time.Now()
	detection := Detection{
		Timestamp: timestamp,
	}
	assert.Equal(t, timestamp, detection.GetTimestamp())
}

func TestDetection_GetConfidence(t *testing.T) {
	detection := Detection{
		Confidence: 0.85,
	}
	assert.Equal(t, float32(0.85), detection.GetConfidence())
}

func TestDetectionResult_GetProcessorName(t *testing.T) {
	result := DetectionResult{
		Processor: "yolo",
	}
	assert.Equal(t, "yolo", result.GetProcessorName())
}

func TestDetectionResult_GetTimestamp(t *testing.T) {
	timestamp := time.Now()
	result := DetectionResult{
		Timestamp: timestamp,
	}
	assert.Equal(t, timestamp, result.GetTimestamp())
}

func TestDetectionResult_GetConfidence(t *testing.T) {
	tests := []struct {
		name       string
		detections []Detection
		expected   float32
	}{
		{
			name:       "no detections",
			detections: []Detection{},
			expected:   0,
		},
		{
			name: "single detection",
			detections: []Detection{
				{Confidence: 0.75},
			},
			expected: 0.75,
		},
		{
			name: "multiple detections",
			detections: []Detection{
				{Confidence: 0.75},
				{Confidence: 0.90},
				{Confidence: 0.60},
			},
			expected: 0.90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectionResult{Detections: tt.detections}
			assert.Equal(t, tt.expected, result.GetConfidence())
		})
	}
}

func TestSLAMResult_GetProcessorName(t *testing.T) {
	result := SLAMResult{
		Processor: "slam",
	}
	assert.Equal(t, "slam", result.GetProcessorName())
}

func TestSLAMResult_GetTimestamp(t *testing.T) {
	timestamp := time.Now()
	result := SLAMResult{
		Timestamp: timestamp,
	}
	assert.Equal(t, timestamp, result.GetTimestamp())
}

func TestSLAMResult_GetConfidence(t *testing.T) {
	result := SLAMResult{}
	assert.Equal(t, float32(1.0), result.GetConfidence())
}

func TestGestureResult_GetProcessorName(t *testing.T) {
	result := GestureResult{
		Processor: "gesture",
	}
	assert.Equal(t, "gesture", result.GetProcessorName())
}

func TestGestureResult_GetTimestamp(t *testing.T) {
	timestamp := time.Now()
	result := GestureResult{
		Timestamp: timestamp,
	}
	assert.Equal(t, timestamp, result.GetTimestamp())
}

func TestGestureResult_GetConfidence(t *testing.T) {
	result := GestureResult{
		Confidence: 0.95,
	}
	assert.Equal(t, float32(0.95), result.GetConfidence())
}

func TestDepthResult_GetProcessorName(t *testing.T) {
	result := DepthResult{
		Processor: "depth",
	}
	assert.Equal(t, "depth", result.GetProcessorName())
}

func TestDepthResult_GetTimestamp(t *testing.T) {
	timestamp := time.Now()
	result := DepthResult{
		Timestamp: timestamp,
	}
	assert.Equal(t, timestamp, result.GetTimestamp())
}

func TestDepthResult_GetConfidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence []float32
		expected   float32
	}{
		{
			name:       "no confidence values",
			confidence: []float32{},
			expected:   0,
		},
		{
			name:       "single confidence value",
			confidence: []float32{0.80},
			expected:   0.80,
		},
		{
			name:       "multiple confidence values",
			confidence: []float32{0.75, 0.85, 0.90},
			expected:   0.8333333, // Average
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DepthResult{Confidence: tt.confidence}
			assert.InDelta(t, tt.expected, result.GetConfidence(), 0.001)
		})
	}
}

func TestProcessingError_Error(t *testing.T) {
	err := ProcessingError{
		Processor: "yolo",
		Message:   "model not found",
	}
	assert.Equal(t, "yolo: model not found", err.Error())
}

func TestEnhancedVideoFrame_NewEnhancedVideoFrame(t *testing.T) {
	data := []byte("test video data")
	timestamp := time.Now()
	seqNum := 42

	frame := NewEnhancedVideoFrame(data, timestamp, seqNum)

	assert.Equal(t, data, frame.Data)
	assert.Equal(t, timestamp, frame.Timestamp)
	assert.Equal(t, seqNum, frame.SeqNum)
	assert.Equal(t, len(data), frame.Size)
	assert.False(t, frame.Processed)
	assert.NotNil(t, frame.MLResults)
	assert.Empty(t, frame.MLResults)
}

func TestEnhancedVideoFrame_AddResult(t *testing.T) {
	frame := NewEnhancedVideoFrame([]byte("test"), time.Now(), 1)
	result := Detection{ClassName: "person", Confidence: 0.85}

	frame.AddResult("yolo", &result)

	assert.Contains(t, frame.MLResults, "yolo")
	assert.Equal(t, &result, frame.MLResults["yolo"])
}

func TestEnhancedVideoFrame_GetResult(t *testing.T) {
	frame := NewEnhancedVideoFrame([]byte("test"), time.Now(), 1)
	result := Detection{ClassName: "person", Confidence: 0.85}
	frame.AddResult("yolo", &result)

	// Test existing result
	retrieved, exists := frame.GetResult("yolo")
	assert.True(t, exists)
	assert.Equal(t, &result, retrieved)

	// Test non-existing result
	_, exists = frame.GetResult("nonexistent")
	assert.False(t, exists)
}

func TestEnhancedVideoFrame_GetAllResults(t *testing.T) {
	frame := NewEnhancedVideoFrame([]byte("test"), time.Now(), 1)

	// Initially empty
	results := frame.GetAllResults()
	assert.Empty(t, results)

	// Add some results
	result1 := Detection{ClassName: "person", Confidence: 0.85}
	result2 := Detection{ClassName: "car", Confidence: 0.90}
	frame.AddResult("yolo", &result1)
	frame.AddResult("detector", &result2)

	results = frame.GetAllResults()
	assert.Len(t, results, 2)
	assert.Contains(t, results, "yolo")
	assert.Contains(t, results, "detector")
}

func TestEnhancedVideoFrame_MarkProcessed(t *testing.T) {
	frame := NewEnhancedVideoFrame([]byte("test"), time.Now(), 1)
	assert.False(t, frame.IsProcessed())

	frame.MarkProcessed()
	assert.True(t, frame.IsProcessed())
}

func TestEnhancedVideoFrame_Cleanup(t *testing.T) {
	frame := NewEnhancedVideoFrame([]byte("test"), time.Now(), 1)
	frame.Image = "some image data" // Simulate image data

	frame.Cleanup()
	assert.Nil(t, frame.Image)
}

func TestProcessorType_Constants(t *testing.T) {
	assert.Equal(t, ProcessorType("yolo"), ProcessorTypeYOLO)
	assert.Equal(t, ProcessorType("face"), ProcessorTypeFace)
	assert.Equal(t, ProcessorType("slam"), ProcessorTypeSLAM)
	assert.Equal(t, ProcessorType("gesture"), ProcessorTypeGesture)
	assert.Equal(t, ProcessorType("segmentation"), ProcessorTypeSegmentation)
	assert.Equal(t, ProcessorType("custom"), ProcessorTypeCustom)
}

func TestPoint3D_Structure(t *testing.T) {
	point := Point3D{X: 1.0, Y: 2.0, Z: 3.0}
	assert.Equal(t, float32(1.0), point.X)
	assert.Equal(t, float32(2.0), point.Y)
	assert.Equal(t, float32(3.0), point.Z)
}

func TestPose6D_Structure(t *testing.T) {
	position := Point3D{X: 1.0, Y: 2.0, Z: 3.0}
	orientation := Point3D{X: 0.0, Y: 0.0, Z: 90.0}
	covariance := [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}
	timestamp := time.Now()

	pose := Pose6D{
		Position:    position,
		Orientation: orientation,
		Covariance:  covariance,
		Timestamp:   timestamp,
	}

	assert.Equal(t, position, pose.Position)
	assert.Equal(t, orientation, pose.Orientation)
	assert.Equal(t, covariance, pose.Covariance)
	assert.Equal(t, timestamp, pose.Timestamp)
}

func TestFeature_Structure(t *testing.T) {
	feature := Feature{
		ID:         1,
		Position:   image.Point{X: 100, Y: 200},
		Descriptor: []float32{0.1, 0.2, 0.3},
		Size:       16.0,
		Angle:      45.0,
		Response:   0.85,
		Octave:     0,
		ClassID:    1,
	}

	assert.Equal(t, 1, feature.ID)
	assert.Equal(t, image.Point{X: 100, Y: 200}, feature.Position)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, feature.Descriptor)
	assert.Equal(t, float32(16.0), feature.Size)
	assert.Equal(t, float32(45.0), feature.Angle)
	assert.Equal(t, float32(0.85), feature.Response)
	assert.Equal(t, 0, feature.Octave)
	assert.Equal(t, 1, feature.ClassID)
}

func TestPipelineMetrics_Structure(t *testing.T) {
	metrics := PipelineMetrics{
		FPS:            30.0,
		Latency:        33 * time.Millisecond,
		DroppedFrames:  5,
		ProcessorStats: map[string]float64{"yolo": 10.5},
		MemoryUsage:    1024 * 1024,
		GPUUsage:       75.5,
		LastUpdate:     time.Now(),
	}

	assert.Equal(t, 30.0, metrics.FPS)
	assert.Equal(t, 33*time.Millisecond, metrics.Latency)
	assert.Equal(t, int64(5), metrics.DroppedFrames)
	assert.Equal(t, map[string]float64{"yolo": 10.5}, metrics.ProcessorStats)
	assert.Equal(t, int64(1024*1024), metrics.MemoryUsage)
	assert.Equal(t, 75.5, metrics.GPUUsage)
}

func TestProcessorStats_Structure(t *testing.T) {
	stats := ProcessorStats{
		ProcessTime:   10 * time.Millisecond,
		SuccessCount:  100,
		ErrorCount:    2,
		AvgLatency:    15 * time.Millisecond,
		LastProcessed: time.Now(),
	}

	assert.Equal(t, 10*time.Millisecond, stats.ProcessTime)
	assert.Equal(t, int64(100), stats.SuccessCount)
	assert.Equal(t, int64(2), stats.ErrorCount)
	assert.Equal(t, 15*time.Millisecond, stats.AvgLatency)
}

func TestMLConfig_Structure(t *testing.T) {
	processorConfig := ProcessorConfig{
		Name:     "yolo",
		Type:     ProcessorTypeYOLO,
		Enabled:  true,
		Priority: 1,
		Config:   map[string]interface{}{"threshold": 0.5},
	}

	pipelineConfig := PipelineConfig{
		MaxConcurrentProcessors: 4,
		FrameBufferSize:         100,
		WorkerPoolSize:          2,
		EnableMetrics:           true,
		TargetFPS:               30,
	}

	overlayConfig := OverlayConfig{
		Enabled:        true,
		ShowFPS:        true,
		ShowDetections: true,
		Colors:         map[string]string{"person": "red"},
		LineWidth:      2,
		FontSize:       12,
		FontScale:      0.5,
	}

	config := MLConfig{
		Processors: []ProcessorConfig{processorConfig},
		Pipeline:   pipelineConfig,
		Overlay:    overlayConfig,
	}

	assert.Len(t, config.Processors, 1)
	assert.Equal(t, processorConfig, config.Processors[0])
	assert.Equal(t, pipelineConfig, config.Pipeline)
	assert.Equal(t, overlayConfig, config.Overlay)
}
