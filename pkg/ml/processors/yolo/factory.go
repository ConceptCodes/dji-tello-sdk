package yolo

import (
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
)

// YOLOFactory creates YOLO processors
type YOLOFactory struct{}

// NewYOLOFactory creates a new YOLO factory
func NewYOLOFactory() *YOLOFactory {
	return &YOLOFactory{}
}

// CreateProcessor creates a new YOLO processor
func (yf *YOLOFactory) CreateProcessor(config map[string]interface{}) (processors.MLProcessor, error) {
	processor := NewYOLOProcessor("yolo_detector")

	if err := processor.Configure(config); err != nil {
		return nil, err
	}

	return processor, nil
}

// GetProcessorType returns the processor type this factory creates
func (yf *YOLOFactory) GetProcessorType() ml.ProcessorType {
	return ml.ProcessorTypeYOLO
}

// GetDefaultConfig returns default configuration for YOLO processor
func (yf *YOLOFactory) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"model_path":    "yolov8n.onnx",
		"confidence":    0.5,
		"nms_threshold": 0.4,
		"input_size":    []interface{}{640, 640},
		"classes": []string{
			"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck",
			"boat", "traffic light", "fire hydrant", "stop sign", "parking meter", "bench",
			"bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra",
			"giraffe", "backpack", "umbrella", "handbag", "tie", "suitcase", "frisbee",
			"skis", "snowboard", "sports ball", "kite", "baseball bat", "baseball glove",
			"skateboard", "surfboard", "tennis racket", "bottle", "wine glass", "cup",
			"fork", "knife", "spoon", "bowl", "banana", "apple", "sandwich", "orange",
			"broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair", "couch",
			"potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
			"remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink",
			"refrigerator", "book", "clock", "vase", "scissors", "teddy bear", "hair drier",
			"toothbrush",
		},
		"device": "cpu",
	}
}
