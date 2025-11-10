# DJI Tello Go SDK

A comprehensive and easy-to-use Go SDK for the DJI Tello drone, featuring priority command queuing, real-time video streaming, and extensive telemetry capabilities.

## Features

- 🚁 **Complete Drone Control** - Takeoff, landing, movement, flips, and advanced flight patterns
- 🎮 **Gamepad Support** - Xbox, PlayStation, and generic USB controller support with customizable mappings
- 📹 **Video Streaming** - Real-time H.264 video stream processing and display
- 🎥 **Video Recording** - H.264 and MP4 recording with FFmpeg integration
- 🤖 **Machine Learning** - Real-time ML processing pipeline with YOLO object detection, face recognition, and gesture control
- 📊 **Telemetry Monitoring** - Real-time battery, altitude, attitude, and sensor data
- ⚡ **Priority Command Queue** - Intelligent command prioritization for responsive control
- 🖥️ **CLI Tool** - Comprehensive command-line interface (`telloctl`)
- 🌐 **Web Interface** - Browser-based video display and control
- 🔧 **Modular Architecture** - Clean, testable, and extensible design

## Installation

```bash
go get github.com/conceptcodes/dji-tello-sdk-go
```

### Dependencies

For video recording (MP4 format), install FFmpeg:
```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt-get install ffmpeg

# Windows
# Download from https://ffmpeg.org/download.html
```

For ML features (optional), install additional dependencies:
```bash
# macOS
brew install opencv

# Ubuntu/Debian
sudo apt-get install libopencv-dev

# Install GoCV
go get -u gocv.io/x/gocv
```

## Quick Start

```go
package main

import (
    "log"
    "time"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
    // Initialize the drone
    drone, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    // Enter SDK mode
    if err := drone.Init(); err != nil {
        log.Fatal(err)
    }

    // Take off
    if err := drone.TakeOff(); err != nil {
        log.Fatal(err)
    }

    // Fly in a square
    drone.Forward(100)
    drone.Right(100)
    drone.Backward(100)
    drone.Left(100)

    // Land
    if err := drone.Land(); err != nil {
        log.Fatal(err)
    }
}
```

## Architecture

### High Level Architecture

```mermaid
graph TD
    subgraph Application["Application Layer"]
        CLI[telloctl CLI]
        SDK[Go SDK Applications]
    end
    
    subgraph Core["SDK Core"]
        CMDR[Tello Commander]
        QUEUE[Priority Command Queue]
    end
    
    subgraph Transport["Transport Layer"]
        UDP_CMD[UDP Command Client]
        UDP_STATE[UDP State Listener]
        UDP_VIDEO[UDP Video Listener]
    end
    
    subgraph Processing["Processing Layer"]
        H264[H.264 Parser]
        REC[Video Recorder]
        GUI[Video Display]
        ML[ML Pipeline]
    end
    
    subgraph Hardware["Hardware"]
        DRONE[DJI Tello Drone]
    end
    
    CLI --> CMDR
    SDK --> CMDR
    CMDR --> QUEUE
    CMDR --> UDP_CMD
    CMDR --> UDP_STATE
    CMDR --> UDP_VIDEO
    UDP_VIDEO --> H264
    H264 --> REC
    H264 --> GUI
    H264 --> ML
    UDP_CMD --> DRONE
    DRONE --> UDP_STATE
    DRONE --> UDP_VIDEO
```

### Priority Command Queue

The SDK implements a priority-based command queue to ensure responsive drone control:

```mermaid
sequenceDiagram
    participant App as Application
    participant Queue as Priority Queue
    participant Drone as Tello Drone
    
    Note over App,Drone: High Priority Commands (Read Operations)
    App->>Queue: EnqueueRead("battery?")
    Queue->>Drone: Send "battery?"
    Drone-->>Queue: Response: "85"
    Queue-->>App: Battery: 85%
    
    Note over App,Drone: Low Priority Commands (Control Operations)
    App->>Queue: EnqueueControl("takeoff")
    Queue->>Drone: Send "takeoff"
    Drone-->>Queue: Response: "ok"
    Queue-->>App: Takeoff complete
```

### Video Streaming Pipeline

