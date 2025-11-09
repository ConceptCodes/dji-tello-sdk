# Video Streaming Examples

This document provides examples of how to use the video streaming functionality in the DJI Tello Go SDK.

## Basic Video Streaming

See `video_streaming_example.go` for a complete example that demonstrates:

- Initializing the drone and entering SDK mode
- Starting video streaming
- Monitoring video frames with statistics
- Graceful shutdown

## Running the Video Example

```bash
cd examples
go run video_streaming_example.go
```

## Video Recording Example

```go
package main

import (
    "log"
    "time"
    
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
    // Initialize drone
    commander, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    // Enter SDK mode
    if err := commander.Init(); err != nil {
        log.Fatal(err)
    }

    // Start video stream
    if err := commander.StreamOn(); err != nil {
        log.Fatal(err)
    }

    // Create video recorder
    recorder, err := transport.NewVideoRecorder(":11111", "tello_video.h264")
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
    recorder.Close()

    // Stop video stream
    commander.StreamOff()
}
```

## Video Frame Processing

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
    "github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
    commander, err := tello.Initialize()
    if err != nil {
        log.Fatal(err)
    }

    if err := commander.Init(); err != nil {
        log.Fatal(err)
    }

    if err := commander.StreamOn(); err != nil {
        log.Fatal(err)
    }

    // Set up frame callback
    commander.SetVideoFrameCallback(func(frame tello.VideoFrame) {
        // Process frame
        if frame.IsKeyFrame {
            fmt.Printf("Keyframe received: %d bytes\n", frame.Size)
        }

        // Parse NAL units
        parser := transport.NewH264Parser()
        nalUnits, err := parser.ParseFrame(frame.Data)
        if err != nil {
            log.Printf("Parse error: %v", err)
            return
        }

        // Print NAL unit types
        for _, nalUnit := range nalUnits {
            fmt.Printf("NAL: %s\n", parser.GetNALUTypeName(nalUnit.Type))
        }
    })

    // Keep running
    select {}
}
```

## CLI Video Commands

The `telloctl` tool provides video streaming commands:

```bash
# Start video stream
telloctl streamon

# Stop video stream
telloctl streamoff

# Monitor video stream
telloctl stream

# Record video for 60 seconds
telloctl stream -d 60 -s my_video.h264
```

## Video File Playback

The recorded H.264 files can be played with:

- VLC Media Player
- FFmpeg: `ffplay tello_video.h264`
- Most modern video players that support H.264

## Troubleshooting

### No Video Frames Received

1. Ensure the drone is connected and in SDK mode
2. Check that `streamon` command was sent successfully
3. Verify firewall allows UDP traffic on port 11111
4. Make sure no other application is using port 11111

### Video Quality Issues

1. Check WiFi signal strength
2. Reduce distance between drone and controller
3. Minimize WiFi interference
4. Ensure adequate lighting conditions

### Performance Tips

1. Process frames asynchronously to avoid blocking
2. Use frame dropping for real-time applications
3. Consider buffering frames for smooth playback
4. Monitor frame rate and adjust processing accordingly