# DJI Tello Go SDK

I wanted to create a simple and easy-to-use SDK for the DJI Tello drone. The goal is to provide a simple interface for 
- controlling the drone 
- getting telemetry data
- taking pictures / streaming video

The DJI Tello drone is a small, lightweight drone that is beginner friendly and has a lot of features. 

[DJI Tello 1.3 Docs](https://dl-cdn.ryzerobotics.com/downloads/tello/20180910/Tello%20SDK%20Documentation%20EN_1.3.pdf)


## High Level Architecture
```mermaid
```

### Safety Manager
```mermaid
```

### Transport Layer
```mermaid
```

### Commands
```mermaid
```

### Telemetry
```mermaid
```

### Video Streaming

The SDK now supports real-time video streaming from DJI Tello drone. The video stream provides H.264 encoded video data that can be processed, saved, or displayed.

#### Basic Usage

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

#### Video Recording

```go
import "github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"

// Create a video recorder
recorder, err := transport.NewVideoRecorder(":11111", "output.h264")
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

#### CLI Commands

The `telloctl` CLI now includes video streaming commands:

```bash
# Start video streaming
telloctl streamon

# Stop video streaming  
telloctl streamoff

# Monitor video stream with statistics
telloctl stream

# Monitor for 60 seconds and save to file
telloctl stream -d 60 -s video.h264
```

#### Video Frame Structure

Each video frame contains:

- `Data`: Raw H.264 video data
- `Timestamp`: When the frame was received
- `Size`: Frame size in bytes
- `SeqNum`: Frame sequence number
- `NALUnits`: Parsed H.264 NAL units
- `IsKeyFrame`: Whether the frame contains a keyframe

#### H.264 Parsing

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


## Roadmap
- [x] Video streaming support
- [ ] Add gamepad support
- [ ] Basic ML support
- [ ] Swarm manager