```mermaid
graph LR
    subgraph Drone["Drone"]
        CAM[Camera]
    end
    
    subgraph Transport["Transport"]
        UDP[UDP Port 11111]
    end
    
    subgraph Processing["Processing"]
        PARSER[H.264 Parser]
        BUFFER[Frame Buffer]
    end
    
    subgraph Output["Output"]
        REC[Recording]
        GUI[Display GUI]
        CHAN[Frame Channel]
    end
    
    CAM -->|H.264 Stream| UDP
    UDP --> PARSER
    PARSER -->|Parsed Frames| BUFFER
    BUFFER --> REC
    BUFFER --> GUI
    BUFFER --> CHAN
```

## Machine Learning

The SDK includes a comprehensive machine learning pipeline for real-time video analysis and intelligent drone control. The ML system supports object detection, face recognition, gesture control, and SLAM capabilities.

### ML Architecture

```mermaid
graph TD
    subgraph Input["Video Input"]
        STREAM[Video Stream]
        FRAME[Video Frames]
    end
    
    subgraph Pipeline["ML Pipeline"]
        ENHANCED[Enhanced Frames]
        QUEUE[Frame Queue]
        WORKERS[Worker Pool]
        RESULTS[ML Results]
    end
    
    subgraph Processors["ML Processors"]
        YOLO[YOLO Detection]
        FACE[Face Recognition]
        GESTURE[Gesture Control]
        SLAM_PROC[SLAM Processing]
    end
    
    subgraph Output["Output & Control"]
        OVERLAY[Video Overlay]
        DRONE_CTRL[Drone Control]
        TELEMETRY[ML Telemetry]
    end
    
    STREAM --> FRAME
    FRAME --> ENHANCED
    ENHANCED --> QUEUE
    QUEUE --> WORKERS
    WORKERS --> YOLO
    WORKERS --> FACE
    WORKERS --> GESTURE
    WORKERS --> SLAM_PROC
    YOLO --> RESULTS
    FACE --> RESULTS
    GESTURE --> RESULTS
    SLAM_PROC --> RESULTS
    RESULTS --> OVERLAY
    RESULTS --> DRONE_CTRL
    RESULTS --> TELEMETRY
```

### ML Features

- **🎯 Object Detection** - YOLO-based real-time object detection with multiple classes
- **👤 Face Recognition** - Face detection and tracking for autonomous following
- **👋 Gesture Control** - Hand gesture recognition for drone control
- **🗺️ SLAM** - Simultaneous Localization and Mapping for navigation
- **⚡ Real-time Processing** - Optimized for 15+ FPS performance
- **🔧 Plugin Architecture** - Extensible processor system
- **📊 Performance Metrics** - Built-in monitoring and analytics

### Quick Start with ML

```go
package main

import (
    "log"
    "time"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/config"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
    // Load ML configuration
    mlConfig, err := config.LoadFromFile("configs/ml-pipeline-default.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create ML video integration
    integration, err := transport.NewMLVideoIntegration("0.0.0.0:11111", mlConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Start ML processing
    if err := integration.Start(); err != nil {
        log.Fatal(err)
    }
    defer integration.Stop()

    // Process ML results
    go func() {
        for result := range integration.GetMLResults() {
            switch r := result.(type) {
            case *ml.DetectionResult:
                log.Printf("Detected %d objects", len(r.Detections))
                for _, detection := range r.Detections {
                    log.Printf("  - %s: %.2f confidence", 
                        detection.ClassName, detection.Confidence)
                }
            case *ml.GestureResult:
                log.Printf("Gesture: %s (%.2f confidence)", 
                    r.Gesture, r.Confidence)
            }
        }
    }()

    // Run for 30 seconds
    time.Sleep(30 * time.Second)
}
```

### ML Configuration

The ML system is configured through JSON files with schema validation:

