// Example: MP4 video recording during flight
// This example demonstrates:
// - Starting video streaming
// - Recording to MP4 format using FFmpeg
// - Flying patterns while recording
// - Proper cleanup and resource management

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize the drone
	drone, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize drone: %v", err)
	}

	// Enter SDK mode
	fmt.Println("Entering SDK mode...")
	if err := drone.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Start video streaming
	fmt.Println("Starting video streaming...")
	if err := drone.StreamOn(); err != nil {
		log.Fatalf("Failed to start video streaming: %v", err)
	}

	// Give the video stream time to stabilize
	time.Sleep(2 * time.Second)

	// Create MP4 video recorder
	outputFile := fmt.Sprintf("flight_recording_%s.mp4", time.Now().Format("20060102_150405"))
	fmt.Printf("Creating MP4 recorder: %s\n", outputFile)
	recorder, err := transport.NewVideoRecorderMP4(":11111", outputFile)
	if err != nil {
		log.Fatalf("Failed to create MP4 recorder: %v", err)
	}

	// Start recording
	fmt.Println("Starting recording...")
	if err := recorder.StartRecording(); err != nil {
		log.Fatalf("Failed to start recording: %v", err)
	}

	// Wait for recording to initialize
	time.Sleep(1 * time.Second)

	// Take off
	fmt.Println("Taking off...")
	if err := drone.TakeOff(); err != nil {
		log.Fatalf("Takeoff failed: %v", err)
	}

	// Set up interrupt handler
	go func() {
		<-sigChan
		fmt.Println("\nInterrupt received, stopping...")
		if err := drone.Land(); err != nil {
			log.Printf("Emergency landing failed: %v", err)
		}
		recorder.StopRecording()
		drone.StreamOff()
		os.Exit(0)
	}()

	// Fly patterns while recording
	fmt.Println("Flying patterns...")

	// Pattern 1: Square
	flySquare(drone)
	time.Sleep(1 * time.Second)

	// Pattern 2: Ascending spiral
	flyAscendingSpiral(drone)
	time.Sleep(1 * time.Second)

	// Pattern 3: Hover and rotate
	hoverAndRotate(drone)

	// Land
	fmt.Println("Landing...")
	if err := drone.Land(); err != nil {
		log.Fatalf("Landing failed: %v", err)
	}

	// Wait for landing to complete
	time.Sleep(2 * time.Second)

	// Stop recording
	fmt.Println("Stopping recording...")
	recorder.StopRecording()

	// Stop video streaming
	fmt.Println("Stopping video streaming...")
	if err := drone.StreamOff(); err != nil {
		log.Printf("Failed to stop video streaming: %v", err)
	}

	fmt.Printf("\nRecording saved to: %s\n", outputFile)
	fmt.Println("Example completed successfully")
}

// flySquare flies a square pattern while recording
func flySquare(drone tello.TelloCommander) {
	fmt.Println("Flying square pattern...")
	sideLength := 100

	directions := []struct {
		name string
		fn   func(int) error
		dist int
	}{
		{"Forward", drone.Forward, sideLength},
		{"Right", drone.Right, sideLength},
		{"Backward", drone.Backward, sideLength},
		{"Left", drone.Left, sideLength},
	}

	for _, dir := range directions {
		fmt.Printf("  Moving %s %d cm...\n", dir.name, dir.dist)
		if err := dir.fn(dir.dist); err != nil {
			log.Printf("  %s movement failed: %v", dir.name, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// flyAscendingSpiral flies in an ascending spiral pattern
func flyAscendingSpiral(drone tello.TelloCommander) {
	fmt.Println("Flying ascending spiral...")
	segments := 12
	distancePerSegment := 30
	heightPerSegment := 10
	rotationPerSegment := 360 / segments

	for i := 0; i < segments; i++ {
		fmt.Printf("  Segment %d/%d\n", i+1, segments)

		if err := drone.Forward(distancePerSegment); err != nil {
			log.Printf("  Forward failed: %v", err)
		}
		time.Sleep(300 * time.Millisecond)

		if err := drone.Clockwise(rotationPerSegment); err != nil {
			log.Printf("  Rotation failed: %v", err)
		}
		time.Sleep(300 * time.Millisecond)

		if i < segments-1 {
			if err := drone.Up(heightPerSegment); err != nil {
				log.Printf("  Up failed: %v", err)
			}
			time.Sleep(300 * time.Millisecond)
		}
	}

	// Descend back to original height
	fmt.Println("  Descending...")
	for i := 0; i < segments-1; i++ {
		if err := drone.Down(heightPerSegment); err != nil {
			log.Printf("  Down failed: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// hoverAndRotate hovers in place while rotating 360 degrees
func hoverAndRotate(drone tello.TelloCommander) {
	fmt.Println("Hovering and rotating 360°...")
	segments := 8
	rotationPerSegment := 360 / segments

	for i := 0; i < segments; i++ {
		fmt.Printf("  Rotation %d/%d\n", i+1, segments)
		if err := drone.Clockwise(rotationPerSegment); err != nil {
			log.Printf("  Rotation failed: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
