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
					"model_path":     "models/yolo-v8n",
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
