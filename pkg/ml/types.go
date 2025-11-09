package ml

import (
	"image"
	"time"
)

// ProcessorType defines the type of ML processor
type ProcessorType string

const (
	ProcessorTypeYOLO         ProcessorType = "yolo"
	ProcessorTypeFace         ProcessorType = "face"
	ProcessorTypeSLAM         ProcessorType = "slam"
	ProcessorTypeGesture      ProcessorType = "gesture"
	ProcessorTypeSegmentation ProcessorType = "segmentation"
	ProcessorTypeCustom       ProcessorType = "custom"
)

// MLResult represents the result from ML processing
type MLResult interface {
	GetProcessorName() string
	GetTimestamp() time.Time
	GetConfidence() float32
}

// Detection represents a detected object
type Detection struct {
	ClassID    int                    `json:"class_id"`
	ClassName  string                 `json:"class_name"`
	Confidence float32                `json:"confidence"`
	Box        image.Rectangle        `json:"box"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// GetProcessorName implements MLResult interface
func (d Detection) GetProcessorName() string {
	return "detection"
}

// GetTimestamp implements MLResult interface
func (d Detection) GetTimestamp() time.Time {
	return d.Timestamp
}

// GetConfidence implements MLResult interface
func (d Detection) GetConfidence() float32 {
	return d.Confidence
}

// DetectionResult represents multiple detections
type DetectionResult struct {
	Detections []Detection `json:"detections"`
	Processor  string      `json:"processor"`
	Timestamp  time.Time   `json:"timestamp"`
}

// GetProcessorName implements MLResult interface
func (dr DetectionResult) GetProcessorName() string {
	return dr.Processor
}

// GetTimestamp implements MLResult interface
func (dr DetectionResult) GetTimestamp() time.Time {
	return dr.Timestamp
}

// GetConfidence implements MLResult interface
func (dr DetectionResult) GetConfidence() float32 {
	if len(dr.Detections) == 0 {
		return 0
	}
	maxConf := float32(0)
	for _, det := range dr.Detections {
		if det.Confidence > maxConf {
			maxConf = det.Confidence
		}
	}
	return maxConf
}

// Point3D represents a 3D point
type Point3D struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Pose6D represents 6D pose (position and orientation)
type Pose6D struct {
	Position    Point3D    `json:"position"`
	Orientation Point3D    `json:"orientation"` // Roll, Pitch, Yaw in degrees
	Covariance  [9]float32 `json:"covariance"`
	Timestamp   time.Time  `json:"timestamp"`
}

// SLAMResult represents SLAM processing results
type SLAMResult struct {
	Pose      *Pose6D   `json:"pose"`
	KeyFrame  bool      `json:"key_frame"`
	Features  []Feature `json:"features"`
	Processor string    `json:"processor"`
	Timestamp time.Time `json:"timestamp"`
}

// GetProcessorName implements MLResult interface
func (sr SLAMResult) GetProcessorName() string {
	return sr.Processor
}

// GetTimestamp implements MLResult interface
func (sr SLAMResult) GetTimestamp() time.Time {
	return sr.Timestamp
}

// GetConfidence implements MLResult interface
func (sr SLAMResult) GetConfidence() float32 {
	return 1.0 // SLAM doesn't have confidence in the same way
}

// Feature represents a visual feature
type Feature struct {
	ID         int         `json:"id"`
	Position   image.Point `json:"position"`
	Descriptor []float32   `json:"descriptor"`
	Size       float32     `json:"size"`
	Angle      float32     `json:"angle"`
	Response   float32     `json:"response"`
	Octave     int         `json:"octave"`
	ClassID    int         `json:"class_id"`
}

// GestureResult represents gesture recognition results
type GestureResult struct {
	Gesture     string          `json:"gesture"`
	Confidence  float32         `json:"confidence"`
	BoundingBox image.Rectangle `json:"bounding_box"`
	Landmarks   []image.Point   `json:"landmarks"`
	Processor   string          `json:"processor"`
	Timestamp   time.Time       `json:"timestamp"`
}

// GetProcessorName implements MLResult interface
func (gr GestureResult) GetProcessorName() string {
	return gr.Processor
}

// GetTimestamp implements MLResult interface
func (gr GestureResult) GetTimestamp() time.Time {
	return gr.Timestamp
}

// GetConfidence implements MLResult interface
func (gr GestureResult) GetConfidence() float32 {
	return gr.Confidence
}

// DepthResult represents depth estimation results
type DepthResult struct {
	DepthMap   []float32 `json:"depth_map"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Confidence []float32 `json:"confidence"`
	Processor  string    `json:"processor"`
	Timestamp  time.Time `json:"timestamp"`
}

// GetProcessorName implements MLResult interface
func (dr DepthResult) GetProcessorName() string {
	return dr.Processor
}

// GetTimestamp implements MLResult interface
func (dr DepthResult) GetTimestamp() time.Time {
	return dr.Timestamp
}