```json
{
  "processors": [
    {
      "name": "yolo_detector",
      "type": "yolo",
      "enabled": true,
      "priority": 1,
      "config": {
        "model_path": "models/yolo.onnx",
        "confidence_threshold": 0.5,
        "nms_threshold": 0.4,
        "input_size": [640, 640],
        "classes": ["person", "car", "bicycle", "dog", "cat"]
      }
    },
    {
      "name": "face_detector",
      "type": "face",
      "enabled": true,
      "priority": 2,
      "config": {
        "model_path": "models/face.onnx",
        "confidence_threshold": 0.7,
        "tracking_enabled": true,
        "max_faces": 5
      }
    }
  ],
  "pipeline": {
    "max_concurrent_processors": 4,
    "frame_buffer_size": 100,
    "worker_pool_size": 2,
    "enable_metrics": true,
    "target_fps": 30
  },
  "overlay": {
    "enabled": true,
    "show_fps": true,
    "show_detections": true,
    "show_tracking": true,
    "show_confidence": true,
    "colors": {
      "person": "#00FF00",
      "car": "#FF0000",
      "face": "#0000FF"
    },
    "line_width": 2,
    "font_size": 12,
    "font_scale": 0.5
  }
}
```

### ML CLI Commands

The `telloctl` CLI includes comprehensive ML management commands:

```bash
# Initialize ML configuration
telloctl ml init

# List available processors
telloctl ml processors

# Validate ML configuration
telloctl ml validate --config /path/to/config.json

# Start ML processing with default config
telloctl ml start

# Start ML processing with custom config
telloctl ml start --config /path/to/config.json

# Monitor ML metrics
telloctl ml metrics

# Test ML processors
telloctl ml test --processor yolo
```

### Supported ML Processors

#### YOLO Object Detection
- **Models**: YOLOv5, YOLOv8, YOLO-NAS
- **Format**: ONNX Runtime
- **Classes**: Customizable (COCO, VOC, or custom)
- **Performance**: 15-30 FPS on GPU

#### Face Recognition
- **Detection**: Face detection with bounding boxes
- **Tracking**: Multi-face tracking with IDs
- **Recognition**: Face embedding and matching
- **Features**: Landmark detection, pose estimation

#### Gesture Control
- **Gestures**: Thumbs up, peace sign, pointing, waving
- **Real-time**: Low-latency gesture recognition
- **Actions**: Customizable drone actions per gesture
- **Training**: Support for custom gesture models

#### SLAM Processing
- **Visual Odometry**: Camera pose estimation
- **Mapping**: 3D environment reconstruction
- **Localization**: Position tracking without GPS
- **Features**: Keyframe extraction, loop closure

### Performance Optimization

The ML pipeline is optimized for real-time performance:

- **Concurrent Processing**: Multi-threaded frame processing
- **Memory Management**: Efficient frame buffer management
- **GPU Acceleration**: CUDA and OpenCL support
- **Adaptive Quality**: Dynamic resolution and frame rate adjustment
- **Resource Monitoring**: Built-in performance metrics

### ML Metrics and Monitoring

```go
// Get ML pipeline metrics
metrics := integration.GetMetrics()

fmt.Printf("FPS: %.2f\n", metrics.FPS)
fmt.Printf("Latency: %v\n", metrics.Latency)
fmt.Printf("Dropped Frames: %d\n", metrics.DroppedFrames)

for processor, stats := range metrics.ProcessorStats {
    fmt.Printf("%s: %.2fms avg processing time\n", processor, stats)
}
```

## Video Streaming

The SDK now supports real-time video streaming from DJI Tello drone. The video stream provides H.264 encoded video data that can be processed, saved, or displayed.

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
    // Initialize the Tello commander
    commander, err := tello.Initialize()
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }

    // Enter SDK mode
    if err := commander.Init(); err != nil {
        log.Fatalf("Failed to enter SDK mode: %v", err)
    }

    // Start video streaming
    if err := commander.StreamOn(); err != nil {
        log.Fatalf("Failed to start video stream: %v", err)
    }

    // Set up video frame callback
    commander.SetVideoFrameCallback(func(frame tello.VideoFrame) {
        fmt.Printf("Received frame: %d bytes, keyframe: %v\n", 
            frame.Size, frame.IsKeyFrame)
    })

    // Or use channel-based approach
    frameChan := commander.GetVideoFrameChannel()
    go func() {
        for frame := range frameChan {
            // Process video frame
            fmt.Printf("Frame %d: %d bytes\n", frame.SeqNum, frame.Size)
        }
    }()

    // Keep running
    time.Sleep(30 * time.Second)

    // Stop video streaming
    commander.StreamOff()
}
```

### Video Recording

```go
import "github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"

