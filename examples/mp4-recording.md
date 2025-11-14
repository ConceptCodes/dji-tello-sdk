# MP4 Recording Example

This example demonstrates how to record video from a DJI Tello drone to MP4 format using FFmpeg.

## Overview

The MP4 recording example shows:
- Connecting to a Tello drone and starting video stream
- Setting up MP4 video recording with FFmpeg
- Recording video for a specified duration
- Proper cleanup and file output

## Code

```go
package main

import (
	"fmt"
	"log"
	"time"

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

	// Create MP4 video recorder
	recorder, err := transport.NewVideoRecorderMP4(":11111", "test_output.mp4")
	if err != nil {
		log.Fatalf("Failed to create MP4 recorder: %v", err)
	}

	// Start recording
	if err := recorder.StartRecording(); err != nil {
		log.Fatalf("Failed to start recording: %v", err)
	}

	fmt.Println("Recording MP4 video for 10 seconds...")
	fmt.Println("Make sure FFmpeg is installed on your system")

	// Record for 10 seconds
	time.Sleep(10 * time.Second)

	// Stop recording
	if err := recorder.StopRecording(); err != nil {
		log.Fatalf("Failed to stop recording: %v", err)
	}

	recorder.Close()

	// Stop video stream
	if err := drone.StreamOff(); err != nil {
		log.Printf("Warning: Failed to stop video stream: %v", err)
	}

	fmt.Println("MP4 recording completed! Check test_output.mp4")
}
```

## How to Run

```bash
# Save as mp4_recording_example.go
go run mp4_recording_example.go
```

## Prerequisites

### FFmpeg Installation

**macOS:**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows:**
Download from https://ffmpeg.org/download.html or use package manager like Chocolatey:
```bash
choco install ffmpeg
```

## Features Demonstrated

1. **Drone Connection**: Connects to Tello drone and initializes SDK mode
2. **Video Streaming**: Starts H.264 video stream from drone
3. **MP4 Recording**: Records video stream to MP4 file using FFmpeg
4. **Duration Control**: Records for specified time period
5. **Proper Cleanup**: Ensures clean shutdown and file saving

## Recording Process

1. **Initialization**: Connect to drone and start video stream
2. **Recorder Setup**: Create MP4 recorder with output filename
3. **Recording Start**: Begin capturing video to file
4. **Duration Control**: Record for specified time (10 seconds)
5. **Recording Stop**: Stop recording and save file
6. **Cleanup**: Close recorder and stop video stream

## Expected Output

```
Recording MP4 video for 10 seconds...
Make sure FFmpeg is installed on your system
MP4 recording completed! Check test_output.mp4
```

## Output File

The example creates `test_output.mp4` in the current directory containing:
- H.264 video encoded from drone stream
- AAC audio (if audio is available)
- MP4 container format
- Original video resolution and frame rate

## Configuration Options

### Custom Recording Duration

```go
// Change recording duration
recordingDuration := 30 * time.Second // 30 seconds
time.Sleep(recordingDuration)
```

### Custom Output Filename

```go
// Use timestamp in filename
timestamp := time.Now().Format("20060102_150405")
filename := fmt.Sprintf("tello_recording_%s.mp4", timestamp)
recorder, err := transport.NewVideoRecorderMP4(":11111", filename)
```

### Custom Video Quality

```go
// Configure recording parameters (if supported)
recorder, err := transport.NewVideoRecorderMP4WithOptions(":11111", "output.mp4", transport.RecorderOptions{
    VideoBitrate: "2000k",
    AudioBitrate: "128k",
    FrameRate:    30,
    Resolution:   "960x720",
})
```

## Advanced Usage

### Recording with Telemetry Overlay

```go
// Create recorder with telemetry overlay
recorder, err := transport.NewVideoRecorderWithOverlay(":11111", "output_with_telemetry.mp4", transport.OverlayOptions{
    ShowAltitude:    true,
    ShowSpeed:       true,
    ShowBattery:     true,
    ShowTimestamp:   true,
    FontSize:        24,
    Position:        "bottom-left",
})
```

### Multiple Recording Segments

```go
// Record in segments for easier file management
segmentDuration := 60 * time.Second // 1 minute per segment
totalDuration := 5 * time.Minute    // 5 minutes total

for start := time.Duration(0); start < totalDuration; start += segmentDuration {
    filename := fmt.Sprintf("segment_%d.mp4", start/segmentDuration)
    recorder, err := transport.NewVideoRecorderMP4(":11111", filename)
    if err != nil {
        log.Printf("Failed to create segment recorder: %v", err)
        continue
    }
    
    recorder.StartRecording()
    time.Sleep(segmentDuration)
    recorder.StopRecording()
    recorder.Close()
}
```

### Recording with Callbacks

```go
// Set up recording callbacks
recorder, err := transport.NewVideoRecorderMP4WithCallbacks(":11111", "output.mp4", transport.RecorderCallbacks{
    OnStart: func() {
        log.Println("Recording started")
    },
    OnStop: func() {
        log.Println("Recording stopped")
    },
    OnError: func(err error) {
        log.Printf("Recording error: %v", err)
    },
    OnProgress: func(duration time.Duration) {
        log.Printf("Recording duration: %v", duration)
    },
})
```

## Troubleshooting

### FFmpeg Not Found
```
Error: FFmpeg not found in PATH
```
**Solution**: Install FFmpeg and ensure it's in your system PATH

### Permission Issues
```
Error: Permission denied when creating output file
```
**Solution**: Check write permissions in the output directory

### Network Issues
```
Error: Failed to bind to UDP port 11111
```
**Solution**: Ensure port 11111 is not in use by another application

### Drone Connection Issues
```
Error: Failed to connect to drone
```
**Solution**: Check drone WiFi connection and ensure drone is powered on

## Performance Considerations

### Disk Space
- 1 minute of 960x720 video ≈ 10-15 MB
- Ensure sufficient disk space for planned recording duration
- Consider compression settings for longer recordings

### CPU Usage
- FFmpeg encoding requires CPU resources
- Monitor system performance during recording
- Consider lower quality settings for older hardware

### Network Bandwidth
- Video streaming requires stable WiFi connection
- Monitor signal strength during recording
- Consider reducing video quality if connection is unstable

## File Management

### Automatic Cleanup
```go
// Clean up old recordings after 24 hours
func cleanupOldRecordings(directory string, maxAge time.Duration) {
    files, err := os.ReadDir(directory)
    if err != nil {
        return
    }
    
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        
        info, err := file.Info()
        if err != nil {
            continue
        }
        
        if time.Since(info.ModTime()) > maxAge {
            os.Remove(filepath.Join(directory, file.Name()))
        }
    }
}
```

### Recording Statistics
```go
// Get recording statistics
stats := recorder.GetStats()
fmt.Printf("Frames recorded: %d\n", stats.FramesRecorded)
fmt.Printf("Duration: %v\n", stats.Duration)
fmt.Printf("File size: %d bytes\n", stats.FileSize)
fmt.Printf("Average bitrate: %f kbps\n", stats.AverageBitrate)
```