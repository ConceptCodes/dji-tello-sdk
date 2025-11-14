package yolo

import (
	"context"
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
	"github.com/yalue/onnxruntime_go"
)

// YOLOProcessor implements YOLO object detection using ONNX Runtime
type YOLOProcessor struct {
	*processors.BaseProcessor
	session      *onnxruntime_go.AdvancedSession
	inputTensor  *onnxruntime_go.Tensor[float32]
	outputTensor *onnxruntime_go.Tensor[float32]
	inputName    string
	outputName   string
	config       *YOLOConfig
	classes      []string
	inputSize    image.Point
	running      bool
	mu           sync.Mutex // Thread safety for ONNX session
}

// YOLOConfig defines configuration for YOLO processor
type YOLOConfig struct {
	ModelPath    string   `json:"model_path"`
	Confidence   float32  `json:"confidence"`
	NMSThreshold float32  `json:"nms_threshold"`
	InputSize    [2]int   `json:"input_size"`
	Classes      []string `json:"classes"`
	Device       string   `json:"device"`
}

// NewYOLOProcessor creates a new YOLO processor
func NewYOLOProcessor(name string) *YOLOProcessor {
	return &YOLOProcessor{
		BaseProcessor: processors.NewBaseProcessor(name, ml.ProcessorTypeYOLO),
		config: &YOLOConfig{
			Confidence:   0.5,
			NMSThreshold: 0.4,
			InputSize:    [2]int{640, 640},
			Device:       "cpu",
		},
	}
}

