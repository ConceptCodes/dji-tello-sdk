# Video GUI Example

This example demonstrates how to create a web-based video GUI for displaying live video from a DJI Tello drone.

## Overview

The video GUI example shows:
- Connecting to a Tello drone and starting video stream
- Creating a web-based video display interface
- Setting up FPS overlay on video feed
- Configuring video display options
- Simple and responsive web interface

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

	// Get video frame channel
	frameChan := drone.GetVideoFrameChannel()
	if frameChan == nil {
		log.Fatalf("Failed to get video frame channel")
	}

	// Create ML configuration with overlay enabled (optional)
	mlConfig := &ml.MLConfig{
		Overlay: ml.OverlayConfig{
			Enabled:        true,
			ShowFPS:        true,
			ShowDetections: false, // Disabled for basic video GUI
			ShowTracking:   false, // Disabled for basic video GUI
			ShowConfidence: false, // Disabled for basic video GUI
			Colors: map[string]string{
				"default": "#00FF00",
			},
			LineWidth: 2,
			FontSize:  12,
			FontScale: 0.5,
		},
	}

	// Create video display (web interface)
	display := transport.NewVideoDisplay(transport.DisplayTypeWeb)
	display.SetVideoChannel(frameChan)
	display.SetMLConfig(mlConfig) // Set overlay configuration
	display.SetWebPort(8080)

	// Start display
	if err := display.Start(); err != nil {
		log.Fatalf("Failed to start video display: %v", err)
	}
	defer display.Close()

	log.Println("🎥 Video GUI started!")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("📊 FPS overlay enabled")
	log.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Stopping video GUI...")
}
```

## How to Run

```bash
# Save as video_gui_example.go
go run video_gui_example.go
```

## Features Demonstrated

1. **Drone Connection**: Connects to Tello drone and initializes SDK mode
2. **Video Streaming**: Starts H.264 video stream from drone
3. **Web Interface**: Creates browser-based video display
4. **FPS Overlay**: Shows real-time frame rate on video
5. **Responsive Design**: Adapts to different screen sizes
6. **Graceful Shutdown**: Proper cleanup on exit

## Web Interface Features

### Main Video Display
- **Live Video Feed**: Real-time H.264 video stream
- **Full Screen Mode**: Expand video to full browser window
- **Responsive Layout**: Adapts to mobile and desktop screens
- **Low Latency**: Optimized for real-time viewing

### Overlay Information
- **FPS Counter**: Current video frame rate
- **Connection Status**: Drone connection indicator
- **Video Resolution**: Display resolution information
- **Timestamp**: Current time overlay

### Control Panel
- **Start/Stop**: Control video streaming
- **Quality Settings**: Adjust video quality if supported
- **Screenshot**: Capture still images from video
- **Fullscreen**: Toggle fullscreen viewing mode

## Expected Output

```
🎥 Video GUI started!
🌐 Open http://localhost:8080 in your browser
📊 FPS overlay enabled
Press Ctrl+C to stop
```

The web interface will display:
- **Main Panel**: Live video feed from drone
- **Status Bar**: FPS counter and connection status
- **Control Buttons**: Basic video controls
- **Settings Panel**: Configuration options

## Web Interface Structure

```
http://localhost:8080/
├── Video Container (Main)
│   ├── Video Feed
│   ├── FPS Overlay
│   └── Status Indicators
├── Control Panel (Bottom)
│   ├── Play/Pause
│   ├── Screenshot
│   └── Fullscreen
└── Settings Panel (Side)
    ├── Video Quality
    ├── Overlay Options
    └── Display Settings
```

## Configuration Options

### Basic Configuration
```go
// Custom web port
display.SetWebPort(9090) // Use port 9090 instead of 8080

// Custom overlay settings
mlConfig := &ml.MLConfig{
    Overlay: ml.OverlayConfig{
        ShowFPS:        true,
        ShowTimestamp:  true,
        ShowResolution: true,
        FontSize:       16,
        FontScale:      0.6,
        Colors: map[string]string{
            "fps": "#00FF00",  // Green FPS counter
            "text": "#FFFFFF",  // White text
        },
    },
}
```

### Advanced Configuration
```go
// Display options
display.SetDisplayOptions(transport.DisplayOptions{
    AutoPlay:      true,
    Muted:         false,
    Controls:      false,
    Loop:          true,
    Responsive:    true,
    MaintainAspect: true,
})

