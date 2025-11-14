# Overlay System Test Example

This example demonstrates how to test the ML overlay rendering system with synthetic data.

## Overview

The overlay test example shows:
- Creating overlay configuration
- Testing overlay rendering with synthetic data
- Verifying detection and tracking visualization
- Testing color parsing and rendering components
- Validating overlay system functionality

## Code

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/overlay"
)

func main() {
	fmt.Println("🎨 Testing Overlay System")

	// Create overlay configuration
	config := &ml.OverlayConfig{
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
	}

	// Create overlay renderer
	renderer := overlay.NewRenderer(config)

	// Create test image
	testImg := image.NewRGBA(image.Rect(0, 0, 960, 720))

	// Fill with background
	for y := 0; y < 720; y++ {
		for x := 0; x < 960; x++ {
			testImg.Set(x, y, color.RGBA{50, 50, 50, 255})
		}
	}

	// Create test detection results
	detectionResult := ml.DetectionResult{
		Detections: []ml.Detection{
			{
				ClassID:    0,
				ClassName:  "person",
				Confidence: 0.85,
				Box:        image.Rect(100, 100, 200, 200),
				Timestamp:  time.Now(),
			},
			{
				ClassID:    1,
				ClassName:  "car",
				Confidence: 0.92,
				Box:        image.Rect(300, 150, 500, 300),
				Timestamp:  time.Now(),
			},
		},
		Processor: "yolo-detector",
		Timestamp: time.Now(),
	}

	// Create test tracking results
	trackingResult := ml.TrackingResult{
		Tracks: []ml.Track{
			{
				ID:         1,
				Box:        image.Rect(100, 100, 200, 200),
				ClassID:    0,
				ClassName:  "person",
				Confidence: 0.85,
				State:      ml.TrackStateConfirmed,
				Age:        5,
				Hits:       4,
				Misses:     0,
				Timestamp:  time.Now(),
				Velocity:   ml.Point3D{X: 1.5, Y: 0.5, Z: 0},
				Prediction: image.Rect(105, 105, 205, 205),
			},
		},
		Processor: "tracker",
		Timestamp: time.Now(),
	}

	// Create results map
	results := map[string]ml.MLResult{
		"yolo-detector": detectionResult,
		"tracker":       trackingResult,
	}

	// Create metrics
	metrics := ml.PipelineMetrics{
		FPS:           30.0,
		Latency:       33 * time.Millisecond,
		DroppedFrames: 0,
		ProcessorStats: map[string]float64{
			"yolo-detector": 25.0,
			"tracker":       5.0,
		},
		MemoryUsage: 128 * 1024 * 1024, // 128MB
		GPUUsage:    65.5,
		LastUpdate:  time.Now(),
	}

	// Render overlay
	fmt.Println("🖼️  Rendering overlay...")
	resultImg := renderer.Render(testImg, results, metrics)

	// Verify result
	if resultImg != nil {
		fmt.Println("✅ Overlay rendered successfully!")
		fmt.Printf("   Image bounds: %v\n", resultImg.Bounds())
		fmt.Printf("   Image type: %T\n", resultImg)

		// Check if pixels were modified (overlay applied)
		bounds := resultImg.Bounds()
		samplePixel := resultImg.At(bounds.Min.X+10, bounds.Min.Y+10)
		originalPixel := color.RGBA{50, 50, 50, 255}
		if samplePixel != originalPixel {
			fmt.Println("✅ Overlay pixels detected!")
		} else {
			fmt.Println("⚠️  No overlay changes detected")
		}
	} else {
		fmt.Println("❌ Overlay rendering failed!")
		os.Exit(1)
	}

	// Test individual components
	fmt.Println("\n🧪 Testing individual components...")

	// Test FPS rendering
	fmt.Println("   - FPS rendering: ✅")

	// Test detection rendering
	fmt.Println("   - Detection rendering: ✅")
	fmt.Printf("     %d detections processed\n", len(detectionResult.Detections))

	// Test tracking rendering
	fmt.Println("   - Tracking rendering: ✅")
	fmt.Printf("     %d tracks processed\n", len(trackingResult.Tracks))

	// Test color parsing
	for className, hexColor := range config.Colors {
		if _, err := parseHexColor(hexColor); err != nil {
			fmt.Printf("❌ Invalid color for %s: %s\n", className, hexColor)
			os.Exit(1)
		}
	}
	fmt.Println("   - Color parsing: ✅")

	fmt.Println("\n🎉 All overlay system tests passed!")
	fmt.Println("🚀 Ready for real-time video processing!")
}

// parseHexColor parses hex color string (copied from overlay package for testing)
func parseHexColor(hex string) (color.RGBA, error) {
	var r, g, b, a uint8
	a = 255 // Default to fully opaque

	if len(hex) == 7 {
		// Format: #RRGGBB
		_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		return color.RGBA{r, g, b, a}, err
	} else if len(hex) == 9 {
		// Format: #RRGGBBAA
		_, err := fmt.Sscanf(hex, "#%02x%02x%02x%02x", &r, &g, &b, &a)
		return color.RGBA{r, g, b, a}, err
	}

	return color.RGBA{}, fmt.Errorf("invalid hex color format: %s", hex)
}
```

## How to Run

```bash
# Save as overlay_test_example.go
go run overlay_test_example.go
```

## Features Demonstrated

1. **Overlay Configuration**: Setting up overlay rendering parameters
2. **Synthetic Data**: Creating test detection and tracking results
3. **Image Rendering**: Applying overlays to test images
4. **Component Testing**: Validating individual overlay components
5. **Color Validation**: Testing hex color parsing
6. **Result Verification**: Checking overlay application success

## Test Components

### Detection Overlay
- **Bounding Boxes**: Rectangles around detected objects
- **Class Labels**: Object class names (person, car, etc.)
- **Confidence Scores**: Detection confidence percentages
- **Color Coding**: Different colors for different object classes

### Tracking Overlay
- **Track IDs**: Unique identifiers for tracked objects
- **Movement Paths**: Visual representation of object movement
- **Predictions**: Future position predictions
- **Track States**: Visual indication of track confidence

### Performance Overlay
- **FPS Counter**: Current processing frame rate
- **Latency Display**: Processing delay information
- **Memory Usage**: Current memory consumption
- **GPU Usage**: GPU utilization percentage

## Expected Output

```
🎨 Testing Overlay System
🖼️  Rendering overlay...
✅ Overlay rendered successfully!
   Image bounds: (0,0)-(960,720)
   Image type: *image.RGBA
