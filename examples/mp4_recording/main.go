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