// Performance settings
display.SetPerformanceSettings(transport.PerformanceSettings{
    BufferSize:     30,
    MaxLatency:     100 * time.Millisecond,
    TargetFPS:      30,
    Quality:        "high",
    AdaptiveQuality: true,
})
```

## Advanced Features

### Multiple Video Sources
```go
// Support multiple drones
displays := make([]*transport.VideoDisplay, numDrones)
for i := 0; i < numDrones; i++ {
    drone := initializeDrone(i)
    frameChan := drone.GetVideoFrameChannel()
    
    display := transport.NewVideoDisplay(transport.DisplayTypeWeb)
    display.SetVideoChannel(frameChan)
    display.SetWebPort(8080 + i)
    display.Start()
    
    displays[i] = display
}
```

### Recording Integration
```go
// Add recording capability
recorder := transport.NewVideoRecorderMP4(":11111", "output.mp4")
display.SetRecorder(recorder)

// Recording controls
display.AddControl("record", transport.ControlButton{
    Label:    "Record",
    Action:   recorder.StartRecording,
    OnStop:   recorder.StopRecording,
    Icon:     "record",
    Position: "top-right",
})
```

### Telemetry Integration
```go
// Add telemetry overlay
telemetrySource := drone.GetTelemetryChannel()
display.SetTelemetrySource(telemetrySource)

// Configure telemetry display
display.SetTelemetryConfig(transport.TelemetryConfig{
    ShowAltitude:    true,
    ShowSpeed:       true,
    ShowBattery:     true,
    ShowGPS:         false,
    Position:        "bottom-left",
    FontSize:        14,
    Background:      "rgba(0,0,0,0.7)",
    TextColor:       "#00FF00",
})
```

## Custom Web Interface

### Custom HTML Template
```go
// Use custom template
display.SetTemplatePath("templates/custom_video.html")

// Custom CSS
display.SetStylesheetPath("static/css/custom.css")

// Custom JavaScript
display.SetScriptPath("static/js/custom.js")
```

### WebSocket Integration
```go
// Add real-time communication
display.SetWebSocketHandler(func(ws *websocket.Conn) {
    for {
        // Handle client messages
        messageType, message, err := ws.ReadMessage()
        if err != nil {
            break
        }
        
        // Process commands
        handleClientCommand(messageType, message)
    }
})
```

## Performance Optimization

### Video Quality Settings
```go
// Optimize for different network conditions
display.SetQualityProfile(transport.QualityProfile{
    Name:         "low_bandwidth",
    Resolution:    "640x480",
    FrameRate:     15,
    Bitrate:       "500k",
    KeyFrameInterval: 30,
})
```

### Buffer Management
```go
// Configure buffering for smooth playback
display.SetBufferConfig(transport.BufferConfig{
    MinBufferSize: 5,
    MaxBufferSize: 30,
    TargetLatency: 50 * time.Millisecond,
    AdaptiveBuffering: true,
})
```

## Troubleshooting

### Common Issues

1. **Video Not Displaying**: Check drone connection and video stream
2. **High Latency**: Reduce video quality or check network bandwidth
3. **Browser Compatibility**: Use modern browser with WebRTC support
4. **Port Conflicts**: Change web port if 8080 is in use
5. **Permission Issues**: Ensure network access permissions

### Debug Mode
```go
// Enable debug logging
display.SetDebugMode(true)

// Get performance statistics
stats := display.GetStats()
fmt.Printf("Frames processed: %d\n", stats.FramesProcessed)
fmt.Printf("Average FPS: %.2f\n", stats.AverageFPS)
fmt.Printf("Dropped frames: %d\n", stats.DroppedFrames)
fmt.Printf("Network latency: %v\n", stats.NetworkLatency)
```

### Network Issues
```go
// Monitor network quality
display.SetNetworkMonitor(func(stats transport.NetworkStats) {
    if stats.PacketLoss > 0.05 { // 5% packet loss
        log.Printf("High packet loss: %.2f%%\n", stats.PacketLoss*100)
    }
    
    if stats.Latency > 200*time.Millisecond {
        log.Printf("High latency: %v\n", stats.Latency)
    }
})
```

## Browser Support

### Recommended Browsers
- **Chrome 80+**: Full feature support
- **Firefox 75+**: Good compatibility
- **Safari 13+**: Basic support
- **Edge 80+**: Full feature support

### Mobile Support
- **iOS Safari 13+**: Basic video playback
- **Chrome Mobile**: Full support
- **Firefox Mobile**: Good compatibility

## Security Considerations

### Network Security
```go
// Enable HTTPS
display.SetTLSConfig(&tls.Config{
    Certificates: []tls.Certificate{cert},
    MinVersion:   tls.VersionTLS12,
})

// Add authentication
display.SetAuthHandler(func(username, password string) bool {
    return validateCredentials(username, password)
})
```

### Access Control
```go
// Restrict access to local network
display.SetAllowedOrigins([]string{
    "http://localhost:8080",
    "http://192.168.1.*",
    "http://10.0.0.*",
})
```