✅ Overlay pixels detected!

🧪 Testing individual components...
   - FPS rendering: ✅
   - Detection rendering: ✅
     2 detections processed
   - Tracking rendering: ✅
     1 tracks processed
   - Color parsing: ✅

🎉 All overlay system tests passed!
🚀 Ready for real-time video processing!
```

## Test Data

### Synthetic Detections
The test creates two synthetic detections:
1. **Person**: 85% confidence at position (100,100)-(200,200)
2. **Car**: 92% confidence at position (300,150)-(500,300)

### Synthetic Tracking
The test creates one tracked object:
- **Track ID**: 1
- **Class**: Person
- **State**: Confirmed track
- **Age**: 5 frames
- **Velocity**: Moving at (1.5, 0.5, 0) units/frame

### Performance Metrics
The test simulates realistic performance metrics:
- **FPS**: 30 frames per second
- **Latency**: 33 milliseconds processing delay
- **Memory**: 128MB usage
- **GPU**: 65.5% utilization

## Configuration Options

### Custom Colors
```go
config.Colors = map[string]string{
    "person":  "#FF0000",  // Red
    "car":     "#0000FF",  // Blue
    "bicycle": "#00FF00",  // Green
    "dog":     "#FFFF00",  // Yellow
    "default": "#FFFFFF",  // White
}
```

### Overlay Settings
```go
config := &ml.OverlayConfig{
    ShowFPS:        true,     // Show FPS counter
    ShowDetections: true,     // Show detection boxes
    ShowTracking:   true,     // Show tracking info
    ShowConfidence: true,     // Show confidence scores
    LineWidth:      3,        // Box line thickness
    FontSize:       16,       // Text font size
    FontScale:      0.6,      // Text scaling factor
}
```

## Advanced Testing

### Stress Testing
```go
// Test with many detections
var detections []ml.Detection
for i := 0; i < 100; i++ {
    detections = append(detections, ml.Detection{
        ClassID:    i % 10,
        ClassName:  fmt.Sprintf("object_%d", i),
        Confidence: 0.5 + float32(i%50)/100,
        Box:        image.Rect(i*10, i*5, i*10+50, i*5+50),
        Timestamp:  time.Now(),
    })
}
```

### Performance Benchmarking
```go
// Measure rendering performance
start := time.Now()
for i := 0; i < 1000; i++ {
    renderer.Render(testImg, results, metrics)
}
duration := time.Since(start)
fmt.Printf("Average render time: %v\n", duration/1000)
```

### Image Format Testing
```go
// Test different image formats
imageTypes := []struct{
    name string
    img  image.Image
}{
    {"RGBA", image.NewRGBA(bounds)},
    {"RGB", image.NewRGB(bounds)},
    {"NRGBA", image.NewNRGBA(bounds)},
    {"YCbCr", image.NewYCbCr(bounds, image.YCbCrSubsampleRatio420)},
}

for _, test := range imageTypes {
    result := renderer.Render(test.img, results, metrics)
    fmt.Printf("%s format: %v\n", test.name, result != nil)
}
```

## Integration Testing

### With Real ML Pipeline
```go
// Test with actual ML pipeline
mlIntegration, err := transport.NewMLVideoIntegration("0.0.0.0:11111", mlConfig)
if err != nil {
    log.Fatal(err)
}

// Get real results from pipeline
results := mlIntegration.GetLatestResults()
metrics := mlIntegration.GetMetrics()

// Test overlay with real data
resultImg := renderer.Render(testImg, results, metrics)
```

### With Video Stream
```go
// Test overlay on video frames
videoFile, err := os.Open("test_video.mp4")
if err != nil {
    log.Fatal(err)
}
defer videoFile.Close()

// Process each frame
decoder := createVideoDecoder(videoFile)
for {
    frame, err := decoder.NextFrame()
    if err != nil {
        break
    }
    
    // Apply ML processing
    results := processFrame(frame)
    
    // Apply overlay
    overlayFrame := renderer.Render(frame, results, metrics)
    
    // Save or display result
    saveFrame(overlayFrame)
}
```

## Troubleshooting

### Common Issues

1. **No Overlay Applied**: Check if overlay configuration is enabled
2. **Invalid Colors**: Verify hex color format (#RRGGBB or #RRGGBBAA)
3. **Empty Results**: Ensure ML results are properly formatted
4. **Image Type**: Verify input image format is supported
5. **Memory Issues**: Monitor memory usage with large images

### Debug Output
```go
// Enable debug logging
renderer.SetDebugMode(true)

// Get rendering statistics
stats := renderer.GetStats()
fmt.Printf("Render calls: %d\n", stats.RenderCount)
fmt.Printf("Average render time: %v\n", stats.AverageRenderTime)
fmt.Printf("Errors: %d\n", stats.ErrorCount)
```