// GetConfidence implements MLResult interface
func (dr DepthResult) GetConfidence() float32 {
	if len(dr.Confidence) == 0 {
		return 0
	}
	sum := float32(0)
	for _, conf := range dr.Confidence {
		sum += conf
	}
	return sum / float32(len(dr.Confidence))
}

// ProcessingError represents an error in ML processing
type ProcessingError struct {
	Processor string    `json:"processor"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Error implements error interface
func (pe ProcessingError) Error() string {
	return pe.Processor + ": " + pe.Message
}

// PipelineMetrics represents performance metrics for the ML pipeline
type PipelineMetrics struct {
	FPS            float64            `json:"fps"`
	Latency        time.Duration      `json:"latency"`
	DroppedFrames  int64              `json:"dropped_frames"`
	ProcessorStats map[string]float64 `json:"processor_stats"`
	MemoryUsage    int64              `json:"memory_usage"`
	GPUUsage       float64            `json:"gpu_usage"`
	LastUpdate     time.Time          `json:"last_update"`
}

// ProcessorStats represents statistics for a specific processor
type ProcessorStats struct {
	ProcessTime   time.Duration `json:"process_time"`
	SuccessCount  int64         `json:"success_count"`
	ErrorCount    int64         `json:"error_count"`
	AvgLatency    time.Duration `json:"avg_latency"`
	LastProcessed time.Time     `json:"last_processed"`
}

// MLConfig represents configuration for ML processors
type MLConfig struct {
	Processors []ProcessorConfig `json:"processors"`
	Pipeline   PipelineConfig    `json:"pipeline"`
	Overlay    OverlayConfig     `json:"overlay"`
}

// ProcessorConfig represents configuration for a single processor
type ProcessorConfig struct {
	Name     string                 `json:"name"`
	Type     ProcessorType          `json:"type"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
	Config   map[string]interface{} `json:"config"`
}

// PipelineConfig represents configuration for the ML pipeline
type PipelineConfig struct {
	MaxConcurrentProcessors int  `json:"max_concurrent_processors"`
	FrameBufferSize         int  `json:"frame_buffer_size"`
	WorkerPoolSize          int  `json:"worker_pool_size"`
	EnableMetrics           bool `json:"enable_metrics"`
	TargetFPS               int  `json:"target_fps"`
}

// OverlayConfig represents configuration for overlay rendering
type OverlayConfig struct {
	Enabled        bool              `json:"enabled"`
	ShowFPS        bool              `json:"show_fps"`
	ShowDetections bool              `json:"show_detections"`
	ShowTracking   bool              `json:"show_tracking"`
	ShowConfidence bool              `json:"show_confidence"`
	Colors         map[string]string `json:"colors"`
	LineWidth      int               `json:"line_width"`
	FontSize       int               `json:"font_size"`
	FontScale      float64           `json:"font_scale"`
}

// EnhancedVideoFrame extends the original VideoFrame with ML capabilities
type EnhancedVideoFrame struct {
	// Original frame data
	Data       []byte    `json:"data"`
	Timestamp  time.Time `json:"timestamp"`
	Size       int       `json:"size"`
	SeqNum     int       `json:"seq_num"`
	IsKeyFrame bool      `json:"is_key_frame"`

	// ML-specific fields
	Image     interface{}         `json:"-"` // Will hold image data (gocv.Mat when available)
	MLResults map[string]MLResult `json:"ml_results"`
	Processed bool                `json:"processed"`
	Width     int                 `json:"width"`
	Height    int                 `json:"height"`
	Channels  int                 `json:"channels"`
}

// NewEnhancedVideoFrame creates a new enhanced video frame
func NewEnhancedVideoFrame(data []byte, timestamp time.Time, seqNum int) *EnhancedVideoFrame {
	return &EnhancedVideoFrame{
		Data:      data,
		Timestamp: timestamp,
		SeqNum:    seqNum,
		Size:      len(data),
		MLResults: make(map[string]MLResult),
		Processed: false,
	}
}

// AddResult adds an ML result to the frame
func (evf *EnhancedVideoFrame) AddResult(processorName string, result MLResult) {
	if evf.MLResults == nil {
		evf.MLResults = make(map[string]MLResult)
	}
	evf.MLResults[processorName] = result
}

// GetResult gets an ML result by processor name
func (evf *EnhancedVideoFrame) GetResult(processorName string) (MLResult, bool) {
	if evf.MLResults == nil {
		return nil, false
	}
	result, exists := evf.MLResults[processorName]
	return result, exists
}

// GetAllResults returns all ML results
func (evf *EnhancedVideoFrame) GetAllResults() map[string]MLResult {
	if evf.MLResults == nil {
		return make(map[string]MLResult)
	}
	return evf.MLResults
}

// MarkProcessed marks the frame as processed
func (evf *EnhancedVideoFrame) MarkProcessed() {
	evf.Processed = true
}

// IsProcessed returns whether the frame has been processed
func (evf *EnhancedVideoFrame) IsProcessed() bool {
	return evf.Processed
}

// Cleanup releases resources
func (evf *EnhancedVideoFrame) Cleanup() {
	// TODO: Implement proper cleanup when OpenCV integration is complete
	evf.Image = nil
}