// Process processes a video frame and returns detection results
func (yp *YOLOProcessor) Process(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error) {
	if !yp.running {
		return nil, fmt.Errorf("processor not started")
	}

	startTime := time.Now()
	defer func() {
		yp.UpdateMetrics(time.Since(startTime), true)
	}()

	// Get image from frame
	img := yp.getImageFromFrame(frame)
	if img == nil {
		yp.UpdateMetrics(time.Since(startTime), false)
		return nil, fmt.Errorf("no image data available in frame")
	}

	// Preprocess frame directly into input tensor with thread safety
	yp.mu.Lock()
	err := yp.preprocessFrame(img)
	yp.mu.Unlock()

	if err != nil {
		yp.UpdateMetrics(time.Since(startTime), false)
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	// Run inference with thread safety
	yp.mu.Lock()
	err = yp.session.Run()
	yp.mu.Unlock()

	if err != nil {
		yp.UpdateMetrics(time.Since(startTime), false)
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Postprocess results using the output tensor with thread safety
	yp.mu.Lock()
	detections, err := yp.postprocessResults(yp.outputTensor, image.Rect(0, 0, frame.Width, frame.Height))
	yp.mu.Unlock()

	if err != nil {
		yp.UpdateMetrics(time.Since(startTime), false)
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	// Create detection result
	result := &ml.DetectionResult{
		Detections: detections,
		Timestamp:  time.Now(),
		Processor:  yp.Name(),
	}

	return result, nil
}

// Configure configures the YOLO processor
func (yp *YOLOProcessor) Configure(config map[string]interface{}) error {
	if err := yp.BaseProcessor.Configure(config); err != nil {
		return err
	}

	// Parse YOLO-specific configuration
	if err := yp.parseConfig(config); err != nil {
		return fmt.Errorf("failed to parse YOLO config: %w", err)
	}

	return nil
}

// Start initializes the ONNX Runtime session
func (yp *YOLOProcessor) Start() error {
	if yp.running {
		return fmt.Errorf("processor already running")
	}

	// Initialize ONNX Runtime
	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		return fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}

	// Define input and output names (YOLOv8 standard)
	inputNames := []string{"images"}
	outputNames := []string{"output0"}

	// Create input and output tensors
	inputShape := []int64{1, 3, int64(yp.config.InputSize[1]), int64(yp.config.InputSize[0])}
	outputShape := []int64{1, 84, 8400} // YOLOv8 standard output shape

	inputTensor, err := onnxruntime_go.NewEmptyTensor[float32](inputShape)
	if err != nil {
		return fmt.Errorf("failed to create input tensor: %w", err)
	}

	outputTensor, err := onnxruntime_go.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return fmt.Errorf("failed to create output tensor: %w", err)
	}

	// Load model using AdvancedSession
	session, err := onnxruntime_go.NewAdvancedSession(
		yp.config.ModelPath,
		inputNames,
		outputNames,
		[]onnxruntime_go.Value{inputTensor},
		[]onnxruntime_go.Value{outputTensor},
		nil, // Use default options
	)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	yp.inputName = inputNames[0]
	yp.outputName = outputNames[0]
	yp.session = session
	yp.inputTensor = inputTensor
	yp.outputTensor = outputTensor
	yp.running = true

	// Set input size
	yp.inputSize = image.Point{
		X: yp.config.InputSize[0],
		Y: yp.config.InputSize[1],
	}

	// Set classes
	yp.classes = yp.config.Classes

	return yp.BaseProcessor.Start()
}

// Stop stops the processor and releases resources
func (yp *YOLOProcessor) Stop() error {
	if !yp.running {
		return nil
	}

	yp.running = false

	if yp.session != nil {
		yp.session.Destroy()
		yp.session = nil
	}

	if yp.inputTensor != nil {
		yp.inputTensor.Destroy()
		yp.inputTensor = nil
	}

	if yp.outputTensor != nil {
		yp.outputTensor.Destroy()
		yp.outputTensor = nil
	}

	return yp.BaseProcessor.Stop()
}

// IsRunning returns whether the processor is currently running
func (yp *YOLOProcessor) IsRunning() bool {
	return yp.running
}

// ValidateConfig validates the YOLO configuration
func (yp *YOLOProcessor) ValidateConfig(config map[string]interface{}) error {
	if err := yp.BaseProcessor.ValidateConfig(config); err != nil {
		return err
	}

	// Validate required fields
	if _, ok := config["model_path"]; !ok {
		return fmt.Errorf("model_path is required")
	}

	if _, ok := config["confidence"]; !ok {
		return fmt.Errorf("confidence is required")
	}

	if _, ok := config["nms_threshold"]; !ok {
		return fmt.Errorf("nms_threshold is required")
	}

	if _, ok := config["input_size"]; !ok {
		return fmt.Errorf("input_size is required")
	}

	if _, ok := config["classes"]; !ok {
		return fmt.Errorf("classes is required")
	}

	return nil
}

// parseConfig parses configuration into YOLOConfig
func (yp *YOLOProcessor) parseConfig(config map[string]interface{}) error {
	// Handle both "model_path" and "model" keys for compatibility
	if modelPath, ok := config["model_path"].(string); ok {
		yp.config.ModelPath = modelPath
	} else if model, ok := config["model"].(string); ok {
		yp.config.ModelPath = model
	}

	if confidence, ok := config["confidence"].(float64); ok {
		yp.config.Confidence = float32(confidence)
	}

	if nmsThreshold, ok := config["nms_threshold"].(float64); ok {
		yp.config.NMSThreshold = float32(nmsThreshold)
	}

	if inputSize, ok := config["input_size"].([]interface{}); ok && len(inputSize) == 2 {
		if width, ok := inputSize[0].(float64); ok {
			yp.config.InputSize[0] = int(width)
		}
		if height, ok := inputSize[1].(float64); ok {
			yp.config.InputSize[1] = int(height)
		}
	}

	if classes, ok := config["classes"].([]interface{}); ok {
		yp.config.Classes = make([]string, len(classes))
		for i, class := range classes {
			if className, ok := class.(string); ok {
				yp.config.Classes[i] = className
			}
		}
	}

	if device, ok := config["device"].(string); ok {
		yp.config.Device = device
	}

	return nil
}

// getImageFromFrame extracts image data from EnhancedVideoFrame
func (yp *YOLOProcessor) getImageFromFrame(frame *ml.EnhancedVideoFrame) image.Image {
	if frame.Image == nil {
		return nil
	}

	// Type assert to image.Image
	if img, ok := frame.Image.(image.Image); ok {
		return img
	}

	return nil
}

// calculateOverallConfidence calculates overall confidence from detections
func (yp *YOLOProcessor) calculateOverallConfidence(detections []ml.Detection) float32 {
	if len(detections) == 0 {
		return 0.0
	}

	var totalConfidence float32
	for _, detection := range detections {
		totalConfidence += detection.Confidence
	}

	return totalConfidence / float32(len(detections))
}
