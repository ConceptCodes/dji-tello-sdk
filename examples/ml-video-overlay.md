# ML Video Overlay Example

This example demonstrates how to use machine learning processing with real-time video overlay from a DJI Tello drone.

## Overview

The ML video overlay example shows:
- Connecting to a Tello drone and starting video stream
- Setting up ML processing pipeline with YOLO object detection
- Configuring object tracking
- Creating web-based video display with ML overlays
- Real-time visualization of detections and tracking

## Code

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
	// Initialize drone
	drone, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize drone: %v", err)
	}

	// Enter SDK mode
	if err := drone.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Start video stream
	if err := drone.StreamOn(); err != nil {
		log.Fatalf("Failed to start video stream: %v", err)
	}
	defer drone.StreamOff()

	// Create ML configuration with overlay enabled
	mlConfig := createMLConfig()

	// Create ML video integration
	mlIntegration, err := transport.NewMLVideoIntegration("0.0.0.0:11111", mlConfig)
	if err != nil {
		log.Fatalf("Failed to create ML video integration: %v", err)
	}

	// Create integrated video display with ML overlays
	display := mlIntegration.CreateIntegratedVideoDisplay(transport.DisplayTypeWeb)
	display.SetWebPort(8080)

	// Start ML integration
	if err := mlIntegration.Start(); err != nil {
		log.Fatalf("Failed to start ML integration: %v", err)
	}
	defer mlIntegration.Stop()

	// Start video display
	if err := display.Start(); err != nil {
		log.Fatalf("Failed to start video display: %v", err)
	}
	defer display.Close()

	log.Println("🎥 ML Video GUI with Overlay started!")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("🤖 ML processing enabled with overlay visualization")
	log.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Stopping ML video GUI...")
}

// createMLConfig creates a default ML configuration with overlay enabled
func createMLConfig() *ml.MLConfig {
	return &ml.MLConfig{
		Processors: []ml.ProcessorConfig{
			{
				Name:     "yolo-detector",
				Type:     ml.ProcessorTypeYOLO,
				Enabled:  true,
				Priority: 1,
				Config: map[string]interface{}{
					"model_path":     "models/yolov8n_real.onnx",
					"confidence":     0.5,
					"nms_threshold":  0.4,
					"input_size":     640,
					"max_detections": 100,
				},
			},
			{
				Name:     "tracker",
				Type:     ml.ProcessorTypeCustom,
				Enabled:  true,
				Priority: 2,
				Config: map[string]interface{}{
					"max_age":         30,
					"min_hits":        3,
					"max_disappeared": 10,
					"iou_threshold":   0.3,
				},
			},
		},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 4,
			FrameBufferSize:         30,
			WorkerPoolSize:          2,
			EnableMetrics:           true,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled:        true,
			ShowFPS:        true,
			ShowDetections: true,
			ShowTracking:   true,
			ShowConfidence: true,
			Colors: map[string]string{
				"person":  "#FF0000",
				"car":     "#0000FF",
				"face":    "#FF00FF",
				"default": "#00FF00",
			},
			LineWidth: 2,
			FontSize:  12,
			FontScale: 0.5,
		},
	}
}
```

## How to Run

```bash
# Save as ml_video_overlay_example.go
export DYLD_LIBRARY_PATH=./lib:$DYLD_LIBRARY_PATH
go run ml_video_overlay_example.go
```

## Features Demonstrated

1. **Drone Connection**: Connects to Tello drone and initializes SDK mode
2. **Video Streaming**: Starts video stream from drone
3. **ML Pipeline**: Sets up YOLO object detection and tracking
4. **Web Interface**: Creates web-based video display at localhost:8080
5. **Real-time Overlay**: Shows bounding boxes, labels, and tracking IDs
6. **Configuration**: Customizable ML parameters and overlay settings

## ML Configuration

The example uses two processors:

### YOLO Detector
- **Model**: YOLOv8 Nano for real-time detection
- **Confidence**: 0.5 (minimum detection confidence)
- **NMS Threshold**: 0.4 (non-maximum suppression)
- **Input Size**: 640x640 pixels
- **Max Detections**: 100 objects per frame

### Object Tracker
- **Max Age**: 30 frames (track persistence)
- **Min Hits**: 3 confirmations before track creation
- **Max Disappeared**: 10 frames before track deletion
- **IOU Threshold**: 0.3 (intersection over union for matching)

## Overlay Configuration

The overlay displays:
- **FPS**: Current processing frame rate
- **Detections**: Bounding boxes with class labels
- **Tracking**: Track IDs and movement paths
- **Confidence**: Detection confidence scores
- **Colors**: Different colors for different object classes

## Expected Output

```
🎥 ML Video GUI with Overlay started!
🌐 Open http://localhost:8080 in your browser
🤖 ML processing enabled with overlay visualization
Press Ctrl+C to stop
```

The web interface will show:
- Live video feed from the drone
- Green/red/blue bounding boxes around detected objects
- Track IDs (e.g., "Person #1", "Car #2")
- Confidence scores
- FPS counter
- Object class labels

## Requirements

- DJI Tello drone connected to WiFi
- ONNX Runtime library (`lib/onnxruntime.so`)
- YOLO model file (`models/yolov8n_real.onnx`)
- Sufficient GPU/CPU for real-time processing
- Modern web browser for display interface

## Troubleshooting

1. **ONNX Runtime Error**: Set `DYLD_LIBRARY_PATH=./lib:$DYLD_LIBRARY_PATH` (macOS) or `LD_LIBRARY_PATH=./lib:$LD_LIBRARY_PATH` (Linux)
2. **Model Not Found**: Ensure YOLO model exists in `models/` directory
3. **Drone Connection**: Check drone WiFi connection and SDK mode
4. **Performance**: Reduce input size or confidence threshold for better FPS

## Customization

You can modify the ML configuration to:
- Use different YOLO models (yolov8s, yolov8m, yolov8l)
- Adjust detection thresholds
- Change overlay colors and styles
- Add more processors (face detection, segmentation)
- Modify tracking parameters