// Create a video recorder (H.264 format)
recorder, err := transport.NewVideoRecorder(":11111", "output.h264")
if err != nil {
    log.Fatal(err)
}

// Or create an MP4 video recorder
mp4Recorder, err := transport.NewVideoRecorderMP4(":11111", "output.mp4")
if err != nil {
    log.Fatal(err)
}

// Start recording
if err := recorder.StartRecording(); err != nil {
    log.Fatal(err)
}

// Record for 30 seconds
time.Sleep(30 * time.Second)

// Stop recording
recorder.StopRecording()
```

### CLI Commands

The `telloctl` CLI now includes video streaming commands:

```bash
# Start video streaming
telloctl streamon

# Stop video streaming  
telloctl streamoff

# Monitor video stream with statistics
telloctl stream

# Monitor for 60 seconds and save to file (H.264)
telloctl stream -d 60 -s video.h264

# Monitor for 60 seconds and save to MP4 file
telloctl stream -d 60 -s video.mp4 -f mp4

# Start video GUI (web interface)
telloctl video-gui

# Start video GUI (terminal interface)
telloctl video-gui -t terminal
```

#### ML Commands
```bash
# Initialize ML configuration
telloctl ml init

# List available ML processors
telloctl ml processors

# Validate ML configuration
telloctl ml validate

# Start ML processing
telloctl ml start

# Monitor ML metrics
telloctl ml metrics

# Test ML processors
telloctl ml test --processor yolo
```

#### Gamepad Commands
```bash
# List available gamepads
telloctl gamepad --list

# Start gamepad control with default configuration
telloctl gamepad

# Start gamepad control with custom configuration
telloctl gamepad --config /path/to/config.json

# Start gamepad control with preset (xbox, playstation, default)
telloctl gamepad --preset xbox
```

### Video Frame Structure

Each video frame contains:

- `Data`: Raw H.264 video data
- `Timestamp`: When the frame was received
- `Size`: Frame size in bytes
- `SeqNum`: Frame sequence number
- `NALUnits`: Parsed H.264 NAL units
- `IsKeyFrame`: Whether the frame contains a keyframe

### H.264 Parsing

The SDK includes H.264 parsing capabilities:

```go
parser := transport.NewH264Parser()
nalUnits, err := parser.ParseFrame(frame.Data)
if err != nil {
    log.Printf("Failed to parse frame: %v", err)
    return
}

// Check for keyframes
hasKeyFrame := parser.HasKeyFrame(nalUnits)

// Get frame information
info := parser.GetFrameInfo(nalUnits)
fmt.Printf("NAL units: %v\n", info["nal_types"])
```

## CLI Tool

The SDK includes a comprehensive CLI tool `telloctl` for drone control and monitoring.

### Installation

```bash
go install github.com/conceptcodes/dji-tello-sdk-go/cmd/telloctl@latest
```

### Commands

#### Control Commands
```bash
# Initialize drone
telloctl init

# Takeoff and landing
telloctl takeoff
telloctl land

# Movement
telloctl up 50        # Move up 50cm
telloctl forward 100  # Move forward 100cm
telloctl clockwise 90 # Rotate 90 degrees clockwise

# Emergency stop
telloctl emergency
```

#### Telemetry Commands
```bash
# Get battery level
telloctl battery

# Get height
telloctl height

# Get attitude (pitch, roll, yaw)
telloctl attitude

# Monitor all telemetry
telloctl telemetry
```

#### Video Commands
```bash
# Start/stop video streaming
telloctl streamon
telloctl streamoff

# Monitor video stream
telloctl stream -d 30 -s output.mp4 -f mp4

# Start video GUI
telloctl video-gui -t web -p 8080
```

## Gamepad Support

The SDK includes comprehensive gamepad support for controlling DJI Tello drones with physical controllers. It supports Xbox, PlayStation, and generic USB controllers with fully customizable button and axis mappings.

### Supported Controllers

- **Xbox Controllers** - Xbox One, Xbox Series X/S, and compatible controllers
- **PlayStation Controllers** - DualShock 4, DualSense, and compatible controllers  
- **Generic USB Controllers** - Any standard USB gamepad with SDL2 support

### Quick Start

```bash
# List available gamepads
telloctl gamepad --list

