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
	// Initialize the Tello commander
	commander, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize Tello commander: %v", err)
	}

	// Enter SDK mode
	fmt.Println("Entering SDK mode...")
	if err := commander.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Wait a moment for initialization
	time.Sleep(2 * time.Second)

	// Start video streaming
	fmt.Println("Starting video stream...")
	if err := commander.StreamOn(); err != nil {
		log.Fatalf("Failed to start video stream: %v", err)
	}

	fmt.Println("Video stream started! Monitoring frames...")
	fmt.Println("Press Ctrl+C to stop")

	// Get access to the video stream listener
	// Note: This requires access to internal components
	// For this example, we'll create a separate video listener
	videoListener, err := transport.NewVideoStreamListener(":11111")
	if err != nil {
		log.Fatalf("Failed to create video listener: %v", err)
	}

	// Start the video listener in a goroutine
	go func() {
		if err := videoListener.Start(); err != nil {
			log.Printf("Video listener error: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Statistics
	frameCount := 0
	keyFrameCount := 0
	totalBytes := 0
	startTime := time.Now()

	// Monitor video frames
	frameChan := videoListener.GetFrameChannel()
	
	for {
		select {
		case frame, ok := <-frameChan:
			if !ok {
				fmt.Println("Video frame channel closed")
				goto cleanup
			}

			frameCount++
			totalBytes += frame.Size

			if frame.IsKeyFrame {
				keyFrameCount++
			}

			// Print frame information every 30 frames
			if frameCount%30 == 0 {
				elapsed := time.Since(startTime)
				fps := float64(frameCount) / elapsed.Seconds()
				bps := float64(totalBytes) / elapsed.Seconds()

				fmt.Printf("Frames: %d | Keyframes: %d | FPS: %.2f | Data rate: %.2f KB/s | Last frame: %d bytes\n",
					frameCount, keyFrameCount, fps, bps/1024, frame.Size)

				// Print NAL unit information for the last frame
				if len(frame.NALUnits) > 0 {
					fmt.Printf("  NAL units: ")
					for _, nalUnit := range frame.NALUnits {
						fmt.Printf("%s ", transport.NewH264Parser().GetNALUTypeName(nalUnit.Type))
					}
					fmt.Println()
				}
			}

		case <-c:
			fmt.Println("\nReceived interrupt signal, stopping...")
			goto cleanup
		}
	}

cleanup:
	// Stop video streaming
	fmt.Println("Stopping video stream...")
	if err := commander.StreamOff(); err != nil {
		log.Printf("Failed to stop video stream: %v", err)
	}

	// Stop video listener
	videoListener.Stop()

	// Print final statistics
	elapsed := time.Since(startTime)
	if frameCount > 0 {
		avgFps := float64(frameCount) / elapsed.Seconds()
		avgFrameSize := float64(totalBytes) / float64(frameCount)
		
		fmt.Printf("\n=== Video Stream Statistics ===\n")
		fmt.Printf("Total frames received: %d\n", frameCount)
		fmt.Printf("Total keyframes: %d\n", keyFrameCount)
		fmt.Printf("Average FPS: %.2f\n", avgFps)
		fmt.Printf("Total data received: %.2f MB\n", float64(totalBytes)/(1024*1024))
		fmt.Printf("Average frame size: %.2f KB\n", avgFrameSize/1024)
		fmt.Printf("Duration: %v\n", elapsed)
	}

	fmt.Println("Video streaming example completed")
}