# Start gamepad control with default configuration
telloctl gamepad

# Use Xbox preset configuration
telloctl gamepad --preset xbox
```

### Configuration

Gamepad behavior is controlled through JSON configuration files with schema validation:

```json
{
  "version": "1.0.0",
  "controller": {
    "deadzone": 0.15,
    "sensitivity": 1.0,
    "update_rate": 60,
    "auto_detect": true
  },
  "safety": {
    "rc_limits": {
      "horizontal": 50,
      "vertical": 50,
      "yaw": 50
    },
    "emergency_actions": {
      "connection_timeout": 5000,
      "low_battery_threshold": 20,
      "enable_auto_land": true
    }
  },
  "mappings": {
    "axes": {
      "left_stick_x": {
        "axis": "left_stick_x",
        "invert": false,
        "deadzone": 0.1,
        "rc_mapping": "roll"
      },
      "left_stick_y": {
        "axis": "left_stick_y", 
        "invert": true,
        "deadzone": 0.1,
        "rc_mapping": "throttle"
      }
    },
    "buttons": {
      "button_a": {
        "button": "button_a",
        "action": "takeoff"
      },
      "button_b": {
        "button": "button_b",
        "action": "land"
      }
    }
  }
}
```

### Default Control Mappings

#### Xbox Controller
- **Left Stick**: Throttle (Y) / Yaw (X)
- **Right Stick**: Pitch (Y) / Roll (X)
- **A Button**: Takeoff
- **B Button**: Land
- **X Button**: Emergency Stop
- **Y Button**: Flip Forward
- **Left Bumper**: Flip Left
- **Right Bumper**: Flip Right
- **Start Button**: Start Video Streaming
- **Select Button**: Stop Video Streaming

#### PlayStation Controller
- **Left Stick**: Throttle (Y) / Yaw (X)
- **Right Stick**: Pitch (Y) / Roll (X)
- **Cross Button**: Takeoff
- **Circle Button**: Land
- **Square Button**: Emergency Stop
- **Triangle Button**: Flip Forward
- **L1 Button**: Flip Left
- **R1 Button**: Flip Right
- **Options Button**: Start Video Streaming
- **Share Button**: Stop Video Streaming

### Programmatic Usage

```go
package main

import (
    "log"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/gamepad"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
    // Initialize drone
    drone, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    // Load gamepad configuration
    config, err := gamepad.LoadConfigFromFile("configs/gamepad-default.json")
    if err != nil {
        config = gamepad.DefaultConfig()
    }

    // Create gamepad handler
    handler, err := gamepad.NewHandler(gamepad.HandlerOptions{
        Config: config,
        OnRCValues: func(rcValues gamepad.RCValues) {
            // Send RC values to drone
            drone.SetRcControl(rcValues.A, rcValues.B, rcValues.C, rcValues.D)
        },
        OnDroneAction: func(action gamepad.DroneAction) {
            // Handle discrete actions
            switch action {
            case gamepad.ActionTakeoff:
                drone.TakeOff()
            case gamepad.ActionLand:
                drone.Land()
            case gamepad.ActionEmergency:
                drone.Emergency()
            }
        },
        OnError: func(err error) {
            log.Printf("Gamepad error: %v", err)
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer handler.Stop()

    // Start gamepad processing
    if err := handler.Start(); err != nil {
        log.Fatal(err)
    }

    // Keep running
    select {}
}
```

### Example Application

```go
// Run the gamepad example
go run ./examples/gamepad/main.go
```

This example provides:
- Real-time gamepad state display
- RC value monitoring
- Button press detection
- Controller connection status

## API Reference

### Core Interface

```go
type TelloCommander interface {
    // Control Commands
    Init() error
    TakeOff() error
    Land() error
    StreamOn() error
    StreamOff() error
    Emergency() error
    
    // Movement Commands
    Up(distance int) error
    Down(distance int) error
    Left(distance int) error
    Right(distance int) error
    Forward(distance int) error
    Backward(distance int) error
    Clockwise(angle int) error
    CounterClockwise(angle int) error
    Flip(direction FlipDirection) error
    Go(x, y, z, speed int) error
    Curve(x1, y1, z1, x2, y2, z2, speed int) error
    
    // Settings Commands
    SetSpeed(speed int) error
    SetRcControl(a, b, c, d int) error
    SetWiFiCredentials(ssid, password string) error
    
    // Read Commands
    GetSpeed() (int, error)
    GetBatteryPercentage() (int, error)
    GetTime() (int, error)
    GetHeight() (int, error)
    GetTemperature() (int, error)
    GetAttitude() (int, int, int, error)
    GetBarometer() (int, error)
    GetAcceleration() (int, int, int, error)
    GetTof() (int, error)
    
    // Video Commands
    SetVideoFrameCallback(callback VideoFrameCallback)
    GetVideoFrameChannel() <-chan transport.VideoFrame
}
```

## Examples

### Basic Flight Control

```go
package main

import (
    "log"
    "time"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
    drone, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    drone.Init()
    drone.TakeOff()
    
    // Fly a square pattern
    drone.Forward(100)
    drone.Right(100)
    drone.Backward(100)
    drone.Left(100)
    
    drone.Land()
}
```

### Video Recording

```go
package main

import (
    "log"
    "time"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
    drone, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    drone.Init()
    drone.StreamOn()

    // Create MP4 recorder
    recorder, err := transport.NewVideoRecorderMP4(":11111", "flight.mp4")
    if err != nil {
        log.Fatal(err)
    }

    recorder.StartRecording()
    
    // Record during flight
    drone.TakeOff()
    time.Sleep(2 * time.Second)
    drone.Forward(100)
    time.Sleep(2 * time.Second)
    drone.Land()
    
    recorder.StopRecording()
    drone.StreamOff()
}
```

### Telemetry Monitoring

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
    drone, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    drone.Init()

    // Monitor telemetry
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for i := 0; i < 30; i++ {
        select {
        case <-ticker.C:
            battery, _ := drone.GetBatteryPercentage()
            height, _ := drone.GetHeight()
            pitch, roll, yaw, _ := drone.GetAttitude()
            
            fmt.Printf("Battery: %d%%, Height: %dcm, Attitude: P:%d R:%d Y:%d\n",
                battery, height, pitch, roll, yaw)
        }
    }
}
```

## Development

### Building

```bash
# Build CLI tool
go build -o telloctl ./cmd/telloctl

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Project Structure

```
.
├── cmd/
│   └── telloctl/           # CLI application
│       └── commands/       # CLI command implementations
├── pkg/
│   ├── tello/             # Core SDK functionality
│   │   ├── commander.go   # Main drone interface
│   │   └── priority_command_queue.go  # Command queuing
│   ├── gamepad/           # Gamepad support
│   │   ├── handler.go     # Gamepad input processing
│   │   ├── config.go      # Configuration management
│   │   ├── types.go       # Type definitions
│   │   └── defaults.go    # Preset configurations
│   ├── transport/         # Communication layer
│   │   ├── video.go       # Video streaming
│   │   ├── h264_parser.go # H.264 parsing
│   │   ├── mp4_recorder.go # Video recording
│   │   ├── ml_video_integration.go # ML integration
│   │   └── state.go       # Telemetry
│   ├── ml/                # Machine learning pipeline
│   │   ├── types.go       # ML data structures
│   │   ├── processors/    # ML processor interfaces
│   │   ├── pipeline/      # Concurrent processing
│   │   ├── config/        # Configuration management
│   │   └── overlay/       # Visual result rendering
│   └── utils/             # Utilities
└── examples/              # Example applications
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] Video streaming support
- [x] Priority command queuing
- [x] MP4 recording with FFmpeg
- [x] Web-based video GUI
- [x] Gamepad support
- [x] Machine learning pipeline foundation
- [ ] YOLO object detection processor
- [ ] Face recognition and tracking
- [ ] Gesture control system
- [ ] SLAM navigation
- [ ] Swarm manager
- [ ] Flight path planning
- [ ] Advanced telemetry analytics

## References

- [DJI Tello 1.3 SDK Documentation](https://dl-cdn.ryzerobotics.com/downloads/tello/20180910/Tello%20SDK%20Documentation%20EN_1.3.pdf)
- [H.264/MPEG-4 AVC Video Compression Standard](https://www.itu.int/rec/T-REC-H.264)